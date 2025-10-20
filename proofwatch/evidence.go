package proofwatch

import (
	"time"

	"go.opentelemetry.io/otel/attribute"
)

// Evidence defines the interface for compliance evidence data that can be collected
// and processed by the proofwatch.
type Evidence interface {
	// Marshaler serializes the evidence data to JSON format.
	ToJSON() ([]byte, error)

	// Attributes converts the evidence data into OpenTelemetry attribute key-value pairs
	// for observability, monitoring, and compliance tracking purposes
	Attributes() []attribute.KeyValue

	// Timestamp returns the time when the evidence was generated or collected
	Timestamp() time.Time
}
