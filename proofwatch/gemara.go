package proofwatch

import (
	"encoding/json"
	"time"

	"github.com/ossf/gemara/layer4"
	"go.opentelemetry.io/otel/attribute"
)

var _ Evidence = (*GemaraEvidence)(nil)

// GemaraEvidence represents evidence data from the Gemara compliance assessment framework.
// It embeds both layer4.Metadata and layer4.AssessmentLog to provide comprehensive
// compliance assessment information that can be used for evidence collection and reporting.
type GemaraEvidence struct {
	layer4.Metadata
	layer4.AssessmentLog
}

func (g GemaraEvidence) ToJSON() ([]byte, error) {
	return json.Marshal(g)
}

func (g GemaraEvidence) Attributes() []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.String(POLICY_SOURCE, g.Author.Name),
		attribute.String(COMPLIANCE_CONTROL_ID, g.Requirement.EntryId),
		attribute.String(COMPLIANCE_CONTROL_CATALOG_ID, g.Requirement.ReferenceId),
		attribute.String(POLICY_EVALUATION_STATUS, g.Result.String()),
		// For Layer 4, we assumed the enforcement action in audit and action is taken in Layer 5.
		attribute.String(POLICY_ENFORCEMENT_ACTION, "audit"),
		attribute.String(POLICY_ID, g.Procedure.EntryId),
	}

	if g.Message != "" {
		attrs = append(attrs, attribute.String(POLICY_STATUS_DETAIL, g.Message))
	}

	if g.Recommendation != "" {
		attrs = append(attrs, attribute.String(COMPLIANCE_CONTROL_REMEDIATION_DESCRIPTION, g.Recommendation))
	}

	return attrs
}

func (g GemaraEvidence) Timestamp() time.Time {
	timestamp, err := time.Parse(time.RFC3339, string(g.End))
	if err != nil {
		return time.Now()
	}
	return timestamp
}
