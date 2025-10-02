package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
)

// ApplyAttributes enriches attributes in the log record with compliance impact data.
func ApplyAttributes(ctx context.Context, client *Client, serverURL string, _ pcommon.Resource, logRecord plog.LogRecord) error {
	attrs := logRecord.Attributes()

	// Retrieve lookup attributes
	policyIDVal, ok := attrs.Get(POLICY_ID)
	if !ok {
		return fmt.Errorf("missing required attribute %q", POLICY_ID)
	}

	policyAction, ok := attrs.Get(POLICY_ENFORCEMENT_ACTION)
	if !ok {
		return fmt.Errorf("missing required attribute %q", POLICY_ENFORCEMENT_ACTION)
	}

	policySourceVal, ok := attrs.Get(POLICY_SOURCE)
	if !ok {
		return fmt.Errorf("missing required attribute %q", POLICY_SOURCE)
	}

	policyDecisionVal, ok := attrs.Get(POLICY_EVALUATION_STATUS)
	if !ok {
		return fmt.Errorf("missing required attributes %q", POLICY_EVALUATION_STATUS)
	}
	enrichReq := EnrichmentRequest{
		Evidence: Evidence{
			Timestamp: logRecord.Timestamp().AsTime(),
			Source:    policySourceVal.Str(),
			PolicyId:  policyIDVal.Str(),
			Decision:  policyDecisionVal.Str(),
			Action:    policyAction.Str(),
		},
	}

	enrichRes, err := callEnrichAPI(ctx, client, serverURL, enrichReq)
	if err != nil {
		return err
	}

	attrs.PutStr(COMPLIANCE_STATUS, string(enrichRes.Status.Title))
	attrs.PutStr(COMPLIANCE_CONTROL_ID, enrichRes.Compliance.Control)
	attrs.PutStr(COMPLIANCE_CONTROL_CATALOG_ID, enrichRes.Compliance.Catalog)
	attrs.PutStr(COMPLIANCE_CATEGORY, enrichRes.Compliance.Category)
	requirements := attrs.PutEmptySlice(COMPLIANCE_REQUIREMENTS)
	standards := attrs.PutEmptySlice(COMPLIANCE_STANDARDS)

	if enrichRes.Compliance.Remediation != nil {
		attrs.PutStr(COMPLIANCE_CONTROL_REMEDIATION_DESCRIPTION, *enrichRes.Compliance.Remediation)
	}

	for _, req := range enrichRes.Compliance.Requirements {
		newReq := requirements.AppendEmpty()
		newReq.SetStr(req)
	}
	for _, std := range enrichRes.Compliance.Standards {
		newStd := standards.AppendEmpty()
		newStd.SetStr(std)
	}

	return nil
}

// callEnrichAPI is a helper function to perform the actual HTTP request.
func callEnrichAPI(ctx context.Context, client *Client, serverURL string, req EnrichmentRequest) (*EnrichmentResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	// Create the HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", serverURL+"/v1/enrich", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Perform the request
	resp, err := client.Client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Handle non-200 status codes
	if resp.StatusCode != http.StatusOK {
		var errRes Error
		err := json.NewDecoder(resp.Body).Decode(&errRes)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("API call failed with status %d: %v", resp.StatusCode, errRes.Message)
	}

	// Decode the successful response
	var enrichRes EnrichmentResponse
	if err := json.NewDecoder(resp.Body).Decode(&enrichRes); err != nil {
		return nil, err
	}

	return &enrichRes, nil
}
