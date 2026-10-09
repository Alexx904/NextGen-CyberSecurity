package scanner

import (
	"bufio"
	"encoding/json"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

// ARPCache definisce l'interfaccia per la risoluzione e memorizzazione della corrispondenza IP -> MAC
type ARPCache interface {
	Lookup(ip string) string
	Set(ip, mac string)
	LoadFromFile(filePath string) error
	LoadFromRaw(raw string)
	Refresh() error
	Entries() map[string]string
}

type memoryARPCache struct {
	mu      sync.RWMutex
	entries map[string]string
}

// NewARPCache crea una nuova istanza di ARPCache e tenta un caricamento automatico dalle sorgenti disponibili
func NewARPCache() ARPCache {
	c := &memoryARPCache{
		entries: make(map[string]string),
	}
	_ = c.Refresh()
	return c
}

func (c *memoryARPCache) Lookup(ip string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.entries[strings.TrimSpace(ip)]
}

func (c *memoryARPCache) Set(ip, mac string) {
	normMAC := NormalizeMAC(mac)
	if normMAC == "" {
		return
	}
	cleanIP := strings.TrimSpace(ip)
	if parsed := net.ParseIP(cleanIP); parsed == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[cleanIP] = normMAC
}

func (c *memoryARPCache) Entries() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	res := make(map[string]string, len(c.entries))
	for k, v := range c.entries {
		res[k] = v
	}
	return res
}

func (c *memoryARPCache) LoadFromFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	c.LoadFromRaw(string(data))
	return nil
}

// LoadFromRaw analizza una stringa grezza contenente output di tabelle ARP (Linux /proc/net/arp, Windows arp -a, o JSON)
func (c *memoryARPCache) LoadFromRaw(raw string) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return
	}

	// Tentativo parsing JSON {"ip": "mac"}
	if strings.HasPrefix(trimmed, "{") {
		var jsonMap map[string]string
		if err := json.Unmarshal([]byte(trimmed), &jsonMap); err == nil {
			for ip, mac := range jsonMap {
				c.Set(ip, mac)
			}
			return
		}
	}

	// Parsing riga per riga
	scanner := bufio.NewScanner(strings.NewReader(raw))
	macRegex := regexp.MustCompile(`([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})`)
	ipRegex := regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)

	for scanner.Scan() {
		line := scanner.Text()
		ipMatch := ipRegex.FindString(line)
		macMatch := macRegex.FindString(line)

		if ipMatch != "" && macMatch != "" {
			c.Set(ipMatch, macMatch)
		}
	}
}

// Refresh tenta di ricaricare la tabella ARP dalle sorgenti di sistema
func (c *memoryARPCache) Refresh() error {
	// 1. Controlla file di cache montato (es. in container Docker /app/data/arp_cache.txt)
	candidates := []string{
		"/app/data/arp_cache.txt",
		"./data/arp_cache.txt",
		"./backend/data/arp_cache.txt",
		"/proc/net/arp",
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			_ = c.LoadFromFile(path)
		}
	}

	// 2. Se in ambiente Windows nativo, esegui comando arp -a
	if runtime.GOOS == "windows" {
		cmd := exec.Command("arp", "-a")
		out, err := cmd.Output()
		if err == nil {
			c.LoadFromRaw(string(out))
		}
	}

	return nil
}

// NormalizeMAC normalizza un indirizzo MAC in formato uppercase "XX:XX:XX:XX:XX:XX"
func NormalizeMAC(mac string) string {
	clean := strings.TrimSpace(mac)
	if clean == "" {
		return ""
	}

	clean = strings.ReplaceAll(clean, "-", ":")
	parts := strings.Split(clean, ":")
	if len(parts) != 6 {
		return ""
	}

	var sb strings.Builder
	for i, part := range parts {
		if len(part) == 1 {
			part = "0" + part
		}
		if len(part) != 2 {
			return ""
		}
		sb.WriteString(strings.ToUpper(part))
		if i < 5 {
			sb.WriteString(":")
		}
	}

	return sb.String()
}

