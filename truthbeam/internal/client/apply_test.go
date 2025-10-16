package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
)

// TestApplyAttributes tests the ApplyAttributes functionality with a valid response.
func TestApplyAttributes(t *testing.T) {
	// Create a mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/enrich", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Parse the request body
		var req EnrichmentRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// Verify request content
		assert.Equal(t, "test-policy-123", req.Evidence.PolicyId)
		assert.Equal(t, "test-source", req.Evidence.Source)
		assert.Equal(t, "compliant", req.Evidence.Decision)
		assert.Equal(t, "audit", req.Evidence.Action)

		// Return a mock response
		response := EnrichmentResponse{
			Compliance: Compliance{
				Catalog:      "NIST-800-53",
				Category:     "Access Control",
				Control:      "AC-1",
				Remediation:  stringPtr("Implement proper access controls"),
				Requirements: []string{"req-1", "req-2"},
				Standards:    []string{"NIST-800-53", "ISO-27001"},
			},
			Status: Status{
				Id:    statusIdPtr(1),
				Title: "Pass",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	// Create client
	client, err := NewClient(mockServer.URL)
	require.NoError(t, err)

	// Create test log record
	logRecord, resource := createTestLogRecord()
	logRecord.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	attrs := logRecord.Attributes()

	// Apply attributes for log enrichment
	ctx := context.Background()
	err = ApplyAttributes(ctx, client, mockServer.URL, resource, logRecord)
	require.NoError(t, err)

	// Verify that compliance attributes were added
	assert.Equal(t, "Pass", attrs.AsRaw()[COMPLIANCE_STATUS])
	assert.Equal(t, "AC-1", attrs.AsRaw()[COMPLIANCE_CONTROL_ID])
	assert.Equal(t, "NIST-800-53", attrs.AsRaw()[COMPLIANCE_CONTROL_CATALOG_ID])
	assert.Equal(t, "Access Control", attrs.AsRaw()[COMPLIANCE_CATEGORY])
	assert.Equal(t, "Implement proper access controls", attrs.AsRaw()[COMPLIANCE_CONTROL_REMEDIATION_DESCRIPTION])

	// Check requirements and standards arrays
	requirements := attrs.AsRaw()[COMPLIANCE_REQUIREMENTS].([]interface{})
	assert.Len(t, requirements, 2)
	assert.Contains(t, requirements, "req-1")
	assert.Contains(t, requirements, "req-2")

	standards := attrs.AsRaw()[COMPLIANCE_STANDARDS].([]interface{})
	assert.Len(t, standards, 2)
	assert.Contains(t, standards, "NIST-800-53")
	assert.Contains(t, standards, "ISO-27001")
}

// TestApplyAttributesWithMissingPolicyId tests the ApplyAttributes functionality with a missing policy.id attribute.
func TestApplyAttributesWithMissingPolicyId(t *testing.T) {
	// Create client
	client, err := NewClient("http://localhost:8081")
	require.NoError(t, err)

	// Create test log record without policy.id
	logRecord := plog.NewLogRecord()
	attrs := logRecord.Attributes()
	attrs.PutStr(POLICY_SOURCE, "test-source")
	attrs.PutStr(POLICY_EVALUATION_STATUS, "compliant")
	attrs.PutStr(POLICY_ENFORCEMENT_ACTION, "audit")

	resource := pcommon.NewResource()

	// Apply attributes should fail since no policy.id present
	ctx := context.Background()
	err = ApplyAttributes(ctx, client, "http://localhost:8081", resource, logRecord)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required attribute")
	assert.Contains(t, err.Error(), POLICY_ID)
}

func TestApplyAttributesWithMissingPolicySource(t *testing.T) {
	// Create client
	client, err := NewClient("http://localhost:8081")
	require.NoError(t, err)

	// Create test log record without policy.source
	logRecord := plog.NewLogRecord()
	attrs := logRecord.Attributes()
	attrs.PutStr(POLICY_ID, "test-policy-123")
	attrs.PutStr(POLICY_EVALUATION_STATUS, "compliant")
	attrs.PutStr(POLICY_ENFORCEMENT_ACTION, "audit")

	resource := pcommon.NewResource()

	// Apply attributes should fail
	ctx := context.Background()
	err = ApplyAttributes(ctx, client, "http://localhost:8081", resource, logRecord)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required attribute")
	assert.Contains(t, err.Error(), POLICY_SOURCE)
}

func TestApplyAttributesWithMissingPolicyEvaluationStatus(t *testing.T) {
	// Create client
	client, err := NewClient("http://localhost:8081")
	require.NoError(t, err)

	// Create test log record without policy.evaluation.status
	logRecord := plog.NewLogRecord()
	attrs := logRecord.Attributes()
	attrs.PutStr(POLICY_ID, "test-policy-123")
	attrs.PutStr(POLICY_SOURCE, "test-source")
	attrs.PutStr(POLICY_ENFORCEMENT_ACTION, "audit")

	resource := pcommon.NewResource()

	// Apply attributes should fail
	ctx := context.Background()
	err = ApplyAttributes(ctx, client, "http://localhost:8081", resource, logRecord)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required attribute")
	assert.Contains(t, err.Error(), POLICY_EVALUATION_STATUS)
}

func TestApplyAttributesWithMissingPolicyEnforcementAction(t *testing.T) {
	// Create client
	client, err := NewClient("http://localhost:8081")
	require.NoError(t, err)

	// Create test log record and remove policy.enforcement.action
	logRecord, resource := createTestLogRecord()
	logRecord.Attributes().Remove(POLICY_ENFORCEMENT_ACTION)

	// Apply attributes should fail
	ctx := context.Background()
	err = ApplyAttributes(ctx, client, "http://localhost:8081", resource, logRecord)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required attribute")
	assert.Contains(t, err.Error(), POLICY_ENFORCEMENT_ACTION)
}

func TestApplyAttributesWithHTTPError(t *testing.T) {
	// Create a mock HTTP server that returns an error
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := Error{
			Code:    500,
			Message: "Internal server error",
		}
		json.NewEncoder(w).Encode(errorResponse)
	}))
	defer mockServer.Close()

	// Create client
	client, err := NewClient(mockServer.URL)
	require.NoError(t, err)

	// Create test log record
	logRecord, resource := createTestLogRecord()

	// Apply attributes should fail
	ctx := context.Background()
	err = ApplyAttributes(ctx, client, mockServer.URL, resource, logRecord)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API call failed with status 500")
}

func TestApplyAttributesWithNetworkError(t *testing.T) {
	// Create client with invalid URL
	client, err := NewClient("http://invalid-host:9999")
	require.NoError(t, err)

	// Create test log record
	logRecord, resource := createTestLogRecord()

	// Apply attributes should fail due to network error
	ctx := context.Background()
	err = ApplyAttributes(ctx, client, "http://invalid-host:9999", resource, logRecord)
	assert.Error(t, err)
	// Network error message will vary, just check that it's an error
}

func TestApplyAttributesWithEmptyResponse(t *testing.T) {
	// Create a mock HTTP server that returns empty response
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
	}))
	defer mockServer.Close()

	// Create client
	client, err := NewClient(mockServer.URL)
	require.NoError(t, err)

	// Create test log record
	logRecord, resource := createTestLogRecord()

	// Apply attributes should succeed but with empty compliance data
	ctx := context.Background()
	err = ApplyAttributes(ctx, client, mockServer.URL, resource, logRecord)
	require.NoError(t, err)

	// Check that empty compliance attributes were added
	attrs := logRecord.Attributes()
	assert.Equal(t, "", attrs.AsRaw()[COMPLIANCE_STATUS])
	assert.Equal(t, "", attrs.AsRaw()[COMPLIANCE_CONTROL_ID])
	assert.Equal(t, "", attrs.AsRaw()[COMPLIANCE_CONTROL_CATALOG_ID])
	assert.Equal(t, "", attrs.AsRaw()[COMPLIANCE_CATEGORY])

	// Check that empty arrays were added
	requirements := attrs.AsRaw()[COMPLIANCE_REQUIREMENTS].([]interface{})
	assert.Len(t, requirements, 0)

	standards := attrs.AsRaw()[COMPLIANCE_STANDARDS].([]interface{})
	assert.Len(t, standards, 0)
}

func TestApplyAttributesWithNilRemediation(t *testing.T) {
	// Create a mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a response with nil remediation
		response := EnrichmentResponse{
			Compliance: Compliance{
				Catalog:      "NIST-800-53",
				Category:     "Access Control",
				Control:      "AC-1",
				Remediation:  nil, // No remediation
				Requirements: []string{"req-1"},
				Standards:    []string{"NIST-800-53"},
			},
			Status: Status{
				Id:    statusIdPtr(1),
				Title: "Pass",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	// Create client
	client, err := NewClient(mockServer.URL)
	require.NoError(t, err)

	// Create test log record
	logRecord, resource := createTestLogRecord()

	// Apply attributes
	ctx := context.Background()
	err = ApplyAttributes(ctx, client, mockServer.URL, resource, logRecord)
	require.NoError(t, err)

	// Check that remediation attribute was not added
	attrs := logRecord.Attributes()
	_, exists := attrs.Get(COMPLIANCE_CONTROL_REMEDIATION_DESCRIPTION)
	assert.False(t, exists, "Remediation attribute should not exist when nil")
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func statusIdPtr(id int) *StatusId {
	statusId := StatusId(id)
	return &statusId
}

// createTestLogRecord is a helper function for easy test setup
func createTestLogRecord() (plog.LogRecord, pcommon.Resource) {
	logRecord := plog.NewLogRecord()
	attrs := logRecord.Attributes()
	attrs.PutStr(POLICY_ID, "test-policy-123")
	attrs.PutStr(POLICY_SOURCE, "test-source")
	attrs.PutStr(POLICY_EVALUATION_STATUS, "compliant")
	attrs.PutStr(POLICY_ENFORCEMENT_ACTION, "audit")

	resource := pcommon.NewResource()
	return logRecord, resource
}
