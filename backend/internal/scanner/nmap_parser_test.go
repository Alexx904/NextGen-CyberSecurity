package scanner

import (
	"errors"
	"strings"
	"testing"

	"nextgen-cybersec/internal/domain"
)

// Fixture XML realistica con molteplici host, sistemi operativi differenti, CPE e porte miste
const realisticNmapXML = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE nmaprun>
<?xml-stylesheet href="file:///usr/bin/../share/nmap/nmap.xsl" type="text/xsl"?>
<nmaprun scanner="nmap" args="nmap -sV -O -oX - 192.168.1.0/24" start="1700000000" startstr="Wed Oct  8 07:30:00 2026" version="7.94" xmloutputversion="1.05">
<scaninfo type="syn" protocol="tcp" numservices="1000" services="1-1000"/>
<verbose level="0"/>
<debugging level="0"/>
<host starttime="1700000001" endtime="1700000010">
<status state="up" reason="arp-response" reason_ttl="0"/>
<address addr="192.168.1.50" addrtype="ipv4"/>
<address addr="00:0C:29:84:7B:A2" addrtype="mac" vendor="VMware"/>
<hostnames>
<hostname name="webserver.internal" type="PTR"/>
<hostname name="app.local" type="user"/>
</hostnames>
<ports>
<port protocol="tcp" portid="22">
<state state="open" reason="syn-ack" reason_ttl="64"/>
<service name="ssh" product="OpenSSH" version="8.9p1" extrainfo="Ubuntu Linux" ostype="Linux" method="probed" conf="10">
<cpe>cpe:/a:openbsd:openssh:8.9p1</cpe>
<cpe>cpe:/o:canonical:ubuntu_linux</cpe>
</service>
</port>
<port protocol="tcp" portid="80">
<state state="open" reason="syn-ack" reason_ttl="64"/>
<service name="http" product="Apache httpd" version="2.4.52" extrainfo="(Ubuntu)" method="probed" conf="10">
<cpe>cpe:/a:apache:http_server:2.4.52</cpe>
</service>
</port>
<port protocol="tcp" portid="443">
<state state="open" reason="syn-ack" reason_ttl="64"/>
<service name="ssl/http" product="Apache httpd" version="2.4.52" extrainfo="(Ubuntu)" method="probed" conf="10">
<cpe>cpe:/a:apache:http_server:2.4.52</cpe>
</service>
</port>
<port protocol="tcp" portid="8080">
<state state="closed" reason="reset" reason_ttl="64"/>
</port>
</ports>
<os>
<osmatch name="Linux 5.4 - 5.15" accuracy="96"/>
<osmatch name="Linux 4.15" accuracy="85"/>
</os>
</host>
<host starttime="1700000011" endtime="1700000020">
<status state="up" reason="syn-ack" reason_ttl="128"/>
<address addr="192.168.1.100" addrtype="ipv4"/>
<hostnames>
<hostname name="dc01.corp.local" type="PTR"/>
</hostnames>
<ports>
<port protocol="tcp" portid="445">
<state state="open" reason="syn-ack" reason_ttl="128"/>
<service name="microsoft-ds" product="Windows Server 2019" extrainfo="workgroup: WORKGROUP" ostype="Windows" method="probed" conf="10">
<cpe>cpe:/o:microsoft:windows_server_2019</cpe>
</service>
</port>
<port protocol="tcp" portid="3389">
<state state="open" reason="syn-ack" reason_ttl="128"/>
<service name="ms-wbt-server" product="Microsoft Terminal Services" method="probed" conf="10"/>
</port>
</ports>
<os>
<osmatch name="Microsoft Windows Server 2019" accuracy="98"/>
</os>
</host>
<host starttime="1700000021" endtime="1700000022">
<status state="down" reason="no-response" reason_ttl="0"/>
<address addr="192.168.1.200" addrtype="ipv4"/>
</host>
<runstats>
<finished time="1700000025" timestr="Wed Oct  8 07:30:25 2026" summary="Nmap done at Wed Oct  8 07:30:25 2026; 3 IP addresses (2 hosts up) scanned in 25.50 seconds" elapsed="25.50" exit="success"/>
<hosts up="2" down="1" total="3"/>
</runstats>
</nmaprun>`

// Fixture XML IPv6
const ipv6NmapXML = `<?xml version="1.0" encoding="UTF-8"?>
<nmaprun scanner="nmap" args="nmap -6 -F -oX - 2001:db8::cafe:1" start="1700000100" version="7.94">
<host>
<status state="up" reason="syn-ack"/>
<address addr="2001:db8::cafe:1" addrtype="ipv6"/>
<ports>
<port protocol="tcp" portid="80">
<state state="open" reason="syn-ack"/>
<service name="http" product="nginx" version="1.24.0"/>
</port>
</ports>
</host>
<runstats>
<finished time="1700000105" elapsed="5.0" exit="success"/>
<hosts up="1" down="0" total="1"/>
</runstats>
</nmaprun>`

func TestParseNmapXML_MultiHostFixture(t *testing.T) {
	report, err := ParseNmapXML(strings.NewReader(realisticNmapXML))
	if err != nil {
		t.Fatalf("ParseNmapXML ha restituito un errore inatteso: %v", err)
	}

	// 1. Verifica Metadati Scanner
	if report.Scanner != "nmap" {
		t.Errorf("atteso scanner 'nmap', ottenuto '%s'", report.Scanner)
	}
	if report.Version != "7.94" {
		t.Errorf("attesa versione '7.94', ottenuta '%s'", report.Version)
	}
	if report.Duration != 25.50 {
		t.Errorf("attesa durata 25.50s, ottenuta %f", report.Duration)
	}
	if report.TotalHosts != 3 || report.UpHosts != 2 || report.DownHosts != 1 {
		t.Errorf("statistiche host errate: total=%d, up=%d, down=%d", report.TotalHosts, report.UpHosts, report.DownHosts)
	}

	// 2. Verifica Hosts
	if len(report.Hosts) != 3 {
		t.Fatalf("attesi 3 host, trovati %d", len(report.Hosts))
	}

	// Host 1: Linux Webserver
	h1 := report.Hosts[0]
	if h1.IP != "192.168.1.50" {
		t.Errorf("IP errato per host 1: %s", h1.IP)
	}
	if h1.IsIPv6 {
		t.Errorf("atteso IPv4 per host 1")
	}
	if h1.MAC != "00:0C:29:84:7B:A2" || h1.Vendor != "VMware" {
		t.Errorf("MAC o Vendor errati: MAC=%s, Vendor=%s", h1.MAC, h1.Vendor)
	}
	if h1.PrimaryHostname != "webserver.internal" {
		t.Errorf("hostname primario errato: %s", h1.PrimaryHostname)
	}
	if h1.OS == nil || h1.OS.Name != "Linux 5.4 - 5.15" || h1.OS.Accuracy != 96 {
		t.Errorf("OS non rilevato o errato: %+v", h1.OS)
	}
	if len(h1.Ports) != 4 {
		t.Fatalf("attese 4 porte per host 1, trovate %d", len(h1.Ports))
	}

	// Verifica Porta 22 SSH
	p22 := h1.Ports[0]
	if p22.PortID != 22 || p22.Protocol != "tcp" || p22.State != "open" {
		t.Errorf("stato porta 22 errato: %+v", p22)
	}
	if p22.Service == nil || p22.Service.Name != "ssh" || p22.Service.Product != "OpenSSH" || p22.Service.Version != "8.9p1" {
		t.Errorf("servizio porta 22 errato: %+v", p22.Service)
	}
	if len(p22.Service.CPEs) != 2 || p22.Service.CPEs[0] != "cpe:/a:openbsd:openssh:8.9p1" {
		t.Errorf("CPE porta 22 errati: %v", p22.Service.CPEs)
	}

	// Verifica Porta 80 HTTP e composizione Banner
	p80 := h1.Ports[1]
	if p80.Service == nil || p80.Service.RawBanner != "Apache httpd 2.4.52 (Ubuntu)" {
		t.Errorf("banner grezzo errato per porta 80: %v", p80.Service)
	}

	// Verifica Porta 8080 Chiusa
	p8080 := h1.Ports[3]
	if p8080.PortID != 8080 || p8080.State != "closed" {
		t.Errorf("stato porta 8080 errato: %+v", p8080)
	}

	// Host 2: Windows Domain Controller
	h2 := report.Hosts[1]
	if h2.IP != "192.168.1.100" || h2.PrimaryHostname != "dc01.corp.local" {
		t.Errorf("dati errati per host 2: IP=%s, Hostname=%s", h2.IP, h2.PrimaryHostname)
	}
	if h2.OS == nil || h2.OS.Name != "Microsoft Windows Server 2019" {
		t.Errorf("OS errato per host 2: %+v", h2.OS)
	}
	if len(h2.Ports) != 2 {
		t.Fatalf("attese 2 porte per host 2, trovate %d", len(h2.Ports))
	}

	// Host 3: Down
	h3 := report.Hosts[2]
	if h3.Status != "down" || h3.IP != "192.168.1.200" {
		t.Errorf("stato errato per host 3: status=%s, IP=%s", h3.Status, h3.IP)
	}

	// 3. Verifica ToSummaryStats
	stats := report.ToSummaryStats()
	if stats.TotalHosts != 3 {
		t.Errorf("TotalHosts errato in summary stats: %d", stats.TotalHosts)
	}
	if stats.UpHosts != 2 {
		t.Errorf("UpHosts errato in summary stats: %d", stats.UpHosts)
	}
	if stats.TotalOpenPorts != 5 { // 3 aperte su host 1 (22, 80, 443) + 2 su host 2 (445, 3389)
		t.Errorf("TotalOpenPorts errato: atteso 5, ottenuto %d", stats.TotalOpenPorts)
	}
	if stats.DurationSeconds != 25.50 {
		t.Errorf("DurationSeconds errato: %f", stats.DurationSeconds)
	}

	// 4. Verifica ExtractCPEs
	cpes := report.ExtractCPEs()
	expectedCPEs := map[string]bool{
		"cpe:/a:openbsd:openssh:8.9p1":      true,
		"cpe:/o:canonical:ubuntu_linux":     true,
		"cpe:/a:apache:http_server:2.4.52":  true,
		"cpe:/o:microsoft:windows_server_2019": true,
	}

	if len(cpes) != len(expectedCPEs) {
		t.Errorf("numero di CPE unici errato: attesi %d, trovati %d (%v)", len(expectedCPEs), len(cpes), cpes)
	}
	for _, cpe := range cpes {
		if !expectedCPEs[cpe] {
			t.Errorf("CPE inatteso: %s", cpe)
		}
	}
}

func TestParseNmapXML_IPv6Fixture(t *testing.T) {
	report, err := ParseNmapXMLBytes([]byte(ipv6NmapXML))
	if err != nil {
		t.Fatalf("ParseNmapXMLBytes ha restituito un errore inatteso: %v", err)
	}

	if len(report.Hosts) != 1 {
		t.Fatalf("atteso 1 host IPv6, trovati %d", len(report.Hosts))
	}

	h := report.Hosts[0]
	if h.IP != "2001:db8::cafe:1" {
		t.Errorf("IP IPv6 errato: %s", h.IP)
	}
	if !h.IsIPv6 {
		t.Error("IsIPv6 dovrebbe essere true")
	}
	if len(h.Ports) != 1 || h.Ports[0].PortID != 80 {
		t.Errorf("porta errata: %+v", h.Ports)
	}
	if h.Ports[0].Service == nil || h.Ports[0].Service.Name != "http" {
		t.Errorf("servizio errato: %+v", h.Ports[0].Service)
	}
}

func TestParseNmapXML_MalformedXML(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "XML troncato",
			input:   `<?xml version="1.0"?><nmaprun><host><address addr="192.168.1.1"`,
			wantErr: true,
		},
		{
			name:    "Testo non XML",
			input:   "Questa è una stringa arbitraria non XML",
			wantErr: true,
		},
		{
			name:    "XML privo di radice nmaprun",
			input:   `<otherdoc><status ok="true"/></otherdoc>`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseNmapXML(strings.NewReader(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseNmapXML() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// Test con reader nullo
	_, err := ParseNmapXML(nil)
	if err == nil || !errors.Is(err, domain.ErrValidation) {
		t.Errorf("atteso ErrValidation per reader nullo, ottenuto: %v", err)
	}

	// Test con byte array vuoto
	_, err = ParseNmapXMLBytes([]byte{})
	if err == nil || !errors.Is(err, domain.ErrValidation) {
		t.Errorf("atteso ErrValidation per bytes vuoti, ottenuto: %v", err)
	}
}

func TestStreamNmapXML_Callback(t *testing.T) {
	var streamedHosts []ParsedHost

	callback := func(host ParsedHost) error {
		streamedHosts = append(streamedHosts, host)
		return nil
	}

	stats, err := StreamNmapXML(strings.NewReader(realisticNmapXML), callback)
	if err != nil {
		t.Fatalf("StreamNmapXML ha restituito errore inatteso: %v", err)
	}

	if len(streamedHosts) != 3 {
		t.Fatalf("attesi 3 host ricevuti via streaming, ricevuti %d", len(streamedHosts))
	}

	if streamedHosts[0].IP != "192.168.1.50" {
		t.Errorf("host 1 errato: %s", streamedHosts[0].IP)
	}
	if streamedHosts[1].IP != "192.168.1.100" {
		t.Errorf("host 2 errato: %s", streamedHosts[1].IP)
	}

	if stats.TotalHosts != 3 || stats.UpHosts != 2 || stats.DownHosts != 1 {
		t.Errorf("stats finali errate: %+v", stats)
	}
}

func TestStreamNmapXML_CallbackErrorPropagation(t *testing.T) {
	customErr := errors.New("interruzione da parte della callback")

	callback := func(host ParsedHost) error {
		if host.IP == "192.168.1.100" {
			return customErr
		}
		return nil
	}

	_, err := StreamNmapXML(strings.NewReader(realisticNmapXML), callback)
	if err == nil {
		t.Fatal("atteso errore propagato dalla callback, ottenuto nil")
	}

	if !strings.Contains(err.Error(), "interruzione da parte della callback") {
		t.Errorf("messaggio di errore inatteso: %v", err)
	}
}

