package service_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"nextgen-cybersec/internal/domain"
	"nextgen-cybersec/internal/service"
)

func TestNVDOnlineClient_SearchByCPE(t *testing.T) {
	mockJSON := `{
		"totalResults": 1,
		"vulnerabilities": [
			{
				"cve": {
					"id": "CVE-2021-41773",
					"descriptions": [
						{"lang": "en", "value": "A flaw was found in path normalization in Apache HTTP Server 2.4.49"}
					],
					"metrics": {
						"cvssMetricV31": [
							{
								"cvssData": {
									"baseScore": 7.5,
									"baseSeverity": "HIGH"
								}
							}
						]
					},
					"published": "2021-10-05T00:00:00.000"
				}
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("cpeName") == "" {
			t.Errorf("cpeName non fornito nella query: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockJSON))
	}))
	defer server.Close()

	client := service.NewNVDOnlineClient()
	client.SetBaseURL(server.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	entries, err := client.SearchByCPE(ctx, "cpe:2.3:a:apache:http_server:2.4.49:*:*:*:*:*:*:*")
	if err != nil {
		t.Fatalf("SearchByCPE fallito: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("attesa 1 voce CVE, trovate %d", len(entries))
	}

	e := entries[0]
	if e.CVEID != "CVE-2021-41773" {
		t.Errorf("CVEID atteso CVE-2021-41773, ottenuto %s", e.CVEID)
	}
	if e.Severity != domain.SeverityHigh {
		t.Errorf("Severity attesa HIGH, ottenuta %s", e.Severity)
	}
	if e.CVSSScore == nil || *e.CVSSScore != 7.5 {
		t.Errorf("CVSSScore atteso 7.5, ottenuto %v", e.CVSSScore)
	}
}

func TestNVDOnlineClient_SearchByKeyword(t *testing.T) {
	mockJSON := `{
		"totalResults": 1,
		"vulnerabilities": [
			{
				"cve": {
					"id": "CVE-2024-6387",
					"descriptions": [
						{"lang": "en", "value": "regreSSHion vulnerability in OpenSSH"}
					],
					"metrics": {
						"cvssMetricV31": [
							{
								"cvssData": {
									"baseScore": 8.1,
									"baseSeverity": "HIGH"
								}
							}
						]
					},
					"published": "2024-07-01T00:00:00.000"
				}
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("keywordSearch") != "OpenSSH 9.7p1" {
			t.Errorf("keywordSearch inatteso: %s", r.URL.Query().Get("keywordSearch"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockJSON))
	}))
	defer server.Close()

	client := service.NewNVDOnlineClient()
	client.SetBaseURL(server.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	entries, err := client.SearchByKeyword(ctx, "OpenSSH 9.7p1")
	if err != nil {
		t.Fatalf("SearchByKeyword fallito: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("attesa 1 voce CVE, trovate %d", len(entries))
	}
	if entries[0].CVEID != "CVE-2024-6387" {
		t.Errorf("CVEID atteso CVE-2024-6387, ottenuto %s", entries[0].CVEID)
	}
}

