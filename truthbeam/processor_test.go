package truthbeam

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/processor/processortest"
	"go.uber.org/zap/zaptest"

	"github.com/complytime/complybeacon/truthbeam/internal/client"
)

func TestNewTruthBeamProcessor(t *testing.T) {
	cfg := &Config{
		ClientConfig: confighttp.NewDefaultClientConfig(),
	}
	cfg.ClientConfig.Endpoint = "http://localhost:8081"

	settings := processortest.NewNopSettings(component.MustNewType("test"))
	settings.Logger = zaptest.NewLogger(t)

	processor, err := newTruthBeamProcessor(cfg, settings)
	require.NoError(t, err, "Error creating truth beam processor")
	require.NotNil(t, processor, "Processor should not be nil")
	assert.Equal(t, cfg, processor.config)
	assert.NotNil(t, processor.client)
	assert.NotNil(t, processor.logger)
}

func TestNewTruthBeamProcessorWithInvalidConfig(t *testing.T) {
	processor, err := newTruthBeamProcessor(nil, processortest.NewNopSettings(component.MustNewType("test")))
	assert.Error(t, err, "Expected error with nil config")
	assert.Nil(t, processor, "Processor should be nil with invalid config")

	wrongConfig := struct{}{}
	processor, err = newTruthBeamProcessor(wrongConfig, processortest.NewNopSettings(component.MustNewType("test")))
	assert.Error(t, err, "Expected error with wrong config type")
	assert.Contains(t, err.Error(), "invalid configuration provided")
	assert.Nil(t, processor, "Processor should be nil with wrong config type")
}

func TestProcessLogs(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/enrich", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req client.EnrichmentRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "test-policy-123", req.Evidence.PolicyId)
		assert.Equal(t, "test-source", req.Evidence.Source)
		assert.Equal(t, "compliant", req.Evidence.Decision)
		assert.Equal(t, "audit", req.Evidence.Action)

		response := client.EnrichmentResponse{
			Compliance: client.Compliance{
				Catalog:      "NIST-800-53",
				Category:     "Access Control",
				Control:      "AC-1",
				Remediation:  stringPtr("Implement proper access controls"),
				Requirements: []string{"req-1", "req-2"},
				Standards:    []string{"NIST-800-53", "ISO-27001"},
			},
			Status: client.Status{
				Id:    statusIdPtr(1),
				Title: "Pass",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	processor := createTestProcessor(t, mockServer.URL)
	logs := createTestLogs()
	setRequiredAttributes(logs)

	ctx := context.Background()
	result, err := processor.processLogs(ctx, logs)
	require.NoError(t, err)
	require.NotNil(t, result)

	processedLogRecord := result.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
	attrs := processedLogRecord.Attributes()

	// Verify compliance attributes were added
	assert.Equal(t, "Pass", attrs.AsRaw()["compliance.status"])
	assert.Equal(t, "AC-1", attrs.AsRaw()["compliance.control.id"])
	assert.Equal(t, "NIST-800-53", attrs.AsRaw()["compliance.control.catalog.id"])
	assert.Equal(t, "Access Control", attrs.AsRaw()["compliance.category"])
	assert.Equal(t, "Implement proper access controls", attrs.AsRaw()["compliance.control.remediation.description"])

	requirements := attrs.AsRaw()["compliance.requirements"].([]interface{})
	assert.Len(t, requirements, 2)
	assert.Contains(t, requirements, "req-1")
	assert.Contains(t, requirements, "req-2")

	standards := attrs.AsRaw()["compliance.standards"].([]interface{})
	assert.Len(t, standards, 2)
	assert.Contains(t, standards, "NIST-800-53")
	assert.Contains(t, standards, "ISO-27001")
}

func TestProcessLogsWithMissingAttributes(t *testing.T) {
	processor := createTestProcessor(t, "http://localhost:8081")
	logs := createTestLogs()
	logRecord := logs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)

	// Missing policy.id attribute
	logRecord.Attributes().PutStr("policy.source", "test-source")
	logRecord.Attributes().PutStr("policy.evaluation.status", "compliant")
	logRecord.Attributes().PutStr("policy.enforcement.action", "audit")

	ctx := context.Background()
	result, err := processor.processLogs(ctx, logs)
	require.NoError(t, err, "Processor should not fail even with missing attributes")
	require.NotNil(t, result)
}

func TestProcessLogsWithHTTPError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := client.Error{
			Code:    500,
			Message: "Internal server error",
		}
		_ = json.NewEncoder(w).Encode(errorResponse)
	}))
	defer mockServer.Close()

	processor := createTestProcessor(t, mockServer.URL)
	logs := createTestLogs()
	setRequiredAttributes(logs)

	ctx := context.Background()
	result, err := processor.processLogs(ctx, logs)
	require.NoError(t, err, "Processor should not fail even with HTTP errors")
	require.NotNil(t, result)
}

// Helper functions
func createTestProcessor(t *testing.T, endpoint string) *truthBeamProcessor {
	cfg := &Config{
		ClientConfig: confighttp.NewDefaultClientConfig(),
	}
	cfg.ClientConfig.Endpoint = endpoint

	settings := processortest.NewNopSettings(component.MustNewType("test"))
	settings.Logger = zaptest.NewLogger(t)

	processor, err := newTruthBeamProcessor(cfg, settings)
	require.NoError(t, err)
	return processor
}

func createTestLogs() plog.Logs {
	logs := plog.NewLogs()
	resourceLogs := logs.ResourceLogs().AppendEmpty()
	scopeLogs := resourceLogs.ScopeLogs().AppendEmpty()
	logRecord := scopeLogs.LogRecords().AppendEmpty()
	logRecord.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	return logs
}

func setRequiredAttributes(logs plog.Logs) {
	logRecord := logs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
	logRecord.Attributes().PutStr("policy.id", "test-policy-123")
	logRecord.Attributes().PutStr("policy.source", "test-source")
	logRecord.Attributes().PutStr("policy.evaluation.status", "compliant")
	logRecord.Attributes().PutStr("policy.enforcement.action", "audit")
}

func stringPtr(s string) *string {
	return &s
}

func statusIdPtr(id int) *client.StatusId {
	statusId := client.StatusId(id)
	return &statusId
}
