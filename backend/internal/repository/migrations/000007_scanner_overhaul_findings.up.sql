-- 000007_scanner_overhaul_findings.up.sql
-- SEC-P1: Scanner Engine Overhaul, Invasiveness Classification & Finding Data Model Enrichment

-- 1. Estensione findings con confidence, detection_method, cpe e port_state
ALTER TABLE findings
    ADD COLUMN IF NOT EXISTS confidence VARCHAR(20) NOT NULL DEFAULT 'CONFIRMED',
    ADD COLUMN IF NOT EXISTS detection_method VARCHAR(30) NOT NULL DEFAULT 'banner',
    ADD COLUMN IF NOT EXISTS cpe VARCHAR(255),
    ADD COLUMN IF NOT EXISTS port_state VARCHAR(20) NOT NULL DEFAULT 'open';

CREATE INDEX IF NOT EXISTS idx_findings_confidence ON findings(confidence);
CREATE INDEX IF NOT EXISTS idx_findings_detection_method ON findings(detection_method);
CREATE INDEX IF NOT EXISTS idx_findings_port_state ON findings(port_state);

-- 2. Estensione scan_profiles con classificazione di invasività
ALTER TABLE scan_profiles
    ADD COLUMN IF NOT EXISTS invasiveness VARCHAR(20) NOT NULL DEFAULT 'SAFE';

CREATE INDEX IF NOT EXISTS idx_scan_profiles_invasiveness ON scan_profiles(invasiveness);

-- 3. Assegnazione invasività per i profili di sistema esistenti
UPDATE scan_profiles
SET invasiveness = 'INTRUSIVE'
WHERE name IN (
    'All-in-One 360° Comprehensive Audit',
    'Full Deep Scan',
    'OS & System Daemon Flaws Audit',
    'Critical Services & RCE Surface',
    'Firewall & Filter Bypass Check'
);

UPDATE scan_profiles
SET invasiveness = 'SAFE'
WHERE name NOT IN (
    'All-in-One 360° Comprehensive Audit',
    'Full Deep Scan',
    'OS & System Daemon Flaws Audit',
    'Critical Services & RCE Surface',
    'Firewall & Filter Bypass Check'
);

-- 4. Ricalibrazione parametri profili di default (SEC-P1-02: porta 69 rimossa da TCP, rimozione -n per reverse DNS)
UPDATE scan_profiles
SET nmap_args = '-sV -p 21,23,79,110,143,512,513,514 -T3',
    description = 'Verifica conformità e sicurezza per rilevare servizi in chiaro e protocolli legacy obsoleti (Telnet, FTP, Finger, POP3, IMAP, Rexec, Rsh, Rlogin)'
WHERE name = 'Legacy & Insecure Protocol Check';

UPDATE scan_profiles
SET nmap_args = '-sn -PE -PS22,80,443,445 -T4',
    description = 'Ricognizione LAN ultra-rapida con ping sweep, ICMP echo e probe SYN per mappare tutti gli host accesi e risolvere reverse DNS'
WHERE name = 'Fast Ping & Host Sweep (LAN)';

-- 5. Inserimento Profilo Specialistico UDP (SEC-P1-01)
INSERT INTO scan_profiles (id, name, nmap_args, description, timeout_seconds, category, invasiveness, is_custom, tenant_id)
VALUES (
    gen_random_uuid(),
    'Specialized UDP Critical Services',
    '-sU -p U:53,69,123,161,514,623,5353 -T4',
    'Audit mirato sui servizi UDP ad alto rischio: DNS (53), TFTP (69), NTP (123), SNMP (161), Syslog (514), IPMI (623) e mDNS (5353)',
    1200,
    'PEN_TEST',
    'SAFE',
    false,
    NULL
)
ON CONFLICT (name) DO UPDATE SET
    nmap_args = EXCLUDED.nmap_args,
    description = EXCLUDED.description,
    timeout_seconds = EXCLUDED.timeout_seconds,
    category = EXCLUDED.category,
    invasiveness = EXCLUDED.invasiveness;

