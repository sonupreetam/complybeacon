package evidence

import (
	"context"
	"encoding/json"
	"time"

	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
)

type InstrumentationFn func(ctx context.Context, evidence RawEvidence) error

func NewEmitter(observer *EvidenceObserver) InstrumentationFn {
	return func(ctx context.Context, evidence RawEvidence) error {
		evidenceEvent, err := LogEvidence(ctx, evidence)
		if err != nil {
			return err
		}
		attrs := ToAttributes(evidenceEvent)
		observer.Processed(ctx, attrs...)
		return nil
	}
}

// LogEvidence logs the event to the global logger
func LogEvidence(ctx context.Context, rawEnv RawEvidence) (EvidenceEvent, error) {
	logger := global.Logger("proofwatch")
	event := NewFromEvidence(rawEnv)
	record := log.Record{}
	record.SetEventName(event.Summary)
	record.SetTimestamp(event.Timestamp)
	record.SetObservedTimestamp(time.Now())

	// Adding metadata as attributes and full log details as the body
	record.AddAttributes(
		log.String("policy.source", rawEnv.Source),
		log.String("subject.name", rawEnv.Subject.Name),
		log.String("evidence.id", rawEnv.ID),
		log.String("policy.decision", rawEnv.Decision),
		log.String("policy.id", rawEnv.PolicyID),
		log.String("subject,uri", rawEnv.Subject.URI),
	)

	jsonData, err := json.Marshal(rawEnv.Details)
	if err != nil {
		return event, err
	}
	claimValue := log.BytesValue(jsonData)
	record.SetBody(claimValue)

	logger.Emit(ctx, record)
	return event, nil
}
