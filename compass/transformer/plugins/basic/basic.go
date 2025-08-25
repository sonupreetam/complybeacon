package basic

import (
	"fmt"

	"github.com/ossf/gemara/layer4"

	"github.com/complytime/complybeacon/compass/api"

	"github.com/complytime/complybeacon/compass/transformer"
)

// A basic transformer provide context in a shallow manner by parsing the metadata.

var (
	_  transformer.Transformer = (*Transformer)(nil)
	ID                         = transformer.NewID("basic")
)

type Transformer struct {
	plans []transformer.EvaluationPlan
}

func (m *Transformer) AddEvaluationPlan(plan transformer.EvaluationPlan) {
	m.plans = append(m.plans, plan)
}

func NewBasicTransformer() *Transformer {
	return &Transformer{}
}

func (m *Transformer) PluginName() transformer.ID {
	return ID
}

func (m *Transformer) Transform(evidence api.RawEvidence) layer4.AssessmentProcedure {
	// Make a reasonable attempt to determine result here
	var result layer4.Result
	switch evidence.Decision {
	case "pass", "passed", "success", "compliant":
		result = layer4.Passed
	case "fail", "failed", "failure", "not-compliant":
		result = layer4.Failed
	case "error", "warning":
		result = layer4.NeedsReview
	default:
		// Always default to needs review if the result can't
		// be determined
		result = layer4.NeedsReview

	}
	return layer4.AssessmentProcedure{
		Id:          evidence.PolicyId,
		Name:        evidence.PolicyId,
		Description: fmt.Sprintf("%v", evidence.Details),
		Run:         true,
		Result:      result,
	}
}

func (m *Transformer) Plans() []transformer.EvaluationPlan {
	return m.plans
}
