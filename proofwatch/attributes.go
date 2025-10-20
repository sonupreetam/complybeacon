// DO NOT EDIT, this is an auto-generated file

package proofwatch

// A unique identifier for a specific audit or assessment.
//
// Notes:
// This aligns with a id field of a `gemara` Evaluation or Enforcement Plan. The goal to allow batching
// and correlation of a set of findings
const COMPLIANCE_AUDIT_ID = "compliance.audit.id"

// The category a control framework pertains
const COMPLIANCE_CATEGORY = "compliance.category"

// The unique identifier for the security control catalog
const COMPLIANCE_CONTROL_CATALOG_ID = "compliance.control.catalog.id"

// The unique identifier for the security control.
//
// Notes:
// A control is a prescriptive, actionable set of
// specifications that strengthens security and compliance posture. This value may also reference
// a specific control part
const COMPLIANCE_CONTROL_ID = "compliance.control.id"

// The description of the remediation strategy
const COMPLIANCE_CONTROL_REMEDIATION_DESCRIPTION = "compliance.control.remediation.description"

// The identifiers specific compliance requirements being evaluated
const COMPLIANCE_REQUIREMENTS = "compliance.requirements"

// The risk level associated with non-compliance
const COMPLIANCE_RISK_LEVEL = "compliance.risk.level"

// The identifiers for regulatory or industry standards being evaluated for compliance
const COMPLIANCE_STANDARDS = "compliance.standards"

// The normalized status identifier of the compliance check
const COMPLIANCE_STATUS = "compliance.status"

// The action take by the policy enforcement
const POLICY_ENFORCEMENT_ACTION = "policy.enforcement.action"

// The outcome of the policy enforcement.
//
// Notes:
// This is required if the policy enforcement action is not "audit"
const POLICY_ENFORCEMENT_STATUS = "policy.enforcement.status"

// The outcome of the policy evaluation (e.g., "deny", "allow")
const POLICY_EVALUATION_STATUS = "policy.evaluation.status"

// The subject id or resource the policy was applied to
const POLICY_EVALUATION_SUBJECT_ID = "policy.evaluation.subject.id"

// The subject type or resource the policy was applied to
const POLICY_EVALUATION_SUBJECT_TYPE = "policy.evaluation.subject.type"

// The identifier for the policy that was applied
const POLICY_ID = "policy.id"

// The human-readable name of the policy
const POLICY_NAME = "policy.name"

// The identifier for the source of the policy audit log.
//
// Notes:
// This should identify the policy engine or assessment tool
const POLICY_SOURCE = "policy.source"

// Add contextual details around the policy status
const POLICY_STATUS_DETAIL = "policy.status.detail"
