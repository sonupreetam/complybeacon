package service

import (
	"github.com/ossf/gemara/layer2"
	"github.com/ossf/gemara/layer4"

	"github.com/complytime/complybeacon/compass/api"
	"github.com/complytime/complybeacon/compass/transformer"
)

// Scope defined in scope Layer2 Catalogs by the
// catalog ID
type Scope map[string]layer2.Catalog

// Enrich the raw Evidence with risk attributes based on `gemara` semantics.
func Enrich(rawEnv api.RawEvidence, transformer transformer.Transformer, scope Scope) api.EnrichmentResponse {
	assessmentMethod := transformer.Transform(rawEnv)

	// Determine impacted baselines through Scope
	return api.EnrichmentResponse{
		ImpactedBaselines: determineImpact(scope, assessmentMethod, transformer.Plans()),
		Result:            methodToResult(assessmentMethod),
	}
}

func methodToResult(method layer4.AssessmentMethod) api.Result {
	switch *method.Result {
	case layer4.Passed:
		return api.Passed
	case layer4.Failed:
		return api.Failed
	case layer4.NeedsReview:
		return api.NeedsReview
	default:
		panic("unhandled default case")
	}
}

func determineImpact(scope Scope, inputMethod layer4.AssessmentMethod, plans []transformer.EvaluationPlan) []api.Baseline {
	var baselines []api.Baseline
	for _, plan := range plans {
		catalog, ok := scope[plan.CatalogId]
		if !ok {
			// evaluation is not in scope
			continue
		}

		var impactedRequirements []string
		// Find the Assessment Method in the plan
		for _, eval := range plan.ControlEvaluations {
			for _, requirement := range eval.Assessments {
				for _, method := range requirement.Methods {
					if method.Name == inputMethod.Name {
						impactedRequirements = append(impactedRequirements, requirement.Requirement_Id)
						break
					}
				}
			}
		}

		if len(impactedRequirements) > 0 {
			baseline := api.Baseline{
				Id:           catalog.Metadata.Id,
				Requirements: impactedRequirements,
			}
			baselines = append(baselines, baseline)
		}
	}
	return baselines
}
