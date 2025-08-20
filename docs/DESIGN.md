# ComplyBeacon Design Documentation

## Key Features

- **OpenTelemetry Native**: Built on the OpenTelemetry standard for seamless integration with existing observability pipelines.
- **Flexible Ingestion**: Provides an optional agent (`ProofWatch`) for non-native evidence sources, allowing users to choose the right ingestion path for their environment.
- **Automated Enrichment**: Enriches raw evidence with risk scores, threat mappings, and regulatory requirements via the Compass service.
- **Resilient Transport**: The `TruthBeam` processor, operating within the OpenTelemetry Collector, provides robust, non-blocking enrichment with built-in resilience and scalability.
- **Composability**: Components are designed as a toolkit; they are not required to be used together, and users can compose their own pipelines.
- **Compliance-as-Code**: Leverages the `gemara` model for a robust, auditable, and automated approach to risk assessment.

## Architecture Overview

### Design Principles

* **Modularity:** The system is composed of small, focused, and interchangeable services.

* **Standardization:** The architecture is built on OpenTelemetry to ensure broad compatibility and interoperability.

* **Resilience:** Components are designed to operate independently, preventing single points of failure and enabling graceful degradation.

* **Operational Experience:** The toolkit is built for easy deployment, configuration, and maintenance using familiar cloud-native practices and protocols.

### Data Flow

The ComplyBeacon architecture is designed to handle two primary data ingestion scenarios, each feeding into a unified enrichment pipeline.

#### The Ingestion Paths
* For Native OpenTelemetry Sources: A policy engine or assessment tool with native OpenTelemetry support sends LogRecords and metrics directly to the `collector`. This path is ideal for streamlined integration.
* For Non-Native Sources: The `proofwatch` agent acts as an adapter. It consumes output from a policy engine, transforms the raw evidence into a standardized `RawEvidence` format, and then converts this into a structured `LogRecord` before sending it to the `collector`. 
  This ensures the required attributes are present while retaining the original data within the log body.

#### The Collector Pipeline
Once a LogRecord is ingested into the `collector` via a configured receiver, it proceeds through the following pipeline:

1. The LogRecord is received and forwarded to the `truthbeam` processor.
2. The `truthbeam` processor extracts key attributes (e.g., `policy.id`) from the log record.
3. It then sends an enrichment request containing this data to the `compass` API.
4. The `compass` service performs a lookup based on the provided attributes and returns a response with compliance-related context (e.g., impacted baselines, requirements, and a compliance result).
5. `truthbeam` adds these new attributes to the original LogRecord.

The now-enriched log record is exported from the `collector` to a final destination (e.g., a SIEM, logging backend, or data lake) for analysis and correlation.
```
┌─────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                                                                                             │
│                                                    ┌─────────────────────────┐                              │
│                                                    │                         │                              │
│                                                    │     Beacon Collector    │                              │
│   ┌────────────────────┐   ┌───────────────────┐   │                         │                              │
│   │                    │   │                   │   ├─────────────────────────┤                              │
│   │                    ├───┤    ProofWatch     ├───┼────┐                    │                              │
│   │     Non-Native     │   │                   │   │    │                    │                              │
│   │                    │   └───────────────────┘   │   ┌┴─────────────────┐  │                              │
│   │    Policy Engine   │                           │   │                  │  │                              │
│   │                    │                           │   │      OTLP        │  │                              │
│   │                    │                           │   │      Reciever    │  │                              │
│   │                    │  ┌────────────────────────┼───┤                  │  │                              │
│   └────────────────────┘  │                        │   └────────┬─────────┘  │                              │
│                           │                        │            │            │               ┌─────────────┐│
│                           │                        │   ┌────────┴─────────┐  │               │             ││
│                           │                        │   │                  │  │               │             ││
│                           │                        │   │    TruthBeam     │──┼──────────────►│ Compass API ││
│   ┌───────────────────────┴───┐                    │   │    Processor     │  │               │             ││
│   │                           │                    │   │                  │  │               │             ││
│   │                           │                    │   └────────┬─────────┘  │               └─────────────┘│
│   │                           │                    │            │            │                              │
│   │    Policy Engine with     │                    │   ┌────────┴─────────┐  │                              │
│   │                           │                    │   │    Exporter      │  │                              │
│   │  Naitve OTEL Integration  │                    │   │   (e.g. Loki     │  │                              │
│   │                           │                    │   │   Splunk)        │  │                              │
│   │                           │                    │   └──────────────────┘  │                              │
│   │                           │                    └─────────────────────────┘                              │
│   └───────────────────────────┘                                                                             │
│                                                                                                             │
└─────────────────────────────────────────────────────────────────────────────────────────────────────────────┘          
```

### Deployment Patterns

ComplyBeacon is designed to be a flexible toolkit, not a framework. Its components can be used in different combinations to fit a variety of operational needs.

1. **Full Pipeline for Non-Native Sources:** This is the most common use case.  
   `Policy Scanner Output -> ProofWatch -> Beacon (w/ TruthBeam) -> Compass -> Final Destination`  
   This provides a complete end-to-end solution for sources that do not natively support OpenTelemetry.

2. **Bypassing `proofwatch` for OTEL-native Sources:**  
   `OTLP-native Source -> Beacon (w/ TruthBeam) -> Compass -> Final Destination`  
   For sources that already emit logs and with the required attributes, the `proofwatch` agent can be omitted, reducing overhead and streamlining the pipeline. The `truthbeam` processor can simply be used to act on the incoming logs directly.

3. **Include TruthBeam in an existing Collector Distro**  
   `OTLP-native Source -> Existing Distro (w/ TruthBeam) -> Compass -> Final Destination`
    If you already have or are already using another distribution, simply add `truthbeam` to your distribution manifest.

4. **Using `compass` as a Standalone Service:**  
   `Existing Pipeline -> Compass -> Existing Pipeline`  
   The `compass` service can be deployed as an independent API, allowing it to be called by any application or a different enrichment processor within an existing OpenTelemetry or custom logging pipeline.

## Component Analysis

### 1. ProofWatch

> This component's architecture is inspired by Promtail. It uses a source and pipeline model to ingest data and then exports it to the OpenTelemetry Collector.

**Purpose**: To serve as the integration layer between policy scanners and the OpenTelemetry pipeline. It standardizes the output of disparate tools into a single, predictable format.

**Key Responsibilities**:

* Parsing the policy scanner's raw output.

* Converting the parsed data into a structured OpenTelemetry log record.

* Emitting the log record to the ComplyBeacon pipeline.

**Key Data Structures**:
`proofwatch` produces an OTEL `LogRecord` with the following attributes and body, based on the policy scanner's output:

```
type RawEvidence struct {
    Metadata `json:,inline`
    Details  json.RawMessage `json:"details"`
    Resource Resource        `json:"resource"`
}

type Metadata struct {
    ID        string    `json:"id"`
    Timestamp time.Time `json:"timestamp"`
    Source    string    `json:"source"`
    PolicyID  string    `json:"policyId"`
    Decision  string    `json:"decision"`
}

// OTLP Log Record produced by proofwatch
// Attributes
log.String("policy.source", rawEnv.Source),
log.String("resource.name", rawEnv.Resource.Name),
log.String("evidence.id", rawEnv.ID),
log.String("policy.decision", rawEnv.Decision),
log.String("policy.id", rawEnv.PolicyID),
log.Slice("resource.hashes", hashes...),

// Body
log.BytesValue(jsonData) // Raw JSON from rawEnv.Details
```

**Possibly Deployment Patterns**

ProofWatch is packaged as a simple CLI to allow it to be deployed in multitude of ways. Some examples could be as a Kubernetes Sidecar, Job, or DaemonSet, in a CI/CD pipeline, or a serverless function.

**Design Patterns**:

* **Adapter Pattern**: `proofwatch` acts as an adapter, translating the output of various policy scanners into a consistent OpenTelemetry format.

### 2. Beacon Collector Distro

**Purpose**: A minimal OpenTelemetry Collector distribution that acts as the runtime environment for the `complybeacon` evidence pipeline, specifically by hosting the `truthbeam` processor.

**Key Responsibilities**:

* Receiving log records from `proofwatch`.

* Running the `truthbeam` log processor on each log record.

* Exporting the processed, enriched logs to a configured backend.

### 3. TruthBeam

**Purpose**: To enrich log records with compliance-related context by querying the `compass` service. This is the core logic that transforms a simple policy check into an actionable compliance event.

**Key Responsibilities**:

* Extracting key attributes from an incoming log record.

* Building an `EnrichmentRequest` payload for the `compass` service.

* Making a synchronous `HTTP POST` request to `compass` at the `/v1/enrich` endpoint.

* Processing the `EnrichmentResponse` and appending new attributes to the original log record.

**Key Data Structures**:
`truthbeam` uses the following data structures for its interaction with `compass`:

```
// EnrichmentRequest is the payload sent to the Compass service.
type EnrichmentRequest struct {
   ClaimId   uuid.UUID       `json:"claimId"`
   Timestamp time.Time       `json:"timestamp"`
   Evidence  RawEvidence     `json:"evidence"`
}

// EnrichmentResponse is the payload received from the Compass service.
type EnrichmentResponse struct {
   Result            string         `json:"result"`
   ImpactedBaselines []ImpactedBase `json:"impactedBaselines"`
}

type ImpactedBase struct {
   Id           string   `json:"id"`
   Requirements []string `json:"requirements"`
}

// Attributes added to the log record by TruthBeam
log.String("compliance.result", enrichRes.Result)
log.Slice("compliance.baselines", baselines...)
log.Slice("compliance.requirements", requirements...)

```

**Design Patterns**:

* **Client-Server Pattern**: `truthbeam` acts as a client to the `compass` server.

* **Decorator Pattern**: `truthbeam` decorates or enriches the original log record with additional information.

### 4. Compass

**Purpose**: To act as a centralized lookup and transformation service that provides compliance and risk context for policy decision events. It's the source of truth for mapping policies to standards and risk attributes.

**Key Responsibilities**:

* Receiving an `EnrichmentRequest` from `truthbeam` or another client.

* Using internal "transformers" to digest specific types of policy outputs.

* Performing a lookup based on the `policy.id` and the policy details.

* Returning an `EnrichmentResponse` with a compliance result, relevant baselines, and requirements.

**Design Patterns**:

* **Service Pattern**: `compass` is a standalone service that provides a specific, well-defined function.

* **Strategy Pattern**: The use of different "transformers" to handle specific policy types is an example of the Strategy Pattern.

## Benefits

Traditional compliance platforms often isolate data in a separate silo, detached from operational insights.
By building `complybeacon` as a modular, OpenTelemetry-based toolkit instead of a monolithic agent-service, we deliberately chose to 
treat compliance as an operational concern. This approach allows for continuous measurement of control effectiveness and seamless correlation with overall system behavior.

* **Resilience:** Each component is a separate service. If the `compass` API becomes unavailable, `truthbeam` can be configured to handle the error gracefully, preventing a full pipeline failure. The logging pipeline remains intact, and unprocessed events can be retried.

* **Scalability:** The modular design allows for independent scaling. If there is a high volume of policy scanner output, multiple instances of `proofwatch` can be deployed without impacting the `Collector` or `compass` services. Similarly, `compass` can scale horizontally to handle increased enrichment requests without requiring changes to the rest of the pipeline.
