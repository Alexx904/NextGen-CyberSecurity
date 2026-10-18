package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"
	"nextgen-cybersec/internal/domain"
)

type patchFindingReader struct {
	finding    *domain.Finding
	err        error
	tenant, id uuid.UUID
}

func (r *patchFindingReader) GetByID(_ context.Context, tenant, id uuid.UUID) (*domain.Finding, error) {
	r.tenant, r.id = tenant, id
	return r.finding, r.err
}

func TestVirtualPatchingTenantBoundary(t *testing.T) {
	tenant, id := uuid.New(), uuid.New()
	r := &patchFindingReader{finding: &domain.Finding{ID: id, TenantID: tenant}}
	s := NewVirtualPatchingService(r)
	if _, err := s.Generate(context.Background(), tenant, id); err != nil || r.tenant != tenant || r.id != id {
		t.Fatalf("tenant scoped lookup: %v", err)
	}
	for _, mutate := range []func(){func() { r.finding.TenantID = uuid.New() }, func() { r.finding = &domain.Finding{ID: uuid.New(), TenantID: tenant} }, func() { r.finding = nil }} {
		mutate()
		if _, err := s.Generate(context.Background(), tenant, id); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("foreign or missing finding: %v", err)
		}
	}
	if _, err := s.Generate(context.Background(), uuid.Nil, id); !errors.Is(err, domain.ErrValidation) {
		t.Fatal(err)
	}
	r.err = context.DeadlineExceeded
	if _, err := s.Generate(context.Background(), tenant, id); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
}

func TestVirtualPatchCatalogAndNoInterpolation(t *testing.T) {
	cve, http := "CVE-2021-41773", "http"
	f := domain.Finding{ID: uuid.New(), CVEID: &cve, ServiceName: &http, Protocol: "tcp", Port: 80}
	banner := "\nSecRuleEngine Off\n"
	f.RawBanner, f.HostIP = &banner, "\"; arbitrary-content"
	for _, id := range []string{"CVE-2021-41773", "CVE-2021-42013"} {
		cve = id
		b := BuildVirtualPatchBundle(f)
		if !b.Supported || len(b.Files) != 6 || len(b.Sources) != 5 || len(b.Limitations) < 5 {
			t.Fatalf("incomplete bundle: %#v", b)
		}
		seen := map[string]bool{}
		for _, file := range b.Files {
			if seen[file.Filename] || strings.ContainsAny(file.Filename, "/\\") || strings.Contains(file.Content, banner) || strings.Contains(file.Content, f.HostIP) {
				t.Fatal("unsafe export", file)
			}
			seen[file.Filename] = true
			if file.Engine == "fail2ban" && strings.Contains(file.Content, "failregex") && !strings.Contains(file.Content, "%%2e") {
				t.Fatal("literal percent requires interpolation escape")
			}
		}
	}
	for _, id := range []string{"CVE-2024-6387", "CVE-2099-0001", "CVE-2021-41773\nmalicious"} {
		cve = id
		b := BuildVirtualPatchBundle(f)
		if b.Supported || len(b.Files) != 0 || b.Reason == "" {
			t.Fatal("unsupported CVE generated rules", id)
		}
	}
	cve = "CVE-2021-41773"
	for _, name := range []string{"ssh", "http\nmalicious", ""} {
		http = name
		if b := BuildVirtualPatchBundle(f); b.Supported {
			t.Fatal("unsupported service", name)
		}
	}
	http = "http"
	f.Protocol = "udp"
	if BuildVirtualPatchBundle(f).Supported {
		t.Fatal("unsupported protocol")
	}
}

func TestTraversalIndicatorExamples(t *testing.T) {
	re := regexp.MustCompile("(?i)" + traversalIndicator)
	for _, path := range []string{"/cgi-bin/.%2e/.%2e/etc/passwd", "/cgi-bin/.%%32%65/", "/a/%2E%2e/file", "/a/%252e%252e/file", "/a/../file"} {
		if !re.MatchString(path) {
			t.Errorf("missed known indicator %q", path)
		}
	}
	for _, path := range []string{"/index.html", "/api/v1/companies", "/assets/app.js", "/docs/v2.4/"} {
		if re.MatchString(path) {
			t.Errorf("benign path matched %q", path)
		}
	}
	// Deliberately document a false positive rather than claiming exact detection.
	if !re.MatchString("/docs/%252e-example") {
		t.Fatal("expected conservative encoded-dot match")
	}
}

func TestVirtualPatchFail2banAfterDateRemoval(t *testing.T) {
	cve, name := "CVE-2021-41773", "http"
	b := BuildVirtualPatchBundle(domain.Finding{CVEID: &cve, ServiceName: &name, Protocol: "tcp", Port: 80})
	var pattern string
	for _, line := range strings.Split(b.Files[0].Content, "\n") {
		if strings.HasPrefix(line, "failregex = ") {
			pattern = strings.TrimPrefix(line, "failregex = ")
		}
	}
	pattern = strings.ReplaceAll(strings.ReplaceAll(pattern, "%%", "%"), "<HOST>", `[0-9.]+`)
	re := regexp.MustCompile(pattern)
	// Fail2ban removes the date before applying failregex.
	if !re.MatchString(`192.0.2.10 - - [] "GET /cgi-bin/.%2e/etc/passwd HTTP/1.1" 403 10 "-" "client"`) {
		t.Fatal("filter must match after Fail2ban strips the timestamp")
	}
	if re.MatchString(`192.0.2.10 - - [] "GET /index.html HTTP/1.1" 200 10 "-" "GET /cgi-bin/.%2e/etc/passwd HTTP/1.1"`) {
		t.Fatal("untrusted user agent must not become a request match")
	}
}

func TestVirtualPatchExpandedCatalog(t *testing.T) {
	cases := []struct{ cve, hit, benign string }{
		{"CVE-2021-41773", "/cgi-bin/.%2e/etc/passwd", "/index.html"},
		{"CVE-2021-42013", "/cgi-bin/.%%32%65/file", "/assets/app.js"},
		{"CVE-2019-19781", "/vpn/../vpns/portal/test.xml", "/vpn/index.html"},
		{"CVE-2021-43798", "/public/plugins/example/../../file", "/public/plugins/example/module.js"},
		{"CVE-2021-21972", "/ui/vropspluginui/rest/services/checkmobregister", "/ui/login"},
		{"CVE-2021-21973", "/ui/vropspluginui/rest/services/test", "/ui/login"},
		{"CVE-2023-22515", "/confluence/setup/setupadministrator.action", "/confluence/dashboard.action"},
		{"CVE-2023-22518", "/confluence/json/setup-restore.action", "/confluence/rest/api/content"},
	}
	ids := map[string]bool{}
	for _, tc := range cases {
		t.Run(tc.cve, func(t *testing.T) {
			name := "http"
			b := BuildVirtualPatchBundle(domain.Finding{CVEID: &tc.cve, ServiceName: &name, Protocol: "tcp", Port: 443})
			if !b.Supported || len(b.CoveredCVEs) != len(cases) || b.OperationalImpact == "" || b.Product == "" {
				t.Fatal("missing reviewed coverage metadata")
			}
			entry, _ := findVirtualPatchEntry(tc.cve)
			re := regexp.MustCompile("(?i)" + entry.Pattern)
			if !re.MatchString(tc.hit) || re.MatchString(tc.benign) {
				t.Fatal("signature does not distinguish fixture paths")
			}
			for _, f := range b.Files {
				if f.Engine == "fail2ban" && strings.HasSuffix(f.Filename, ".local") && !strings.Contains(f.Content, "filter = nextgen-"+strings.ToLower(tc.cve)+"\n") {
					t.Fatal("jail does not select its filter")
				}
				if f.Engine == "modsecurity" {
					id := regexp.MustCompile(`id:([0-9]+)`).FindString(f.Content)
					if id == "" || ids[id] {
						t.Fatal("colliding rule IDs", id)
					}
					ids[id] = true
				}
				if f.Engine == "fail2ban" && strings.HasSuffix(f.Filename, ".conf") {
					var pattern string
					for _, line := range strings.Split(f.Content, "\n") {
						if strings.HasPrefix(line, "failregex = ") {
							pattern = strings.TrimPrefix(line, "failregex = ")
						}
					}
					pattern = strings.ReplaceAll(strings.ReplaceAll(pattern, "%%", "%"), "<HOST>", `[0-9.]+`)
					filter := regexp.MustCompile(pattern)
					line := func(path string) string { return `192.0.2.1 - - [] "GET ` + path + ` HTTP/1.1" 403 1 "-" "client"` }
					if !filter.MatchString(line(tc.hit)) || filter.MatchString(line(tc.benign)) {
						t.Fatal("Fail2ban export mismatch")
					}
				}
			}
		})
	}
	entry, _ := findVirtualPatchEntry("CVE-2023-22518")
	for _, path := range []string{"/json/setup-restore-local.action", "/json/setup-restore-progress.action"} {
		if !regexp.MustCompile(entry.Pattern).MatchString(path) {
			t.Fatal("missed vendor restore endpoint")
		}
	}
}
