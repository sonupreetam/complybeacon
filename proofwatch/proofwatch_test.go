package proofwatch

import (
	"testing"
	"time"

	ocsf "github.com/Santiago-Labs/go-ocsf/ocsf/v1_5_0"
	"go.opentelemetry.io/otel/attribute"
)

// Helper function to create test evidence
func createTestEvidence() Evidence {
	policyName := "test-policy"
	productName := "test-product"
	status := "success"

	return Evidence{
		ScanActivity: ocsf.ScanActivity{
			Time: time.Now().UnixMilli(),
			Metadata: ocsf.Metadata{
				Product: ocsf.Product{
					Name: &productName,
				},
			},
			Status: &status,
		},
		Policy: ocsf.Policy{
			Uid:  &policyName,
			Name: &policyName,
		},
	}
}

func TestMapOCSFToAttributes(t *testing.T) {
	evidence := createTestEvidence()
	attrs := MapOCSFToAttributes(evidence)

	// Check that required attributes are present
	attrMap := make(map[string]interface{})
	for _, attr := range attrs {
		attrMap[string(attr.Key)] = attr.Value.AsInterface()
	}

	// Verify policy attributes
	if attrMap[POLICY_ID] != "test-policy" {
		t.Errorf("Expected POLICY_ID 'test-policy', got %v", attrMap[POLICY_ID])
	}

	if attrMap[POLICY_NAME] != "test-policy" {
		t.Errorf("Expected POLICY_NAME 'test-policy', got %v", attrMap[POLICY_NAME])
	}

	if attrMap[POLICY_SOURCE] != "test-product" {
		t.Errorf("Expected POLICY_SOURCE 'test-product', got %v", attrMap[POLICY_SOURCE])
	}

	// Verify evaluation status mapping
	if attrMap[POLICY_EVALUATION_STATUS] != "pass" {
		t.Errorf("Expected POLICY_EVALUATION_STATUS 'pass', got %v", attrMap[POLICY_EVALUATION_STATUS])
	}
}

func TestMapEvaluationStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   *string
		expected string
	}{
		{
			name:     "success status",
			status:   stringPtr("success"),
			expected: "pass",
		},
		{
			name:     "failure status",
			status:   stringPtr("failure"),
			expected: "fail",
		},
		{
			name:     "unknown status",
			status:   stringPtr("unknown"),
			expected: "unknown",
		},
		{
			name:     "nil status",
			status:   nil,
			expected: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapEvaluationStatus(tt.status)
			if result != tt.expected {
				t.Errorf("mapEvaluationStatus() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMapEnforcementAction(t *testing.T) {
	tests := []struct {
		name          string
		actionID      *int32
		dispositionID *int32
		expected      string
	}{
		{
			name:     "denied action",
			actionID: int32Ptr(2),
			expected: "block",
		},
		{
			name:     "modified action",
			actionID: int32Ptr(4),
			expected: "mutate",
		},
		{
			name:     "observed action",
			actionID: int32Ptr(3),
			expected: "audit",
		},
		{
			name:     "no action",
			actionID: int32Ptr(16),
			expected: "audit",
		},
		{
			name:     "logged action",
			actionID: int32Ptr(17),
			expected: "audit",
		},
		{
			name:     "nil action",
			actionID: nil,
			expected: "audit",
		},
		{
			name:     "unknown action",
			actionID: int32Ptr(99),
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapEnforcementAction(tt.actionID, tt.dispositionID)
			if result != tt.expected {
				t.Errorf("mapEnforcementAction() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMapEnforcementStatus(t *testing.T) {
	tests := []struct {
		name          string
		actionID      *int32
		dispositionID *int32
		expected      string
	}{
		{
			name:     "nil action",
			actionID: nil,
			expected: "success",
		},
		{
			name:          "successful block",
			actionID:      int32Ptr(2),
			dispositionID: int32Ptr(2),
			expected:      "success",
		},
		{
			name:          "successful correction",
			actionID:      int32Ptr(4),
			dispositionID: int32Ptr(11),
			expected:      "success",
		},
		{
			name:          "failed enforcement",
			actionID:      int32Ptr(2),
			dispositionID: int32Ptr(1),
			expected:      "fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapEnforcementStatus(tt.actionID, tt.dispositionID)
			if result != tt.expected {
				t.Errorf("mapEnforcementStatus() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidateEvidenceFields(t *testing.T) {
	tests := []struct {
		name     string
		evidence Evidence
		wantErr  bool
	}{
		{
			name: "valid evidence",
			evidence: func() Evidence {
				policyName := "test-policy"
				productName := "test-product"
				status := "success"
				return Evidence{
					ScanActivity: ocsf.ScanActivity{
						Time: time.Now().UnixMilli(),
						Metadata: ocsf.Metadata{
							Product: ocsf.Product{
								Name: &productName,
							},
						},
						Status: &status,
					},
					Policy: ocsf.Policy{
						Uid:  &policyName,
						Name: &policyName,
					},
				}
			}(),
			wantErr: false,
		},
		{
			name: "missing policy id",
			evidence: func() Evidence {
				productName := "test-product"
				status := "success"
				return Evidence{
					Policy: ocsf.Policy{},
					ScanActivity: ocsf.ScanActivity{
						Time: time.Now().UnixMilli(),
						Metadata: ocsf.Metadata{
							Product: ocsf.Product{
								Name: &productName,
							},
						},
						Status: &status,
					},
				}
			}(),
			wantErr: true,
		},
		{
			name: "missing product name",
			evidence: func() Evidence {
				policyName := "test-policy"
				status := "success"
				return Evidence{
					ScanActivity: ocsf.ScanActivity{
						Time:     time.Now().UnixMilli(),
						Metadata: ocsf.Metadata{},
						Status:   &status,
					},
					Policy: ocsf.Policy{
						Uid:  &policyName,
						Name: &policyName,
					},
				}
			}(),
			wantErr: true,
		},
		{
			name: "missing status",
			evidence: func() Evidence {
				policyName := "test-policy"
				productName := "test-product"
				return Evidence{
					ScanActivity: ocsf.ScanActivity{
						Time: time.Now().UnixMilli(),
						Metadata: ocsf.Metadata{
							Product: ocsf.Product{
								Name: &productName,
							},
						},
					},
					Policy: ocsf.Policy{
						Uid:  &policyName,
						Name: &policyName,
					},
				}
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEvidenceFields(tt.evidence)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateEvidenceFields() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStringVal(t *testing.T) {
	tests := []struct {
		name         string
		s            *string
		defaultValue string
		expected     string
	}{
		{
			name:         "valid string",
			s:            stringPtr("test"),
			defaultValue: "default",
			expected:     "test",
		},
		{
			name:         "nil string",
			s:            nil,
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "empty string",
			s:            stringPtr(""),
			defaultValue: "default",
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringVal(tt.s, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("stringVal() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestVersion(t *testing.T) {
	version := Version()
	if version == "" {
		t.Error("Version() returned empty string")
	}
	if version != "0.1.0" {
		t.Errorf("Version() = %v, want 0.1.0", version)
	}
}

func TestToLogKeyValues(t *testing.T) {
	attrs := []attribute.KeyValue{
		attribute.String("key1", "value1"),
		attribute.Int("key2", 42),
		attribute.Bool("key3", true),
	}

	logAttrs := ToLogKeyValues(attrs)

	if len(logAttrs) != len(attrs) {
		t.Errorf("ToLogKeyValues() returned %d items, want %d", len(logAttrs), len(attrs))
	}

	for i, logAttr := range logAttrs {
		if logAttr.Key != string(attrs[i].Key) {
			t.Errorf("ToLogKeyValues()[%d].Key = %v, want %v", i, logAttr.Key, string(attrs[i].Key))
		}
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}
