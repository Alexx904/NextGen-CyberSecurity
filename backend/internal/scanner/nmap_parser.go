package scanner

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"nextgen-cybersec/internal/domain"
)

// ParsedService rappresenta le informazioni su un servizio di rete identificato da Nmap
type ParsedService struct {
	Name        string   `json:"name"`
	Product     string   `json:"product,omitempty"`
	Version     string   `json:"version,omitempty"`
	ExtraInfo   string   `json:"extra_info,omitempty"`
	OSType      string   `json:"os_type,omitempty"`
	Method      string   `json:"method,omitempty"`
	Confidence  int      `json:"confidence"`
	CPEs        []string `json:"cpes,omitempty"`
	RawBanner   string   `json:"raw_banner,omitempty"`
}

// ParsedPort rappresenta una singola porta analizzata
type ParsedPort struct {
	PortID   int            `json:"port_id"`
	Protocol string         `json:"protocol"`
	State    string         `json:"state"` // open, closed, filtered, unfiltered
	Reason   string         `json:"reason,omitempty"`
	Service  *ParsedService `json:"service,omitempty"`
}

// ParsedOS racchiude il risultato del fingerprinting del sistema operativo
type ParsedOS struct {
	Name     string `json:"name"`
	Accuracy int    `json:"accuracy"`
}

// ParsedHost rappresenta un singolo host bersaglio rilevato nella scansione
type ParsedHost struct {
	IP              string       `json:"ip"`
	IsIPv6          bool         `json:"is_ipv6"`
	MAC             string       `json:"mac,omitempty"`
	Vendor          string       `json:"vendor,omitempty"`
	Hostnames       []string     `json:"hostnames,omitempty"`
	PrimaryHostname string       `json:"primary_hostname,omitempty"`
	Status          string       `json:"status"` // up, down
	StatusReason    string       `json:"status_reason,omitempty"`
	LatencyMs       float64      `json:"latency_ms,omitempty"`
	OS              *ParsedOS    `json:"os,omitempty"`
	Ports           []ParsedPort `json:"ports,omitempty"`
}

// ScanRunStats racchiude le metriche generali di completamento della scansione
type ScanRunStats struct {
	Scanner    string  `json:"scanner"`
	Version    string  `json:"version"`
	TotalHosts int     `json:"total_hosts"`
	UpHosts    int     `json:"up_hosts"`
	DownHosts  int     `json:"down_hosts"`
	Duration   float64 `json:"duration_seconds"`
}

// ParsedScanReport racchiude l'intero report strutturato ricavato dal parsing XML
type ParsedScanReport struct {
	Scanner    string       `json:"scanner"`
	Args       string       `json:"args"`
	Version    string       `json:"version"`
	StartTime  time.Time    `json:"start_time"`
	EndTime    time.Time    `json:"end_time"`
	Duration   float64      `json:"duration_seconds"`
	TotalHosts int          `json:"total_hosts"`
	UpHosts    int          `json:"up_hosts"`
	DownHosts  int          `json:"down_hosts"`
	Hosts      []ParsedHost `json:"hosts"`
}

// GetUpHosts estrae e restituisce esclusivamente gli host aventi stato 'up' (esclude host down/offline)
func (r *ParsedScanReport) GetUpHosts() []ParsedHost {
	var alive []ParsedHost
	for _, h := range r.Hosts {
		if strings.EqualFold(h.Status, "up") {
			alive = append(alive, h)
		}
	}
	return alive
}

// ToSummaryStats converte le metriche del report nella struct di dominio ScanSummaryStats
func (r *ParsedScanReport) ToSummaryStats() domain.ScanSummaryStats {
	openPorts := r.ExtractOpenPorts()

	upHosts := r.UpHosts
	downHosts := r.DownHosts
	if upHosts == 0 && downHosts == 0 {
		for _, h := range r.Hosts {
			if strings.EqualFold(h.Status, "up") {
				upHosts++
			} else if strings.EqualFold(h.Status, "down") {
				downHosts++
			}
		}
	}

	totalHosts := r.TotalHosts
	if totalHosts == 0 {
		totalHosts = len(r.Hosts)
	}

	return domain.ScanSummaryStats{
		TotalHosts:      totalHosts,
		UpHosts:         upHosts,
		DownHosts:       downHosts,
		TotalOpenPorts:  openPorts,
		DurationSeconds: r.Duration,
	}
}

// ExtractOpenPorts calcola il numero complessivo di porte in stato 'open' su tutti gli host attivi (up)
func (r *ParsedScanReport) ExtractOpenPorts() int {
	count := 0
	for _, h := range r.GetUpHosts() {
		for _, p := range h.Ports {
			if strings.EqualFold(p.State, "open") {
				count++
			}
		}
	}
	return count
}

// ExtractCPEs estrae una lista deduplicata di tutti i CPE identificati nei servizi degli host attivi (up)
func (r *ParsedScanReport) ExtractCPEs() []string {
	seen := make(map[string]bool)
	var cpes []string

	for _, h := range r.GetUpHosts() {
		for _, p := range h.Ports {
			if p.Service != nil {
				for _, cpe := range p.Service.CPEs {
					trimmed := strings.TrimSpace(cpe)
					if trimmed != "" && !seen[trimmed] {
						seen[trimmed] = true
						cpes = append(cpes, trimmed)
					}
				}
			}
		}
	}

	return cpes
}

// ============================================================================
// MODELLI XML PRIVATI PER LO STREAMING DECODER (DTD NMAP)
// ============================================================================

type xmlAddress struct {
	Addr     string `xml:"addr,attr"`
	AddrType string `xml:"addrtype,attr"`
	Vendor   string `xml:"vendor,attr"`
}

type xmlHostname struct {
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
}

type xmlStatus struct {
	State  string `xml:"state,attr"`
	Reason string `xml:"reason,attr"`
}

type xmlService struct {
	Name      string   `xml:"name,attr"`
	Product   string   `xml:"product,attr"`
	Version   string   `xml:"version,attr"`
	ExtraInfo string   `xml:"extrainfo,attr"`
	OSType    string   `xml:"ostype,attr"`
	Method    string   `xml:"method,attr"`
	Conf      int      `xml:"conf,attr"`
	CPEs      []string `xml:"cpe"`
}

type xmlState struct {
	State  string `xml:"state,attr"`
	Reason string `xml:"reason,attr"`
}

type xmlPort struct {
	Protocol string     `xml:"protocol,attr"`
	PortID   int        `xml:"portid,attr"`
	State    xmlState   `xml:"state"`
	Service  xmlService `xml:"service"`
}

type xmlOSMatch struct {
	Name     string `xml:"name,attr"`
	Accuracy int    `xml:"accuracy,attr"`
}

type xmlTimes struct {
	SRTT   string `xml:"srtt,attr"`
	RTTVAR string `xml:"rttvar,attr"`
	TO     string `xml:"to,attr"`
}

type xmlHost struct {
	Status    xmlStatus     `xml:"status"`
	Addresses []xmlAddress  `xml:"address"`
	Hostnames []xmlHostname `xml:"hostnames>hostname"`
	Ports     []xmlPort     `xml:"ports>port"`
	OSMatches []xmlOSMatch  `xml:"os>osmatch"`
	Times     xmlTimes      `xml:"times"`
}

type xmlFinished struct {
	Time    int64   `xml:"time,attr"`
	Elapsed float64 `xml:"elapsed,attr"`
	Summary string  `xml:"summary,attr"`
	Exit    string  `xml:"exit,attr"`
}

type xmlHostsStats struct {
	Up    int `xml:"up,attr"`
	Down  int `xml:"down,attr"`
	Total int `xml:"total,attr"`
}

type xmlRunStats struct {
	Finished xmlFinished   `xml:"finished"`
	Hosts    xmlHostsStats `xml:"hosts"`
}

// buildRawBanner compone il banner descrittivo del servizio unendo Product, Version ed ExtraInfo
func buildRawBanner(s *xmlService) string {
	var parts []string
	if s.Product != "" {
		parts = append(parts, s.Product)
	}
	if s.Version != "" {
		parts = append(parts, s.Version)
	}
	if s.ExtraInfo != "" {
		parts = append(parts, s.ExtraInfo)
	}
	if len(parts) == 0 {
		return s.Name
	}
	return strings.Join(parts, " ")
}

// convertXMLHost converte il nodo XML interno nella struct pubblica ParsedHost
func convertXMLHost(xh *xmlHost) ParsedHost {
	host := ParsedHost{
		Status:       xh.Status.State,
		StatusReason: xh.Status.Reason,
		Hostnames:    make([]string, 0, len(xh.Hostnames)),
		Ports:        make([]ParsedPort, 0, len(xh.Ports)),
	}

	// Elaborazione indirizzi IP e MAC
	for _, addr := range xh.Addresses {
		switch addr.AddrType {
		case "ipv4":
			if host.IP == "" {
				host.IP = addr.Addr
				host.IsIPv6 = false
			}
		case "ipv6":
			if host.IP == "" || host.IsIPv6 {
				host.IP = addr.Addr
				host.IsIPv6 = true
			}
		case "mac":
			host.MAC = addr.Addr
			if addr.Vendor != "" {
				host.Vendor = addr.Vendor
			}
		}
	}

	// Elaborazione smoothed round-trip time (latenza RTT in millisecondi)
	if xh.Times.SRTT != "" {
		if srttUsec, err := strconv.ParseFloat(xh.Times.SRTT, 64); err == nil && srttUsec > 0 {
			host.LatencyMs = math.Round((srttUsec/1000.0)*100) / 100
		}
	}

	// Elaborazione hostname
	for _, hn := range xh.Hostnames {
		if hn.Name != "" {
			host.Hostnames = append(host.Hostnames, hn.Name)
		}
	}
	if len(host.Hostnames) > 0 {
		host.PrimaryHostname = host.Hostnames[0]
	}

	// Elaborazione OS fingerprint
	if len(xh.OSMatches) > 0 {
		bestMatch := xh.OSMatches[0]
		for _, m := range xh.OSMatches {
			if m.Accuracy > bestMatch.Accuracy {
				bestMatch = m
			}
		}
		host.OS = &ParsedOS{
			Name:     bestMatch.Name,
			Accuracy: bestMatch.Accuracy,
		}
	}

	// Elaborazione porte e servizi
	for _, p := range xh.Ports {
		parsedPort := ParsedPort{
			PortID:   p.PortID,
			Protocol: p.Protocol,
			State:    p.State.State,
			Reason:   p.State.Reason,
		}

		if p.Service.Name != "" || p.Service.Product != "" {
			parsedPort.Service = &ParsedService{
				Name:        p.Service.Name,
				Product:     p.Service.Product,
				Version:     p.Service.Version,
				ExtraInfo:   p.Service.ExtraInfo,
				OSType:      p.Service.OSType,
				Method:      p.Service.Method,
				Confidence:  p.Service.Conf,
				CPEs:        p.Service.CPEs,
				RawBanner:   buildRawBanner(&p.Service),
			}
		}

		host.Ports = append(host.Ports, parsedPort)
	}

	return host
}

// ============================================================================
// API PUBBLICA DI PARSING STREAMING (NFR-PERF-02)
// ============================================================================

// ParseNmapXML decodifica uno stream XML prodotto da Nmap in modalità streaming
func ParseNmapXML(r io.Reader) (*ParsedScanReport, error) {
	if r == nil {
		return nil, fmt.Errorf("%w: input reader nullo", domain.ErrValidation)
	}

	decoder := xml.NewDecoder(r)
	report := &ParsedScanReport{
		Hosts: make([]ParsedHost, 0),
	}

	hasRoot := false

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("errore parsing XML Nmap: %w", err)
		}

		switch se := tok.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "nmaprun":
				hasRoot = true
				for _, attr := range se.Attr {
					switch attr.Name.Local {
					case "scanner":
						report.Scanner = attr.Value
					case "args":
						report.Args = attr.Value
					case "version":
						report.Version = attr.Value
					case "start":
						if s, parseErr := strconv.ParseInt(attr.Value, 10, 64); parseErr == nil {
							report.StartTime = time.Unix(s, 0)
						}
					}
				}
			case "host":
				var xh xmlHost
				if err := decoder.DecodeElement(&xh, &se); err != nil {
					return nil, fmt.Errorf("errore decodifica elemento host: %w", err)
				}
				report.Hosts = append(report.Hosts, convertXMLHost(&xh))
			case "runstats":
				var xrs xmlRunStats
				if err := decoder.DecodeElement(&xrs, &se); err != nil {
					return nil, fmt.Errorf("errore decodifica runstats: %w", err)
				}
				report.TotalHosts = xrs.Hosts.Total
				report.UpHosts = xrs.Hosts.Up
				report.DownHosts = xrs.Hosts.Down
				report.Duration = xrs.Finished.Elapsed
				if xrs.Finished.Time > 0 {
					report.EndTime = time.Unix(xrs.Finished.Time, 0)
				}
			}
		}
	}

	if !hasRoot {
		return nil, fmt.Errorf("%w: documento XML non conforme (radice <nmaprun> assente)", domain.ErrValidation)
	}

	return report, nil
}

// ParseNmapXMLBytes decodifica direttamente una sequenza di byte XML
func ParseNmapXMLBytes(data []byte) (*ParsedScanReport, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("%w: payload XML vuoto", domain.ErrValidation)
	}
	return ParseNmapXML(bytes.NewReader(data))
}

// HostCallback definisce la funzione handler per elaborare in streaming ogni singolo host
type HostCallback func(host ParsedHost) error

// StreamNmapXML elabora lo stream XML in tempo reale, invocando callback per ciascun host
func StreamNmapXML(r io.Reader, callback HostCallback) (*ScanRunStats, error) {
	if r == nil {
		return nil, fmt.Errorf("%w: input reader nullo", domain.ErrValidation)
	}

	decoder := xml.NewDecoder(r)
	stats := &ScanRunStats{}
	hasRoot := false

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("errore streaming XML Nmap: %w", err)
		}

		switch se := tok.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "nmaprun":
				hasRoot = true
				for _, attr := range se.Attr {
					switch attr.Name.Local {
					case "scanner":
						stats.Scanner = attr.Value
					case "version":
						stats.Version = attr.Value
					}
				}
			case "host":
				var xh xmlHost
				if err := decoder.DecodeElement(&xh, &se); err != nil {
					return nil, fmt.Errorf("errore decodifica elemento host: %w", err)
				}
				parsedHost := convertXMLHost(&xh)
				if callback != nil {
					if err := callback(parsedHost); err != nil {
						return nil, fmt.Errorf("callback host ha restituito errore: %w", err)
					}
				}
			case "runstats":
				var xrs xmlRunStats
				if err := decoder.DecodeElement(&xrs, &se); err != nil {
					return nil, fmt.Errorf("errore decodifica runstats: %w", err)
				}
				stats.TotalHosts = xrs.Hosts.Total
				stats.UpHosts = xrs.Hosts.Up
				stats.DownHosts = xrs.Hosts.Down
				stats.Duration = xrs.Finished.Elapsed
			}
		}
	}

	if !hasRoot {
		return nil, fmt.Errorf("%w: documento XML non conforme (radice <nmaprun> assente)", domain.ErrValidation)
	}

	return stats, nil
}

