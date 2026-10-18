package domain

import "github.com/google/uuid"

type VirtualPatchFile struct {
	Engine   string `json:"engine"`
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

// VirtualPatchBundle is advisory output only; it is never executed by the server.
type VirtualPatchBundle struct {
	FindingID         uuid.UUID          `json:"finding_id"`
	CVEID             string             `json:"cve_id"`
	CatalogRevision   string             `json:"catalog_revision"`
	Supported         bool               `json:"supported"`
	Product           string             `json:"product"`
	Strategy          string             `json:"strategy"`
	OperationalImpact string             `json:"operational_impact"`
	CoveredCVEs       []string           `json:"covered_cves"`
	Reason            string             `json:"reason"`
	Limitations       []string           `json:"limitations"`
	Sources           []string           `json:"sources"`
	Files             []VirtualPatchFile `json:"files"`
}
