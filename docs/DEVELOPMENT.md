# ComplyBeacon Development Guide

This guide provides comprehensive instructions for setting up, building, and testing the ComplyBeacon project.
It complements the [DESIGN.md](./DESIGN.md) document by focusing on the practical aspects of development.

<!-- TOC -->
* [ComplyBeacon Development Guide](#complybeacon-development-guide)
  * [Prerequisites](#prerequisites)
    * [Required Software](#required-software)
  * [Development Environment Setup](#development-environment-setup)
    * [1. Clone the Repository](#1-clone-the-repository)
    * [2. Install Task (if needed)](#2-install-task-if-needed)
    * [3. Initialize Go Workspace](#3-initialize-go-workspace)
    * [4. Install Dependencies](#4-install-dependencies)
    * [5. Verify Installation](#5-verify-installation)
  * [Project Structure](#project-structure)
  * [Testing](#testing)
    * [Running Tests](#running-tests)
    * [Integration Testing](#integration-testing)
  * [Component Development](#component-development)
    * [1. ProofWatch Development](#1-proofwatch-development)
    * [2. Compass Development](#2-compass-development)
    * [3. TruthBeam Development](#3-truthbeam-development)
    * [4. Beacon Distro Development](#4-beacon-distro-development)
  * [Debugging and Troubleshooting](#debugging-and-troubleshooting)
    * [Debugging Tools](#debugging-tools)
  * [Code Generation](#code-generation)
    * [1. OpenTelemetry Semantic Conventions](#1-opentelemetry-semantic-conventions)
    * [2. Manual Code Generation](#2-manual-code-generation)
  * [Deployment and Demo](#deployment-and-demo)
    * [Local Development Demo](#local-development-demo)
  * [Additional Resources](#additional-resources)
<!-- TOC -->

## Prerequisites

### Required Software

- **Go 1.26+**: The project uses Go 1.26.3
- **Podman**: For containerized development and deployment (Docker is not supported)
- **Task**: For build automation ([installation guide](https://taskfile.dev/installation/))
- **Git**: For version control
- **openssl**: Cryptography toolkit

## Development Environment Setup

### 1. Clone the Repository

```bash
git clone https://github.com/complytime/complybeacon.git
cd complybeacon
```

### 2. Install Task (if needed)

The project uses [Task](https://taskfile.dev) for build automation. Install it if you don't have it:

```bash
# macOS
brew install go-task/tap/go-task

# Linux
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b ~/.local/bin

# Or using Go
go install github.com/go-task/task/v3/cmd/task@latest

# Verify installation
task --version
```

### 3. Initialize Go Workspace

The project uses Go workspaces to manage multiple modules:

```bash
task workspace
```

This creates a `go.work` file that includes all project modules:
- `./proofwatch`
- `./truthbeam`

### 4. Install Dependencies

Dependencies are managed per module. Install them for all modules:

```bash
task deps
```

This automatically:
- Syncs the Go workspace
- Runs `go mod tidy` for each module
- Verifies and downloads dependencies

### 5. Verify Installation

```bash
# Run tests to verify everything works
task test

# Run all quality gates (lint + test)
task check
```

## Project Structure

```
complybeacon/
├── compose.yaml                # Container orchestration configuration
├── Taskfile.yml                # Build automation
├── .taskfiles/                 # Task modules and helper scripts
├── docs/                       # Documentation
│   ├── DESIGN.md              # Architecture and design documentation
│   ├── DEVELOPMENT.md         # This file
│   └── attributes/            # Attribute documentation
├── model/                      # OpenTelemetry semantic conventions
│   ├── attributes.yaml        # Attribute definitions
│   └── entities.yaml          # Entity definitions
├── proofwatch/                 # ProofWatch instrumentation library
│   ├── attributes.go          # Attribute definitions
│   ├── evidence.go            # Evidence types
│   └── proofwatch.go          # Main library
├── truthbeam/                  # TruthBeam processor module
│   ├── internal/              # Internal packages
│   ├── config.go              # Configuration
│   └── processor.go           # Main processor logic
├── beacon-distro/              # OpenTelemetry Collector distribution
│   ├── config.yaml            # Collector configuration
│   └── Containerfile.collector # Container definition
├── configs/                    # Deployment configs (collector, Loki)
│   ├── collector-base.yaml    # Base layer: OCSF transform + Loki
│   ├── collector-storage.yaml # Storage layer: adds S3 export
│   ├── collector-enrichment.yaml # Enrichment layer: adds TruthBeam
│   └── loki.yaml              # Loki configuration
├── certs/                      # TLS certificate generation
├── deploy/                     # Deployment infrastructure (Terraform)
├── tests/                      # Test infrastructure
│   └── integration/           # E2E Ginkgo tests, mock Compass, fixtures
└── bin/                        # Built binaries (created by task infra:deploy)
```

## Testing

### Running Tests

```bash
# Run all tests (includes version checks and coverage)
task test

# Run tests with race detection
task test-race

# Generate coverage reports
task dev:coverage-report

# Run tests for specific module
cd proofwatch && go test -v ./...
cd truthbeam && go test -v ./...
```

### Integration Testing

The project includes automated integration tests using [Ginkgo](https://onsi.github.io/ginkgo/) that validate the evidence pipeline at three deployment layers:

| Layer | Profile | What it tests |
|-------|---------|--------------|
| Base | *(default)* | OCSF transform + Loki export |
| Storage | `storage` | S3 evidence export + partitioning |
| Enrichment | `enrichment` | TruthBeam enrichment via mock Compass |

**Prerequisites:**
- Podman and podman-compose
- Go 1.26+ (Ginkgo CLI is managed via `tool` directive in root `go.mod`)

**Run all layers:**

```bash
task integration:test
```

**Run a single layer:**

```bash
task integration:test-profile PROFILE=base
task integration:test-profile PROFILE=storage
task integration:test-profile PROFILE=enrichment
```

Each run builds the collector image, starts the appropriate services, runs the matching Ginkgo test suite (filtered by label), and tears down. Certificates are generated automatically if missing. Test output is written to `.test-output/integration/`.

For details on test cases, fixtures, and mock Compass configuration, see [tests/integration/README.md](../tests/integration/README.md).

## Component Development

### 1. ProofWatch Development

ProofWatch is an instrumentation library for emitting compliance evidence.

**Key Files:**
- `proofwatch/proofwatch.go` - Main library interface
- `proofwatch/evidence.go` - Evidence type definition
- `proofwatch/attributes.go` - OpenTelemetry attributes

**Development Workflow:**
```bash
cd proofwatch

# Run tests
go test -v ./...

# Run linting (from root)
cd ..
task lint

# Format code
go fmt ./...
```

### 2. Compass Development

Compass is an external enrichment service that TruthBeam connects to for compliance lookups. It must be provided separately and is not included in the demo stack.

### 3. TruthBeam Development

TruthBeam is an OpenTelemetry Collector processor for enriching logs.

**Key Files:**
- `truthbeam/processor.go` - Main processor logic
- `truthbeam/config.go` - Configuration structures
- `truthbeam/factory.go` - Processor factory

**Development Workflow:**
```bash
cd truthbeam

# Run tests
go test -v ./...

# Test with collector (requires beacon-distro)
cd ../beacon-distro
# Modify config to use local truthbeam
# Run collector with local processor
```

**Local development config**

If you want locally test the TruthBeam, remember to change the [manifest.yaml](../beacon-distro/manifest.yaml)

Add replace directive at the end of [manifest.yaml](../beacon-distro/manifest.yaml), to make sure collector use your `truthbeam` code. Default collector will use `- gomod: github.com/complytime/complybeacon/truthbeam main`

For example:
```yaml
replaces:
  - github.com/complytime/complybeacon/truthbeam => github.com/AlexXuan233/complybeacon/truthbeam 52e4a76ea0f72a7049e73e7a5d67d988116a3892
```
or
```yaml
replaces:
  - github.com/complytime/complybeacon/truthbeam => github.com/AlexXuan233/complybeacon/truthbeam main
```

### 4. Beacon Distro Development

The Beacon distribution is a custom OpenTelemetry Collector.

**Key Files:**
- `beacon-distro/config.yaml` - Collector configuration
- `beacon-distro/Containerfile.collector` - Container definition

**Development Workflow:**
```bash
# Build the collector image
podman build -f beacon-distro/Containerfile.collector -t complybeacon-collector beacon-distro/

# Or force rebuild without cache
podman build --no-cache -f beacon-distro/Containerfile.collector -t complybeacon-collector beacon-distro/

# Run locally for quick testing
podman run --rm complybeacon-collector --config /etc/otelcol-beacon/config.yaml

# Full stack deployment for integration testing
task infra:deploy
```

## Debugging and Troubleshooting

### Debugging Tools

```bash
# View all container logs
podman-compose -f compose.yaml logs -f

# View specific service logs
podman-compose -f compose.yaml ps            # List running services
podman-compose -f compose.yaml logs -f collector

# Check container status
podman images | grep complybeacon            # List built images
podman inspect complybeacon-collector        # Inspect image details
```

## Code Generation

The project uses several code generation tools:

### 1. OpenTelemetry Semantic Conventions

Generate documentation and Go code from semantic convention models:

```bash
# Generate documentation
task codegen:weaver-docsgen

# Generate Go code
task codegen:weaver-codegen

# Validate models
task codegen:weaver-check

# Validate logs against semantic conventions
task codegen:weaver-semantic-check
```

### 2. Manual Code Generation

If you modify the semantic conventions:

```bash
# Update semantic conventions
vim model/attributes.yaml
vim model/entities.yaml

# Regenerate all code (API + weaver)
task codegen:api-codegen
task codegen:weaver-codegen
```

## Deployment and Demo

### Local Development Demo

The demo environment orchestrates multiple containers (Grafana, Loki, Beacon Collector, Compass).

1. **Generate self-signed certificate**

Since compass and truthbeam enable TLS by default, first generate self-signed certificates for testing/development:

```bash
task infra:generate-self-signed-cert
```

2. **Start the full stack:**
```bash
# Interactive mode (shows logs in terminal)
task deploy

# Or background/detached mode
podman-compose -f compose.yaml up -d
```

This automatically:
- Syncs OTel versions from truthbeam to beacon-distro
- Builds the beacon collector image
- Starts all services (Grafana, Loki, Collector)

3. **Test the pipeline:**
```bash
curl -X POST http://localhost:8088/eventsource/receiver \
  -H "Content-Type: application/json" \
  -d @tests/integration/fixtures/evidence-fail.json
```

4. **View results:**
- Grafana: <http://localhost:3000>
- View logs: `podman-compose -f compose.yaml logs -f`

5. **Stop the stack:**
```bash
task infra:undeploy
```

---

## Additional Resources

- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [Go Documentation](https://golang.org/doc/)
- [Podman Documentation](https://docs.podman.io/)
- [Project Design Document](./DESIGN.md)
- [Attribute Documentation](./attributes/)
- [Containers Guide](https://github.com/complytime/community/blob/main/CONTAINERS_GUIDE.md)

For questions or support, please open an issue in the GitHub repository.
