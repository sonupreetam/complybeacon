package basic

import (
	"testing"
	"time"

	"github.com/ossf/gemara/layer4"
	"github.com/stretchr/testify/assert"

	"github.com/complytime/complybeacon/compass/api"
	"github.com/complytime/complybeacon/compass/mapper"
)

func TestNewBasicMapper(t *testing.T) {
	basicMapper := NewBasicMapper()

	assert.NotNil(t, basicMapper)
	assert.Equal(t, ID, basicMapper.PluginName())
	assert.NotNil(t, basicMapper.plans)
	assert.Empty(t, basicMapper.plans)
}

func TestBasicMapper_Map(t *testing.T) {
	tests := []struct {
		name          string
		decision      string
		expectedTitle api.StatusTitle
		expectedId    api.StatusId
	}{
		{
			name:          "compliance status is pass",
			decision:      "pass",
			expectedTitle: api.Pass,
			expectedId:    api.N1,
		},
		{
			name:          "compliance status is fail",
			decision:      "fail",
			expectedTitle: api.Fail,
			expectedId:    api.N2,
		},
		{
			name:          "compliance status is warning",
			decision:      "Warning",
			expectedTitle: api.Warning,
			expectedId:    api.N3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			basicMapper := NewBasicMapper()
			evidence := api.RawEvidence{
				Id:        "test-raw-evidence",
				Source:    "test-policy-engine",
				PolicyId:  "AC-1",
				Decision:  tt.decision,
				Timestamp: time.Now(),
			}
			scope := make(mapper.Scope)

			_, status := basicMapper.Map(evidence, scope)

			assert.Equal(t, tt.expectedTitle, status.Title)
			assert.Equal(t, tt.expectedId, *status.Id)
		})
	}
}

func TestBasicMapper_AddEvaluationPlan(t *testing.T) {
	t.Run("adds evaluation plan", func(t *testing.T) {
		basicMapper := NewBasicMapper()
		plans := []layer4.AssessmentPlan{
			{Control: layer4.Mapping{ReferenceId: "AC-1"}},
		}

		basicMapper.AddEvaluationPlan("test-catalog", plans...)

		assert.Len(t, basicMapper.plans, 1)
		assert.Contains(t, basicMapper.plans, "test-catalog")
		assert.Len(t, basicMapper.plans["test-catalog"], 1)
		assert.Equal(t, "AC-1", basicMapper.plans["test-catalog"][0].Control.ReferenceId)
	})

	t.Run("appends to existing evaluation plans", func(t *testing.T) {
		basicMapper := NewBasicMapper()
		initialPlans := []layer4.AssessmentPlan{
			{Control: layer4.Mapping{ReferenceId: "AC-1"}},
		}
		additionalPlans := []layer4.AssessmentPlan{
			{Control: layer4.Mapping{ReferenceId: "AC-2"}},
		}

		basicMapper.AddEvaluationPlan("test-catalog", initialPlans...)
		basicMapper.AddEvaluationPlan("test-catalog", additionalPlans...)

		assert.Len(t, basicMapper.plans, 1)
		assert.Len(t, basicMapper.plans["test-catalog"], 2)
		assert.Equal(t, "AC-1", basicMapper.plans["test-catalog"][0].Control.ReferenceId)
		assert.Equal(t, "AC-2", basicMapper.plans["test-catalog"][1].Control.ReferenceId)
	})
}
