package report_test

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/pkg/report"
)

func TestGenerateCSV_Success(t *testing.T) {
	now := time.Now().UTC()
	scanID := uuid.New()
	findingID := uuid.New()
	tenantID := uuid.New()

	hostname := "target.local"
	svcName := "http"
	svcVer := "Apache 2.4.49"
	cvss := 9.8
	cve := "CVE-2021-41773"
	banner := "Apache/2.4.49 (Unix)"
	remediation := "Upgrade Apache"

	findings := []*domain.Finding{
		{
			ID:             findingID,
			ScanID:         scanID,
			TenantID:       tenantID,
			HostIP:         "192.168.1.100",
			Hostname:       &hostname,
			Port:           80,
			Protocol:       "tcp",
			ServiceName:    &svcName,
			ServiceVersion: &svcVer,
			Severity:       domain.SeverityCritical,
			CVSSScore:      &cvss,
			CVEID:          &cve,
			RawBanner:      &banner,
			Remediation:    &remediation,
			CreatedAt:      now,
		},
	}

	csvBytes, err := report.GenerateCSV(findings)
	if err != nil {
		t.Fatalf("generazione CSV fallita: %v", err)
	}

	r := csv.NewReader(bytes.NewReader(csvBytes))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("lettura CSV fallita: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("attesi 2 record (header + 1 riga), ottenuti %d", len(records))
	}

	header := records[0]
	if header[0] != "finding_id" || header[2] != "host_ip" || header[8] != "severity" {
		t.Errorf("header CSV non conforme: %v", header)
	}

	row := records[1]
	if row[0] != findingID.String() {
		t.Errorf("finding_id errato: %s", row[0])
	}
	if row[2] != "192.168.1.100" {
		t.Errorf("host_ip errato: %s", row[2])
	}
	if row[3] != hostname {
		t.Errorf("hostname errato: %s", row[3])
	}
	if row[4] != "80" {
		t.Errorf("port errata: %s", row[4])
	}
	if row[8] != string(domain.SeverityCritical) {
		t.Errorf("severity errata: %s", row[8])
	}
	if row[9] != "9.8" {
		t.Errorf("cvss_score errato: %s", row[9])
	}
	if row[10] != cve {
		t.Errorf("cve_id errato: %s", row[10])
	}
}

func TestGenerateCSV_RFC4180_Escaping(t *testing.T) {
	bannerWithSpecials := "Banner with \"quotes\", commas, and\nnewlines"
	remWithSpecials := "Action: step 1, step 2, \"critical\"\nstep 3"

	findings := []*domain.Finding{
		{
			ID:          uuid.New(),
			ScanID:      uuid.New(),
			TenantID:    uuid.New(),
			HostIP:      "10.0.0.1",
			Port:        443,
			Protocol:    "tcp",
			Severity:    domain.SeverityHigh,
			RawBanner:   &bannerWithSpecials,
			Remediation: &remWithSpecials,
			CreatedAt:   time.Now().UTC(),
		},
	}

	csvBytes, err := report.GenerateCSV(findings)
	if err != nil {
		t.Fatalf("generazione CSV con caratteri speciali fallita: %v", err)
	}

	// Verifica parsing conforme RFC 4180
	r := csv.NewReader(bytes.NewReader(csvBytes))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("lettura CSV con escape fallita: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("attesi 2 record, ottenuti %d", len(records))
	}

	row := records[1]
	if row[11] != bannerWithSpecials {
		t.Errorf("valore raw_banner non preservato correttamente: %q vs %q", row[11], bannerWithSpecials)
	}
	if row[12] != remWithSpecials {
		t.Errorf("valore remediation non preservato correttamente: %q vs %q", row[12], remWithSpecials)
	}
}

func TestGenerateCSV_NilFields(t *testing.T) {
	findings := []*domain.Finding{
		{
			ID:        uuid.New(),
			ScanID:    uuid.New(),
			TenantID:  uuid.New(),
			HostIP:    "192.168.1.1",
			Port:      22,
			Protocol:  "tcp",
			Severity:  domain.SeverityInfo,
			CreatedAt: time.Now().UTC(),
		},
	}

	csvBytes, err := report.GenerateCSV(findings)
	if err != nil {
		t.Fatalf("generazione CSV fallita con campi nil: %v", err)
	}

	r := csv.NewReader(bytes.NewReader(csvBytes))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("lettura CSV fallita: %v", err)
	}

	row := records[1]
	// Hostname, ServiceName, ServiceVersion, CVSSScore, CVEID, RawBanner, Remediation devono essere ""
	if row[3] != "" || row[6] != "" || row[7] != "" || row[9] != "" || row[10] != "" || row[11] != "" || row[12] != "" {
		t.Errorf("i campi nil devono essere stringhe vuote, ottenuto: %v", row)
	}
}

func TestGenerateCSV_EmptyFindings(t *testing.T) {
	csvBytes, err := report.GenerateCSV([]*domain.Finding{})
	if err != nil {
		t.Fatalf("generazione CSV con findings vuoti fallita: %v", err)
	}

	r := csv.NewReader(bytes.NewReader(csvBytes))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("lettura CSV vuoto fallita: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("atteso solo 1 record (header), ottenuti %d", len(records))
	}
	if !strings.HasPrefix(records[0][0], "finding_id") {
		t.Errorf("header inatteso: %v", records[0])
	}
}

