package transformer

import (
	"github.com/ossf/gemara/layer4"

	"github.com/complytime/complybeacon/compass/api"
)

// Transformer defines a set of methods a plugin must implement for
// transforming RawEvidence into a `gemara` AssessmentMethod.
type Transformer interface {
	PluginName() ID
	Transform(evidence api.RawEvidence) layer4.AssessmentProcedure
	AddEvaluationPlan(plan EvaluationPlan)
	Plans() []EvaluationPlan
}

// EvaluationPlan defines evaluation or assessment strategies from
// from `gemara` Layer 4 for
type EvaluationPlan struct {
	CatalogId          string                      `json:"catalog-id"`
	ControlEvaluations []*layer4.ControlEvaluation `json:"control-evaluations"`
}

// ID represents the identity for a transformer.
type ID string

// NewID returns a new ID for a given id string.
func NewID(id string) ID {
	return ID(id)
}

// Set defines Transformers by ID
type Set map[ID]Transformer
