package basic

import (
	"strings"

	"github.com/ossf/gemara/layer2"
	"github.com/ossf/gemara/layer4"

	"github.com/complytime/complybeacon/compass/api"
	"github.com/complytime/complybeacon/compass/mapper"
)

// ProcedureInfo represents information about a procedure including its control and requirement IDs
type ProcedureInfo struct {
	ControlID     string
	RequirementID string
	Documentation string
}

// ControlData represents control information including mappings and category
type ControlData struct {
	Mappings []layer2.Mapping
	Category string
}

// A basic mapper processes assessment plans and maps evidence to compliance controls,
// requirements, and standards using the gemara framework.

var (
	_  mapper.Mapper = (*Mapper)(nil)
	ID               = mapper.NewID("basic")
)

type Mapper struct {
	plans map[string][]layer4.AssessmentPlan
}

func (m *Mapper) AddEvaluationPlan(catalogId string, plans ...layer4.AssessmentPlan) {
	existingPlans, ok := m.plans[catalogId]
	if !ok {
		m.plans[catalogId] = plans
	} else {
		existingPlans = append(existingPlans, plans...)
		m.plans[catalogId] = existingPlans
	}
}

func NewBasicMapper() *Mapper {
	return &Mapper{
		plans: make(map[string][]layer4.AssessmentPlan),
	}
}

func (m *Mapper) PluginName() mapper.ID {
	return ID
}

func (m *Mapper) Map(evidence api.RawEvidence, scope mapper.Scope) (api.Compliance, api.Status) {

	// Map decision to status
	status, statusId := m.mapDecision(evidence.Decision)

	// Process each catalog
	for catalogId, plans := range m.plans {
		catalog, ok := scope[catalogId]
		if !ok {
			continue
		}

		// Build procedures map
		proceduresById := m.buildProceduresMap(plans)

		// Build control data map
		controlData := m.buildControlDataMap(catalog)

		// Look up policy in procedures
		if procedureInfo, ok := proceduresById[evidence.PolicyId]; ok {

			// Look up control data
			if ctrlData, ok := controlData[procedureInfo.ControlID]; ok {
				compliance := api.Compliance{
					Catalog:      catalogId,
					Control:      procedureInfo.RequirementID,
					Requirements: m.extractRequirements(ctrlData.Mappings),
					Standards:    m.extractStandards(ctrlData.Mappings),
					Category:     ctrlData.Category,
					Remediation:  &procedureInfo.Documentation,
				}

				return compliance, api.Status{Title: status, Id: &statusId}
			}
		}
	}

	return api.Compliance{}, api.Status{Title: status, Id: &statusId}
}

// mapDecision maps a decision string to status and status ID.
func (m *Mapper) mapDecision(decision string) (api.StatusTitle, api.StatusId) {
	switch strings.ToLower(decision) {
	case "pass", "success":
		return api.Pass, api.N1
	case "fail", "failure":
		return api.Fail, api.N2
	case "other", "warning", "unknown":
		return api.Warning, api.N3
	default:
		return api.Warning, api.N3
	}
}

// buildProceduresMap builds a map of procedure ID to procedure info.
func (m *Mapper) buildProceduresMap(plans []layer4.AssessmentPlan) map[string]ProcedureInfo {
	proceduresById := make(map[string]ProcedureInfo)

	for _, plan := range plans {
		for _, requirement := range plan.Assessments {
			for _, procedure := range requirement.Procedures {
				proceduresById[procedure.Id] = ProcedureInfo{
					ControlID:     plan.Control.EntryId,
					RequirementID: requirement.Requirement.EntryId,
					Documentation: procedure.Documentation,
				}
			}
		}
	}

	return proceduresById
}

// buildControlDataMap builds a map of control ID to control data.
func (m *Mapper) buildControlDataMap(catalog layer2.Catalog) map[string]ControlData {
	controlData := make(map[string]ControlData)

	for _, family := range catalog.ControlFamilies {
		for _, control := range family.Controls {
			controlData[control.Id] = ControlData{
				Mappings: control.GuidelineMappings,
				Category: family.Title,
			}
		}
	}

	return controlData
}

// extractRequirements extracts requirement IDs from mappings.
func (m *Mapper) extractRequirements(mappings []layer2.Mapping) []string {
	var requirements []string
	for _, mapping := range mappings {
		for _, entry := range mapping.Entries {
			requirements = append(requirements, entry.ReferenceId)
		}
	}
	return requirements
}

// extractStandards extracts standard IDs from mappings.
func (m *Mapper) extractStandards(mappings []layer2.Mapping) []string {
	var standards []string
	for _, mapping := range mappings {
		standards = append(standards, mapping.ReferenceId)
	}
	return standards
}
