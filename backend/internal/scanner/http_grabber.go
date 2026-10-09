package scanner

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var titleRegex = regexp.MustCompile(`(?i)<title[^>]*>(.*?)</title>`)

// GrabHTTPInfo tenta di connettersi alle porte web aperte di un host per estrarne il titolo HTML e l'header Server
func GrabHTTPInfo(ctx context.Context, ip string, openPorts []int) (title string, serverHeader string) {
	if ip == "" || len(openPorts) == 0 {
		return "", ""
	}

	// Identifica porte web candidate
	webPorts := make([]int, 0, 4)
	for _, p := range openPorts {
		switch p {
		case 80, 443, 8080, 8443, 3000, 5000, 8000, 8081, 8888, 9090:
			webPorts = append(webPorts, p)
		}
	}

	if len(webPorts) == 0 {
		if len(openPorts) <= 3 {
			webPorts = openPorts
		} else {
			return "", ""
		}
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Ignora certificati self-signed per apparati LAN
		MaxIdleConns:    10,
		IdleConnTimeout: 1 * time.Second,
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   600 * time.Millisecond, // Timeout rapido per non rallentare la scansione
	}

	for _, port := range webPorts {
		schemes := []string{"http"}
		if port == 443 || port == 8443 {
			schemes = []string{"https", "http"}
		} else if port == 80 {
			schemes = []string{"http", "https"}
		}

		for _, scheme := range schemes {
			url := fmt.Sprintf("%s://%s:%d/", scheme, ip, port)
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				continue
			}
			req.Header.Set("User-Agent", "Mozilla/5.0 (LAN-Discovery)")

			resp, err := client.Do(req)
			if err != nil {
				continue
			}

			if serverHeader == "" {
				serverHeader = resp.Header.Get("Server")
			}

			// Leggi i primi 4KB per individuare il tag <title>
			bodyReader := io.LimitReader(resp.Body, 4096)
			bodyBytes, _ := io.ReadAll(bodyReader)
			resp.Body.Close()

			matches := titleRegex.FindSubmatch(bodyBytes)
			if len(matches) > 1 {
				extracted := strings.TrimSpace(string(matches[1]))
				// Pulisci entità HTML basilari o newline
				extracted = strings.ReplaceAll(extracted, "\n", " ")
				extracted = strings.ReplaceAll(extracted, "\r", "")
				extracted = strings.Join(strings.Fields(extracted), " ")
				if extracted != "" && !strings.EqualFold(extracted, "404 not found") && !strings.EqualFold(extracted, "error") {
					return extracted, serverHeader
				}
			}

			if serverHeader != "" {
				return "", serverHeader
			}
		}
	}

	return "", serverHeader
}

