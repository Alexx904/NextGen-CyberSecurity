package handler

import (
	"encoding/json"
	"github.com/google/uuid"
	"nextgen-cybersec/internal/domain"
	"testing"
)

func TestFindingResponseAutomaticRegulatoryContext(t *testing.T) {
	service := "telnet"
	f := &domain.Finding{ID: uuid.New(), ScanID: uuid.New(), TenantID: uuid.New(), HostIP: "192.0.2.2", Port: 23, Protocol: "tcp", PortState: "open", ServiceName: &service, Confidence: domain.ConfidenceProbable, DetectionMethod: domain.DetectionMethodBanner}
	response := toFindingResponse(f)
	if response.ID != f.ID.String() || response.TenantID != f.TenantID.String() || response.PortState != "open" || response.Confidence != f.Confidence || len(response.RegulatoryContext.Implications) != 1 {
		t.Fatalf("context/metadata lost: %+v", response)
	}
	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]any
	if err := json.Unmarshal(encoded, &object); err != nil {
		t.Fatal(err)
	}
	if object["regulatory_context"] == nil || object["detection_method"] != "banner" {
		t.Fatal("REST contract missing context")
	}
	f.PortState = "closed"
	if len(toFindingResponse(f).RegulatoryContext.Implications) != 0 {
		t.Fatal("negative result mapped")
	}
}
