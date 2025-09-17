package proofwatch

import (
	"context"
	"encoding/json"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
)

type ProofWatch struct {
	name          string
	provider      log.LoggerProvider
	logger        log.Logger
	observer      *EvidenceObserver
	levelSeverity log.Severity
	opts          []log.LoggerOption
}

func NewProofWatch(name string, meter metric.Meter, options ...log.LoggerOption) (*ProofWatch, error) {
	provider := global.GetLoggerProvider()
	observer, err := NewEvidenceObserver(meter)
	if err != nil {
		return nil, err
	}
	return &ProofWatch{
		name:          name,
		provider:      provider,
		logger:        provider.Logger("proofwatch"),
		observer:      observer,
		levelSeverity: log.SeverityInfo,
		opts:          options,
	}, nil
}

func (w *ProofWatch) Log(ctx context.Context, event Evidence) error {
	attrs, err := w.logEvidence(ctx, event)
	if err != nil {
		return err
	}
	w.observer.Processed(ctx, attrs...)
	return nil
}

// LogEvidence logs the event to the global logger
func (w *ProofWatch) logEvidence(ctx context.Context, event Evidence) ([]attribute.KeyValue, error) {
	record := log.Record{}

	eventId, attrs := ToAttributes(event)
	record.SetEventName(eventId)
	record.SetObservedTimestamp(time.Now())

	var logAttrs []log.KeyValue
	for _, attr := range attrs {
		logAttrs = append(logAttrs, log.KeyValueFromAttribute(attr))
	}
	record.AddAttributes(logAttrs...)

	jsonData, err := json.Marshal(event)
	if err != nil {
		return attrs, err
	}
	evidenceLogData := log.StringValue(string(jsonData))
	record.SetBody(evidenceLogData)

	w.logger.Emit(ctx, record)
	return attrs, nil
}

func ToAttributes(event Evidence) (string, []attribute.KeyValue) {
	var defaultValue = "unknown"
	policySource := defaultValue
	evidenceId := defaultValue
	policyDecision := defaultValue
	policyId := defaultValue

	if event.Metadata.Product.Name != nil {
		policySource = *event.Metadata.Product.Name
	}

	if event.Metadata.Uid != nil {
		evidenceId = *event.Metadata.Uid
	}

	if event.Policy.Uid != nil {
		policyId = *event.Policy.Uid
	}

	if event.Status != nil {
		policyDecision = *event.Status
	}

	return evidenceId, []attribute.KeyValue{
		attribute.Int("category.id", int(event.CategoryUid)),
		attribute.Int("class.id", int(event.ClassUid)),
		attribute.String("policy.source", policySource),
		attribute.String("policy.id", policyId),
		attribute.String("policy.decision", policyDecision),
		attribute.String("evidence.id", evidenceId),
	}
}
