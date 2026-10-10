package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"nextgen-cybersec/internal/domain"
)

// OnlineCVEClient definisce l'interfaccia per il recupero online di vulnerabilità CVE
type OnlineCVEClient interface {
	SearchByCPE(ctx context.Context, cpe23URI string) ([]*domain.CVEEntry, error)
	SearchByKeyword(ctx context.Context, keyword string) ([]*domain.CVEEntry, error)
}

// NVDOnlineClient implementa OnlineCVEClient interrogando le API REST v2.0 ufficiali del NIST NVD
type NVDOnlineClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewNVDOnlineClient istanzia un client verso le API NIST NVD con timeout sicuro (4s)
func NewNVDOnlineClient() *NVDOnlineClient {
	return &NVDOnlineClient{
		baseURL: "https://services.nvd.nist.gov/rest/json/cves/2.0",
		httpClient: &http.Client{
			Timeout: 4 * time.Second,
		},
	}
}

// SetBaseURL consente di sovrascrivere l'endpoint di base (utile per test con mock server)
func (c *NVDOnlineClient) SetBaseURL(url string) {
	c.baseURL = url
}

type nvdResponse struct {
	TotalResults    int `json:"totalResults"`
	Vulnerabilities []struct {
		CVE struct {
			ID           string `json:"id"`
			Descriptions []struct {
				Lang  string `json:"lang"`
				Value string `json:"value"`
			} `json:"descriptions"`
			Metrics struct {
				CvssMetricV31 []struct {
					CvssData struct {
						BaseScore    float64 `json:"baseScore"`
						BaseSeverity string  `json:"baseSeverity"`
					} `json:"cvssData"`
				} `json:"cvssMetricV31"`
				CvssMetricV30 []struct {
					CvssData struct {
						BaseScore    float64 `json:"baseScore"`
						BaseSeverity string  `json:"baseSeverity"`
					} `json:"cvssData"`
				} `json:"cvssMetricV30"`
				CvssMetricV2 []struct {
					CvssData struct {
						BaseScore float64 `json:"baseScore"`
					} `json:"cvssData"`
					BaseSeverity string `json:"baseSeverity"`
				} `json:"cvssMetricV2"`
			} `json:"metrics"`
			Published string `json:"published"`
		} `json:"cve"`
	} `json:"vulnerabilities"`
}

// SearchByCPE interroga l'endpoint NVD cpeName per il CPE 2.3 specificato
func (c *NVDOnlineClient) SearchByCPE(ctx context.Context, cpe23URI string) ([]*domain.CVEEntry, error) {
	cpeClean := strings.TrimSpace(cpe23URI)
	if cpeClean == "" {
		return nil, nil
	}

	reqURL := fmt.Sprintf("%s?cpeName=%s&resultsPerPage=10", c.baseURL, url.QueryEscape(cpeClean))
	return c.executeRequest(ctx, reqURL, cpeClean)
}

// SearchByKeyword cerca le vulnerabilità relative a un nome di prodotto e versione
func (c *NVDOnlineClient) SearchByKeyword(ctx context.Context, keyword string) ([]*domain.CVEEntry, error) {
	kwClean := strings.TrimSpace(keyword)
	if kwClean == "" {
		return nil, nil
	}

	reqURL := fmt.Sprintf("%s?keywordSearch=%s&resultsPerPage=10", c.baseURL, url.QueryEscape(kwClean))
	return c.executeRequest(ctx, reqURL, "")
}

func (c *NVDOnlineClient) executeRequest(ctx context.Context, reqURL, cpeURI string) ([]*domain.CVEEntry, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "NextGen-CyberSec-Scanner/2.0")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("risposta NVD non valida (status %d)", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024)) // Limite 1MB per sicurezza
	if err != nil {
		return nil, err
	}

	var nvdResp nvdResponse
	if err := json.Unmarshal(body, &nvdResp); err != nil {
		return nil, err
	}

	return parseNVDItems(&nvdResp, cpeURI), nil
}

func parseNVDItems(resp *nvdResponse, cpeURI string) []*domain.CVEEntry {
	if resp == nil || len(resp.Vulnerabilities) == 0 {
		return nil
	}

	var entries []*domain.CVEEntry
	now := time.Now().UTC()

	for _, v := range resp.Vulnerabilities {
		cve := v.CVE
		if cve.ID == "" {
			continue
		}

		desc := ""
		for _, d := range cve.Descriptions {
			if d.Lang == "en" {
				desc = d.Value
				break
			}
		}
		if desc == "" && len(cve.Descriptions) > 0 {
			desc = cve.Descriptions[0].Value
		}

		var cvssScore *float64
		var severity domain.SeverityLevel = domain.SeverityMedium

		if len(cve.Metrics.CvssMetricV31) > 0 {
			score := cve.Metrics.CvssMetricV31[0].CvssData.BaseScore
			cvssScore = &score
			severity = parseSeverityStr(cve.Metrics.CvssMetricV31[0].CvssData.BaseSeverity, score)
		} else if len(cve.Metrics.CvssMetricV30) > 0 {
			score := cve.Metrics.CvssMetricV30[0].CvssData.BaseScore
			cvssScore = &score
			severity = parseSeverityStr(cve.Metrics.CvssMetricV30[0].CvssData.BaseSeverity, score)
		} else if len(cve.Metrics.CvssMetricV2) > 0 {
			score := cve.Metrics.CvssMetricV2[0].CvssData.BaseScore
			cvssScore = &score
			severity = parseSeverityStr(cve.Metrics.CvssMetricV2[0].BaseSeverity, score)
		}

		pubDate, err := time.Parse(time.RFC3339, cve.Published)
		if err != nil {
			pubDate = now
		}

		matchCriteria := []domain.CPEMatchCriterion{}
		if cpeURI != "" {
			matchCriteria = append(matchCriteria, domain.CPEMatchCriterion{
				CPE23URI:   cpeURI,
				Vulnerable: true,
			})
		}

		entry := &domain.CVEEntry{
			CVEID:            cve.ID,
			CVSSScore:        cvssScore,
			Severity:         severity,
			Description:      desc,
			CPEMatchCriteria: matchCriteria,
			PublishedDate:    &pubDate,
			UpdatedAt:        now,
		}
		entries = append(entries, entry)
	}

	return entries
}

func parseSeverityStr(s string, score float64) domain.SeverityLevel {
	switch strings.ToUpper(s) {
	case "CRITICAL":
		return domain.SeverityCritical
	case "HIGH":
		return domain.SeverityHigh
	case "MEDIUM":
		return domain.SeverityMedium
	case "LOW":
		return domain.SeverityLow
	default:
		return domain.SeverityFromCVSS(score)
	}
}

