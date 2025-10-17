package proofwatch

import (
	"context"
	"testing"
	"time"

	ocsf "github.com/Santiago-Labs/go-ocsf/ocsf/v1_5_0"
	"go.opentelemetry.io/otel/attribute"
	olog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/noop"
	mnoop "go.opentelemetry.io/otel/metric/noop"
	tnoop "go.opentelemetry.io/otel/trace/noop"
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

// TestWithMeterProvider tests the WithMeterProvider option.
func TestWithMeterProvider(t *testing.T) {
	t.Run("with valid provider", func(t *testing.T) {
		provider := mnoop.NewMeterProvider()
		pw, err := NewProofWatch(WithMeterProvider(provider))
		if err != nil {
			t.Fatalf("NewProofWatch() with meter provider error = %v", err)
		}
		if pw == nil {
			t.Fatal("NewProofWatch() returned nil")
		}
	})

	t.Run("with nil provider", func(t *testing.T) {
		// Should fall back to default provider
		pw, err := NewProofWatch(WithMeterProvider(nil))
		if err != nil {
			t.Fatalf("NewProofWatch() with nil meter provider error = %v", err)
		}
		if pw == nil {
			t.Fatal("NewProofWatch() returned nil")
		}
	})
}

// TestWithLoggerProvider tests the WithLoggerProvider option.
func TestWithLoggerProvider(t *testing.T) {
	t.Run("with valid provider", func(t *testing.T) {
		provider := noop.NewLoggerProvider()
		pw, err := NewProofWatch(WithLoggerProvider(provider))
		if err != nil {
			t.Fatalf("NewProofWatch() with logger provider error = %v", err)
		}
		if pw == nil {
			t.Fatal("NewProofWatch() returned nil")
		}
	})

	t.Run("with nil provider", func(t *testing.T) {
		// Should fall back to default provider
		pw, err := NewProofWatch(WithLoggerProvider(nil))
		if err != nil {
			t.Fatalf("NewProofWatch() with nil logger provider error = %v", err)
		}
		if pw == nil {
			t.Fatal("NewProofWatch() returned nil")
		}
	})
}

// TestWithTracerProvider tests the WithTracerProvider option.
func TestWithTracerProvider(t *testing.T) {
	t.Run("with valid provider", func(t *testing.T) {
		provider := tnoop.NewTracerProvider()
		pw, err := NewProofWatch(WithTracerProvider(provider))
		if err != nil {
			t.Fatalf("NewProofWatch() with tracer provider error = %v", err)
		}
		if pw == nil {
			t.Fatal("NewProofWatch() returned nil")
		}
	})

	t.Run("with nil provider", func(t *testing.T) {
		// Should fall back to default provider
		pw, err := NewProofWatch(WithTracerProvider(nil))
		if err != nil {
			t.Fatalf("NewProofWatch() with nil tracer provider error = %v", err)
		}
		if pw == nil {
			t.Fatal("NewProofWatch() returned nil")
		}
	})
}

// TestLog tests the Log method.
func TestLog(t *testing.T) {
	pw, err := NewProofWatch(
		WithMeterProvider(mnoop.NewMeterProvider()),
		WithLoggerProvider(noop.NewLoggerProvider()),
		WithTracerProvider(tnoop.NewTracerProvider()),
	)
	if err != nil {
		t.Fatalf("NewProofWatch() error = %v", err)
	}

	ctx := context.Background()
	evidence := createTestEvidence()

	err = pw.Log(ctx, evidence)
	if err != nil {
		t.Errorf("Log() error = %v", err)
	}
}

// TestLogWithSeverity tests the LogWithSeverity method.
func TestLogWithSeverity(t *testing.T) {
	pw, err := NewProofWatch(
		WithMeterProvider(mnoop.NewMeterProvider()),
		WithLoggerProvider(noop.NewLoggerProvider()),
		WithTracerProvider(tnoop.NewTracerProvider()),
	)
	if err != nil {
		t.Fatalf("NewProofWatch() error = %v", err)
	}

	ctx := context.Background()
	evidence := createTestEvidence()

	severities := []olog.Severity{
		olog.SeverityTrace,
		olog.SeverityDebug,
		olog.SeverityInfo,
		olog.SeverityWarn,
		olog.SeverityError,
		olog.SeverityFatal,
	}

	for _, severity := range severities {
		t.Run(severity.String(), func(t *testing.T) {
			err := pw.LogWithSeverity(ctx, evidence, severity)
			if err != nil {
				t.Errorf("LogWithSeverity() error = %v for severity %v", err, severity)
			}
		})
	}
}

// TestLogWithInvalidEvidence tests logging with various evidence states.
func TestLogWithInvalidEvidence(t *testing.T) {
	pw, err := NewProofWatch(
		WithMeterProvider(mnoop.NewMeterProvider()),
		WithLoggerProvider(noop.NewLoggerProvider()),
		WithTracerProvider(tnoop.NewTracerProvider()),
	)
	if err != nil {
		t.Fatalf("NewProofWatch() error = %v", err)
	}

	ctx := context.Background()

	tests := []struct {
		name     string
		evidence Evidence
		wantErr  bool
	}{
		{
			name:     "valid evidence",
			evidence: createTestEvidence(),
			wantErr:  false,
		},
		{
			name: "missing policy id",
			evidence: func() Evidence {
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
					Policy: ocsf.Policy{},
				}
			}(),
			wantErr: false, // Should not error but log warning
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
			wantErr: false, // Should not error but log warning
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
			wantErr: false, // Should not error but log warning
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pw.Log(ctx, tt.evidence)
			if (err != nil) != tt.wantErr {
				t.Errorf("Log() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestLogWithCancelledContext tests logging with a cancelled context.
func TestLogWithCancelledContext(t *testing.T) {
	pw, err := NewProofWatch(
		WithMeterProvider(mnoop.NewMeterProvider()),
		WithLoggerProvider(noop.NewLoggerProvider()),
		WithTracerProvider(tnoop.NewTracerProvider()),
	)
	if err != nil {
		t.Fatalf("NewProofWatch() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	evidence := createTestEvidence()

	// Should still process without error (spans and logs should be fire-and-forget)
	err = pw.Log(ctx, evidence)
	if err != nil {
		t.Errorf("Log() with cancelled context error = %v", err)
	}
}

// TestMapOCSFToAttributesWithEnforcement tests attribute mapping with enforcement actions.
func TestMapOCSFToAttributesWithEnforcement(t *testing.T) {
	tests := []struct {
		name                string
		evidence            Evidence
		expectedAction      string
		expectedEnfStatus   string
		expectedEvalStatus  string
	}{
		{
			name: "blocked with successful disposition",
			evidence: func() Evidence {
				e := createTestEvidence()
				e.ActionID = int32Ptr(2)       // Denied
				e.DispositionID = int32Ptr(2)  // Blocked
				return e
			}(),
			expectedAction:     "block",
			expectedEnfStatus:  "success",
			expectedEvalStatus: "pass",
		},
		{
			name: "mutated with correction",
			evidence: func() Evidence {
				e := createTestEvidence()
				e.ActionID = int32Ptr(4)       // Modified
				e.DispositionID = int32Ptr(11) // Corrected
				return e
			}(),
			expectedAction:     "mutate",
			expectedEnfStatus:  "success",
			expectedEvalStatus: "pass",
		},
		{
			name: "audit with no action",
			evidence: func() Evidence {
				e := createTestEvidence()
				e.ActionID = int32Ptr(16) // No Action
				return e
			}(),
			expectedAction:     "audit",
			expectedEnfStatus:  "fail",
			expectedEvalStatus: "pass",
		},
		{
			name: "no enforcement action",
			evidence: func() Evidence {
				e := createTestEvidence()
				return e
			}(),
			expectedAction:     "audit",
			expectedEnfStatus:  "success",
			expectedEvalStatus: "pass",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attrs := MapOCSFToAttributes(tt.evidence)

			attrMap := make(map[string]interface{})
			for _, attr := range attrs {
				attrMap[string(attr.Key)] = attr.Value.AsInterface()
			}

			if attrMap[POLICY_ENFORCEMENT_ACTION] != tt.expectedAction {
				t.Errorf("POLICY_ENFORCEMENT_ACTION = %v, want %v", 
					attrMap[POLICY_ENFORCEMENT_ACTION], tt.expectedAction)
			}

			if attrMap[POLICY_ENFORCEMENT_STATUS] != tt.expectedEnfStatus {
				t.Errorf("POLICY_ENFORCEMENT_STATUS = %v, want %v", 
					attrMap[POLICY_ENFORCEMENT_STATUS], tt.expectedEnfStatus)
			}

			if attrMap[POLICY_EVALUATION_STATUS] != tt.expectedEvalStatus {
				t.Errorf("POLICY_EVALUATION_STATUS = %v, want %v", 
					attrMap[POLICY_EVALUATION_STATUS], tt.expectedEvalStatus)
			}
		})
	}
}

// TestMapOCSFToAttributesWithDefaultValues tests attribute mapping with missing fields.
func TestMapOCSFToAttributesWithDefaultValues(t *testing.T) {
	evidence := Evidence{
		ScanActivity: ocsf.ScanActivity{
			CategoryUid: 1,
			ClassUid:    2001,
			Time:        time.Now().UnixMilli(),
		},
		Policy: ocsf.Policy{},
	}

	attrs := MapOCSFToAttributes(evidence)

	attrMap := make(map[string]interface{})
	for _, attr := range attrs {
		attrMap[string(attr.Key)] = attr.Value.AsInterface()
	}

	// Check default values are applied
	if attrMap[POLICY_ID] != "unknown_policy_id" {
		t.Errorf("POLICY_ID = %v, want unknown_policy_id", attrMap[POLICY_ID])
	}

	if attrMap[POLICY_NAME] != "unknown_policy_name" {
		t.Errorf("POLICY_NAME = %v, want unknown_policy_name", attrMap[POLICY_NAME])
	}

	if attrMap[POLICY_SOURCE] != "unknown_source" {
		t.Errorf("POLICY_SOURCE = %v, want unknown_source", attrMap[POLICY_SOURCE])
	}

	if attrMap[POLICY_EVALUATION_STATUS] != "error" {
		t.Errorf("POLICY_EVALUATION_STATUS = %v, want error", attrMap[POLICY_EVALUATION_STATUS])
	}
}

// TestMapOCSFToAttributesOCSFFields tests OCSF category and class UIDs.
func TestMapOCSFToAttributesOCSFFields(t *testing.T) {
	evidence := createTestEvidence()
	evidence.CategoryUid = 5
	evidence.ClassUid = 3001

	attrs := MapOCSFToAttributes(evidence)

	attrMap := make(map[string]interface{})
	for _, attr := range attrs {
		attrMap[string(attr.Key)] = attr.Value.AsInterface()
	}

	if attrMap["category_uid"] != int64(5) {
		t.Errorf("category_uid = %v, want 5", attrMap["category_uid"])
	}

	if attrMap["class_uid"] != int64(3001) {
		t.Errorf("class_uid = %v, want 3001", attrMap["class_uid"])
	}
}

// TestMapOCSFToAttributesWithMessage tests status detail mapping.
func TestMapOCSFToAttributesWithMessage(t *testing.T) {
	evidence := createTestEvidence()
	message := "Policy violation detected: resource exceeds quota"
	evidence.Message = &message

	attrs := MapOCSFToAttributes(evidence)

	attrMap := make(map[string]interface{})
	for _, attr := range attrs {
		attrMap[string(attr.Key)] = attr.Value.AsInterface()
	}

	if attrMap[POLICY_STATUS_DETAIL] != message {
		t.Errorf("POLICY_STATUS_DETAIL = %v, want %v", attrMap[POLICY_STATUS_DETAIL], message)
	}
}

// TestToLogKeyValuesEmpty tests conversion of empty attributes.
func TestToLogKeyValuesEmpty(t *testing.T) {
	attrs := []attribute.KeyValue{}
	logAttrs := ToLogKeyValues(attrs)

	if len(logAttrs) != 0 {
		t.Errorf("ToLogKeyValues() returned %d items, want 0", len(logAttrs))
	}
}

// TestToLogKeyValuesTypes tests conversion of different attribute types.
func TestToLogKeyValuesTypes(t *testing.T) {
	attrs := []attribute.KeyValue{
		attribute.String("string_key", "string_value"),
		attribute.Int("int_key", 42),
		attribute.Int64("int64_key", 9223372036854775807),
		attribute.Float64("float64_key", 3.14159),
		attribute.Bool("bool_key", true),
		attribute.StringSlice("string_slice_key", []string{"a", "b", "c"}),
		attribute.IntSlice("int_slice_key", []int{1, 2, 3}),
	}

	logAttrs := ToLogKeyValues(attrs)

	if len(logAttrs) != len(attrs) {
		t.Errorf("ToLogKeyValues() returned %d items, want %d", len(logAttrs), len(attrs))
	}

	// Verify keys match
	for i, logAttr := range logAttrs {
		if logAttr.Key != string(attrs[i].Key) {
			t.Errorf("ToLogKeyValues()[%d].Key = %v, want %v", i, logAttr.Key, string(attrs[i].Key))
		}
	}
}

// BenchmarkLog benchmarks the Log method.
func BenchmarkLog(b *testing.B) {
	pw, err := NewProofWatch(
		WithMeterProvider(mnoop.NewMeterProvider()),
		WithLoggerProvider(noop.NewLoggerProvider()),
		WithTracerProvider(tnoop.NewTracerProvider()),
	)
	if err != nil {
		b.Fatalf("NewProofWatch() error = %v", err)
	}

	ctx := context.Background()
	evidence := createTestEvidence()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pw.Log(ctx, evidence)
	}
}

// BenchmarkLogWithSeverity benchmarks the LogWithSeverity method.
func BenchmarkLogWithSeverity(b *testing.B) {
	pw, err := NewProofWatch(
		WithMeterProvider(mnoop.NewMeterProvider()),
		WithLoggerProvider(noop.NewLoggerProvider()),
		WithTracerProvider(tnoop.NewTracerProvider()),
	)
	if err != nil {
		b.Fatalf("NewProofWatch() error = %v", err)
	}

	ctx := context.Background()
	evidence := createTestEvidence()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pw.LogWithSeverity(ctx, evidence, olog.SeverityInfo)
	}
}

// BenchmarkLogParallel benchmarks parallel logging.
func BenchmarkLogParallel(b *testing.B) {
	pw, err := NewProofWatch(
		WithMeterProvider(mnoop.NewMeterProvider()),
		WithLoggerProvider(noop.NewLoggerProvider()),
		WithTracerProvider(tnoop.NewTracerProvider()),
	)
	if err != nil {
		b.Fatalf("NewProofWatch() error = %v", err)
	}

	evidence := createTestEvidence()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			_ = pw.Log(ctx, evidence)
		}
	})
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}
