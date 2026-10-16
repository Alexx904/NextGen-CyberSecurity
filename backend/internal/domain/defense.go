package domain

import "github.com/google/uuid"

// RemediationRequest selects a registered server; rule inputs are server-derived.
type RemediationRequest struct {
	ServerID uuid.UUID `json:"server_id"`
}

type RemediationResult struct {
	NodeID               string             `json:"node_id"`
	ServerID             uuid.UUID          `json:"server_id"`
	Rule                 FirewallRuleParams `json:"rule"`
	Command              string             `json:"command"`
	SafetyTimeoutSeconds int                `json:"safety_timeout_seconds"`
	Firewall             *FirewallStatus    `json:"firewall,omitempty"`
}
