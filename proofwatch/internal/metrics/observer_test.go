package metrics

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/noop"
)

// TestNewEvidenceObserver tests the creation of a new EvidenceObserver.
func TestNewEvidenceObserver(t *testing.T) {
	meter := noop.NewMeterProvider().Meter("test")

	observer, err := NewEvidenceObserver(meter)
	if err != nil {
		t.Fatalf("Failed to create EvidenceObserver: %v", err)
	}

	if observer == nil {
		t.Fatal("EvidenceObserver is nil")
	}

	if observer.meter == nil {
		t.Error("Meter is nil")
	}

	if observer.droppedCounter == nil {
		t.Error("droppedCounter is nil")
	}

	if observer.processedCount == nil {
		t.Error("processedCount is nil")
	}
}

// TestEvidenceObserverProcessed tests the Processed method.
func TestEvidenceObserverProcessed(t *testing.T) {
	meter := noop.NewMeterProvider().Meter("test")
	observer, err := NewEvidenceObserver(meter)
	if err != nil {
		t.Fatalf("Failed to create EvidenceObserver: %v", err)
	}

	ctx := context.Background()
	attrs := []attribute.KeyValue{
		attribute.String("policy.id", "test-policy"),
		attribute.String("policy.source", "test-scanner"),
	}

	// This should not panic
	observer.Processed(ctx, attrs...)
}

// TestEvidenceObserverDropped tests the Dropped method.
func TestEvidenceObserverDropped(t *testing.T) {
	meter := noop.NewMeterProvider().Meter("test")
	observer, err := NewEvidenceObserver(meter)
	if err != nil {
		t.Fatalf("Failed to create EvidenceObserver: %v", err)
	}

	ctx := context.Background()
	attrs := []attribute.KeyValue{
		attribute.String("reason", "invalid_format"),
		attribute.String("policy.id", "test-policy"),
	}

	// This should not panic
	observer.Dropped(ctx, attrs...)
}

// TestEvidenceObserverWithMultipleAttributes tests observer with various attributes.
func TestEvidenceObserverWithMultipleAttributes(t *testing.T) {
	meter := noop.NewMeterProvider().Meter("test")
	observer, err := NewEvidenceObserver(meter)
	if err != nil {
		t.Fatalf("Failed to create EvidenceObserver: %v", err)
	}

	ctx := context.Background()

	tests := []struct {
		name  string
		attrs []attribute.KeyValue
	}{
		{
			name: "single attribute",
			attrs: []attribute.KeyValue{
				attribute.String("policy.id", "policy-001"),
			},
		},
		{
			name: "multiple attributes",
			attrs: []attribute.KeyValue{
				attribute.String("policy.id", "policy-001"),
				attribute.String("policy.source", "scanner-v1"),
				attribute.String("evidence.id", "evidence-123"),
				attribute.String("policy.status", "pass"),
			},
		},
		{
			name: "no attributes",
			attrs: []attribute.KeyValue{},
		},
		{
			name: "attributes with different types",
			attrs: []attribute.KeyValue{
				attribute.String("policy.id", "policy-001"),
				attribute.Int("category.id", 1001),
				attribute.Bool("is_compliant", true),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			observer.Processed(ctx, tt.attrs...)
			observer.Dropped(ctx, tt.attrs...)
		})
	}
}

// TestEvidenceObserverMultipleCalls tests multiple sequential calls.
func TestEvidenceObserverMultipleCalls(t *testing.T) {
	meter := noop.NewMeterProvider().Meter("test")
	observer, err := NewEvidenceObserver(meter)
	if err != nil {
		t.Fatalf("Failed to create EvidenceObserver: %v", err)
	}

	ctx := context.Background()
	attrs := []attribute.KeyValue{
		attribute.String("policy.id", "policy-001"),
	}

	// Test multiple calls
	for i := 0; i < 100; i++ {
		observer.Processed(ctx, attrs...)
		if i%10 == 0 {
			observer.Dropped(ctx, attrs...)
		}
	}
}

// TestEvidenceObserverWithCancelledContext tests observer with cancelled context.
func TestEvidenceObserverWithCancelledContext(t *testing.T) {
	meter := noop.NewMeterProvider().Meter("test")
	observer, err := NewEvidenceObserver(meter)
	if err != nil {
		t.Fatalf("Failed to create EvidenceObserver: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	attrs := []attribute.KeyValue{
		attribute.String("policy.id", "policy-001"),
	}

	// Should still work with cancelled context (metrics should be fire-and-forget)
	observer.Processed(ctx, attrs...)
	observer.Dropped(ctx, attrs...)
}

// TestEvidenceObserverNilAttributes tests observer with nil attributes.
func TestEvidenceObserverNilAttributes(t *testing.T) {
	meter := noop.NewMeterProvider().Meter("test")
	observer, err := NewEvidenceObserver(meter)
	if err != nil {
		t.Fatalf("Failed to create EvidenceObserver: %v", err)
	}

	ctx := context.Background()

	// Should not panic with no attributes
	observer.Processed(ctx)
	observer.Dropped(ctx)
}

// BenchmarkEvidenceObserverProcessed benchmarks the Processed method.
func BenchmarkEvidenceObserverProcessed(b *testing.B) {
	meter := noop.NewMeterProvider().Meter("test")
	observer, err := NewEvidenceObserver(meter)
	if err != nil {
		b.Fatalf("Failed to create EvidenceObserver: %v", err)
	}

	ctx := context.Background()
	attrs := []attribute.KeyValue{
		attribute.String("policy.id", "policy-001"),
		attribute.String("policy.source", "scanner-v1"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		observer.Processed(ctx, attrs...)
	}
}

// BenchmarkEvidenceObserverDropped benchmarks the Dropped method.
func BenchmarkEvidenceObserverDropped(b *testing.B) {
	meter := noop.NewMeterProvider().Meter("test")
	observer, err := NewEvidenceObserver(meter)
	if err != nil {
		b.Fatalf("Failed to create EvidenceObserver: %v", err)
	}

	ctx := context.Background()
	attrs := []attribute.KeyValue{
		attribute.String("reason", "invalid"),
		attribute.String("policy.id", "policy-001"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		observer.Dropped(ctx, attrs...)
	}
}

// BenchmarkEvidenceObserverParallel benchmarks parallel calls to observer.
func BenchmarkEvidenceObserverParallel(b *testing.B) {
	meter := noop.NewMeterProvider().Meter("test")
	observer, err := NewEvidenceObserver(meter)
	if err != nil {
		b.Fatalf("Failed to create EvidenceObserver: %v", err)
	}

	attrs := []attribute.KeyValue{
		attribute.String("policy.id", "policy-001"),
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			observer.Processed(ctx, attrs...)
		}
	})
}

