//go:build integration

package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"nextgen-cybersec/internal/domain"
)

func TestVirtualPatchNginxSyntax(t *testing.T) {
	const container = "nextgen-e28-ui-qa-frontend"
	if os.Getenv("TEST_VIRTUALPATCH_NGINX_CONTAINER") != container {
		t.Fatal("explicit dedicated E28 QA frontend container required")
	}
	config := "events {}\nhttp { access_log off;\n"
	for i, entry := range virtualPatchCatalog {
		service := "http"
		bundle := BuildVirtualPatchBundle(domain.Finding{CVEID: &entry.CVE, ServiceName: &service, Protocol: "tcp", Port: 443})
		for _, file := range bundle.Files {
			if file.Engine == "nginx" {
				config += fmt.Sprintf("server { listen 127.0.0.1:%d;\n%s\nlocation / { return 200; }\n}\n", 18100+i, file.Content)
			}
		}
	}
	config += "}\n"
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", "exec", "-i", container, "sh", "-c", "cat > /tmp/e28-virtualpatch-catalog-check.conf && nginx -t -c /tmp/e28-virtualpatch-catalog-check.conf")
	cmd.Stdin = strings.NewReader(config)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("nginx syntax failed: %v\n%s", err, output)
	}
	t.Logf("nginx syntax PASS for %d catalog entries: %s", len(virtualPatchCatalog), output)
}
