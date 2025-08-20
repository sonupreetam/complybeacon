package evidence

import (
	"encoding/json"
	"time"

	"github.com/in-toto/go-witness/cryptoutil"
)

// EvidenceEvent represents a higher-level, mapped conformance assertion.
type EvidenceEvent struct {
	Summary   string      `json:"summary"`
	Timestamp time.Time   `json:"timestamp"`
	Evidence  RawEvidence `json:"evidence"`
}

func NewFromEvidence(rawEnv RawEvidence) EvidenceEvent {
	event := EvidenceEvent{
		Timestamp: time.Now(),
		Evidence:  rawEnv,
	}
	return event
}

// RawEvidence represents a simplified raw output from a policy engine.
type RawEvidence struct {
	Metadata `json:",inline"`
	Details  []Resource `json:"details"`
}

type Metadata struct {
	ID        string    `json:"id"`
	Collected time.Time `json:"collected"`
	Source    string    `json:"source"`
	PolicyID  string    `json:"policyId"`
	Decision  string    `json:"decision"`
	Subject   Resource  `json:"subject"`
}

type Resource struct {
	Name      string               `json:"name"`
	URI       string               `json:"uri"`
	Content   json.RawMessage      `json:"content"`
	Digest    cryptoutil.DigestSet `json:"digest"`
	MediaType string               `json:"mediaType"`
}
