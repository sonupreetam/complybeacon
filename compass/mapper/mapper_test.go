package mapper

import (
	"testing"
	"time"

	"github.com/ossf/gemara/layer2"
	"github.com/ossf/gemara/layer4"
	"github.com/stretchr/testify/assert"

	"github.com/complytime/complybeacon/compass/api"
)

// mockMapper is a test implementation of the Mapper interface
type mockMapper struct {
	id    ID
	plans map[string][]layer4.AssessmentPlan
}

func (m *mockMapper) PluginName() ID {
	return m.id
}

func (m *mockMapper) Map(evidence api.RawEvidence, scope Scope) (api.Compliance, api.Status) {
	return api.Compliance{
			Catalog:      "test-catalog",
			Control:      evidence.PolicyId,
			Category:     "test-category",
			Requirements: []string{"req-1"},
			Standards:    []string{"NIST-800-53"},
		}, api.Status{
			Title: api.Pass,
			Id:    &[]api.StatusId{api.N1}[0],
		}
}

func (m *mockMapper) AddEvaluationPlan(catalogId string, plans ...layer4.AssessmentPlan) {
	if m.plans == nil {
		m.plans = make(map[string][]layer4.AssessmentPlan)
	}
	m.plans[catalogId] = plans
}

func TestNewID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ID
	}{
		{
			name:     "basic string",
			input:    "test-mapper",
			expected: ID("test-mapper"),
		},
		{
			name:     "empty string",
			input:    "",
			expected: ID(""),
		},
		{
			name:     "special characters",
			input:    "mapper-v1.0",
			expected: ID("mapper-v1.0"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewID(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSet tests map operations (add, retrieve, delete) on the Set type
func TestSet(t *testing.T) {
	set := make(Set)
	assert.NotNil(t, set)
	assert.Empty(t, set)

	policyMapper := &mockMapper{id: "policy-mapper"}
	controlMapper := &mockMapper{id: "control-mapper"}

	// Add mappers
	set[ID("policy-mapper")] = policyMapper
	set[ID("control-mapper")] = controlMapper

	assert.Len(t, set, 2)
	assert.Contains(t, set, ID("policy-mapper"))
	assert.Contains(t, set, ID("control-mapper"))
	assert.Equal(t, policyMapper, set[ID("policy-mapper")])
	assert.Equal(t, controlMapper, set[ID("control-mapper")])

	// Retrieve existing mapper
	retrieved, exists := set[ID("policy-mapper")]
	assert.True(t, exists)
	assert.Equal(t, policyMapper, retrieved)

	// Retrieve non-existent mapper
	_, exists = set[ID("non-existent")]
	assert.False(t, exists)

	// Delete mapper
	delete(set, ID("policy-mapper"))
	assert.Len(t, set, 1)
	assert.NotContains(t, set, ID("policy-mapper"))
	assert.Contains(t, set, ID("control-mapper"))
}

func TestScope(t *testing.T) {
	scope := make(Scope)
	assert.NotNil(t, scope)
	assert.Empty(t, scope)

	catalog1 := layer2.Catalog{Metadata: layer2.Metadata{Id: "catalog-1"}}
	catalog2 := layer2.Catalog{Metadata: layer2.Metadata{Id: "catalog-2"}}

	// Add catalogs
	scope["catalog-1"] = catalog1
	scope["catalog-2"] = catalog2

	assert.Len(t, scope, 2)
	assert.Contains(t, scope, "catalog-1")
	assert.Contains(t, scope, "catalog-2")
	assert.Equal(t, catalog1, scope["catalog-1"])
	assert.Equal(t, catalog2, scope["catalog-2"])

	// Retrieve existing catalog
	retrieved, exists := scope["catalog-1"]
	assert.True(t, exists)
	assert.Equal(t, catalog1, retrieved)

	// Retrieve non-existent catalog
	_, exists = scope["non-existent"]
	assert.False(t, exists)
}

// TestMapperInterfaceAndIDType tests ID type operations and mock mapper interface implementation
func TestMapperInterfaceAndIDType(t *testing.T) {
	t.Run("ID type operations", func(t *testing.T) {
		id1 := ID("test-id")
		id2 := ID("test-id")
		id3 := ID("different-id")

		assert.Equal(t, id1, id2)
		assert.NotEqual(t, id1, id3)

		assert.Equal(t, "test-id", string(id1))
		assert.Equal(t, "different-id", string(id3))
	})

	t.Run("mock mapper implements interface", func(t *testing.T) {
		mapper := &mockMapper{id: "test-mapper"}

		var _ Mapper = mapper

		assert.Equal(t, ID("test-mapper"), mapper.PluginName())

		evidence := api.RawEvidence{
			Id:        "test-raw-evidence",
			PolicyId:  "AC-1",
			Source:    "test-policy-engine",
			Decision:  "pass",
			Timestamp: time.Now(),
		}
		scope := make(Scope)

		compliance, status := mapper.Map(evidence, scope)
		assert.Equal(t, "test-catalog", compliance.Catalog)
		assert.Equal(t, "AC-1", compliance.Control)
		assert.Equal(t, api.Pass, status.Title)

		plans := []layer4.AssessmentPlan{
			{Control: layer4.Mapping{ReferenceId: "AC-1"}},
		}
		mapper.AddEvaluationPlan("test-catalog", plans...)
		assert.Len(t, mapper.plans, 1)
		assert.Contains(t, mapper.plans, "test-catalog")
	})
}
