package proofwatch

import (
	"testing"

	"go.opentelemetry.io/otel/attribute"
)

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
