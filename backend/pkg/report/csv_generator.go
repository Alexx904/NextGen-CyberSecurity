package report

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"nextgen-cybersec/internal/domain"
)

// GenerateCSV serializza un insieme di finding in formato CSV conforme allo standard RFC 4180
func GenerateCSV(findings []*domain.Finding) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Intestazione colonne CSV
	header := []string{
		"finding_id",
		"scan_id",
		"host_ip",
		"hostname",
		"port",
		"protocol",
		"service_name",
		"service_version",
		"severity",
		"cvss_score",
		"cve_id",
		"raw_banner",
		"remediation",
		"created_at",
	}

	if err := w.Write(header); err != nil {
		return nil, fmt.Errorf("scrittura header CSV fallita: %w", err)
	}

	for _, f := range findings {
		// Esclusione categorica di record nulli o host non responsivi/vuoti (DoD TASK-36)
		if f == nil || strings.TrimSpace(f.HostIP) == "" || strings.EqualFold(f.HostIP, "0.0.0.0") {
			continue
		}

		hostname := ""
		if f.Hostname != nil {
			hostname = *f.Hostname
		}

		serviceName := ""
		if f.ServiceName != nil {
			serviceName = *f.ServiceName
		}

		serviceVersion := ""
		if f.ServiceVersion != nil {
			serviceVersion = *f.ServiceVersion
		}

		cvssScore := ""
		if f.CVSSScore != nil {
			cvssScore = fmt.Sprintf("%.1f", *f.CVSSScore)
		}

		cveID := ""
		if f.CVEID != nil {
			cveID = *f.CVEID
		}

		rawBanner := ""
		if f.RawBanner != nil {
			rawBanner = *f.RawBanner
		}

		remediation := ""
		if f.Remediation != nil {
			remediation = *f.Remediation
		}

		createdAt := f.CreatedAt.UTC().Format(time.RFC3339)

		row := []string{
			f.ID.String(),
			f.ScanID.String(),
			f.HostIP,
			hostname,
			fmt.Sprintf("%d", f.Port),
			f.Protocol,
			serviceName,
			serviceVersion,
			string(f.Severity),
			cvssScore,
			cveID,
			rawBanner,
			remediation,
			createdAt,
		}

		if err := w.Write(row); err != nil {
			return nil, fmt.Errorf("scrittura riga CSV fallita: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("flush CSV fallito: %w", err)
	}

	return buf.Bytes(), nil
}

