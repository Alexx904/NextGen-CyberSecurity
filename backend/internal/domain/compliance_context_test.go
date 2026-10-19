package domain

import (
	"encoding/json"
	"github.com/google/uuid"
	"reflect"
	"testing"
)

func TestRegulatoryContextEvidenceBoundaries(t *testing.T) {
	cve, http := "CVE-2021-41773", "http"
	f := Finding{ID: uuid.New(), ScanID: uuid.New(), HostIP: "192.0.2.4", Port: 80, Protocol: "tcp", PortState: "open", CVEID: &cve, ServiceName: &http, Confidence: ConfidenceProbable, DetectionMethod: DetectionMethodBanner}
	context := BuildRegulatoryContext(f)
	if len(context.Implications) != 2 || context.Confidence != ConfidenceProbable || context.DetectionMethod != DetectionMethodBanner {
		t.Fatalf("lost evidence: %+v", context)
	}
	if context.Implications[0].References[0].ID != "NIS2 Art. 21(2)(e)" || context.Implications[1].References[0].ID != "NIS2 Art. 21(2)(h)" {
		t.Fatal("incorrect mapping")
	}
	for _, state := range []string{"closed", "filtered", "unfiltered", "closed|filtered"} {
		f.PortState = state
		if len(BuildRegulatoryContext(f).Implications) != 0 {
			t.Fatalf("mapped negative state %s", state)
		}
	}
	f.PortState = "open"
	f.Port = 0
	if len(BuildRegulatoryContext(f).Implications) != 0 {
		t.Fatal("mapped host discovery")
	}
	f.Port = 443
	f.CVEID = nil
	https := "https"
	f.ServiceName = &https
	if len(BuildRegulatoryContext(f).Implications) != 0 {
		t.Fatal("invented issue from HTTPS")
	}
	blank := ""
	f.CVEID = &blank
	if len(BuildRegulatoryContext(f).Implications) != 0 {
		t.Fatal("mapped empty CVE")
	}
}

func TestRegulatoryContextSSEMatchesDirectContext(t *testing.T) {
	service := "telnet"
	f := Finding{ID: uuid.New(), ScanID: uuid.New(), TenantID: uuid.New(), HostIP: "192.0.2.3", Port: 23, PortState: "open", ServiceName: &service}
	encoded, err := json.Marshal(ScanFindingPayload{ScanID: f.ScanID, Finding: f})
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		ScanID  uuid.UUID `json:"scan_id"`
		Finding struct {
			ID       uuid.UUID         `json:"id"`
			TenantID uuid.UUID         `json:"tenant_id"`
			Context  RegulatoryContext `json:"regulatory_context"`
		} `json:"finding"`
	}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ScanID != f.ScanID || decoded.Finding.ID != f.ID || decoded.Finding.TenantID != f.TenantID || !reflect.DeepEqual(decoded.Finding.Context, BuildRegulatoryContext(f)) {
		t.Fatalf("SSE context mismatch: %s", encoded)
	}
}
