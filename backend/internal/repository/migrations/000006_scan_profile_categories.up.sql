-- 000006_scan_profile_categories.up.sql
-- Estensione scan_profiles con categorie specialistiche, supporto profili custom e isolamento tenant

ALTER TABLE scan_profiles
    ADD COLUMN IF NOT EXISTS category VARCHAR(50) NOT NULL DEFAULT 'SYSTEM_FLAWS',
    ADD COLUMN IF NOT EXISTS is_custom BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_scan_profiles_category ON scan_profiles(category);
CREATE INDEX IF NOT EXISTS idx_scan_profiles_tenant_id ON scan_profiles(tenant_id);

-- Aggiornamento categorie per i profili base esistenti
UPDATE scan_profiles SET category = 'DISCOVERY' WHERE name = 'Fast Discovery';
UPDATE scan_profiles SET category = 'SYSTEM_FLAWS' WHERE name = 'Standard Service Scan';
UPDATE scan_profiles SET category = 'PEN_TEST' WHERE name = 'Full Deep Scan';

-- Inserimento del catalogo profili specialistici di sistema predefiniti (idempotente)
INSERT INTO scan_profiles (id, name, nmap_args, description, timeout_seconds, category, is_custom, tenant_id)
VALUES
    -- 0. All-in-One 360° Comprehensive Audit (Generalista Completo)
    (gen_random_uuid(), 'All-in-One 360° Comprehensive Audit', '-sS -sV -O -T4 --open', 'Scansione generalista completa a 360°: mappatura host vivi, scansione porte standard, identificazione versioni demoni (sV), fingerprinting del sistema operativo (O) ed elusione filtri base (-sS)', 5400, 'ALL_IN_ONE', false, NULL),

    -- 1. Penetration Testing & Offensive Emulation
    (gen_random_uuid(), 'Stealth SYN Penetration Audit', '-sS -Pn -T2 --open', 'Ricognizione stealth a basso rumore con pacchetti SYN senza handshake TCP completo, elusione baseline firewall e solo porte aperte', 3600, 'PEN_TEST', false, NULL),
    (gen_random_uuid(), 'Firewall & Filter Bypass Check', '-sA -Pn -T3', 'Mappatura filtri firewall perimetrali e apparati IDS con pacchetti TCP ACK per verificare se i firewall sono stateful o stateless', 1800, 'PEN_TEST', false, NULL),

    -- 2. Falle di Sistema & Vulnerability Assessment
    (gen_random_uuid(), 'OS & System Daemon Flaws Audit', '-sV -O --version-intensity 7 -T4', 'Rilevamento dettagliato del sistema operativo e fingerprinting approfondito dei demoni di sistema (SSH, SMB, RDP, NetBIOS, RPC)', 3600, 'SYSTEM_FLAWS', false, NULL),
    (gen_random_uuid(), 'Critical Services & RCE Surface', '-sV -p 21,22,23,25,53,80,110,135,139,143,443,445,1433,1521,3306,3389,5432,5900,6379,8080,8443 -T4', 'Audit mirato sui vettori critici più colpite da exploit remoti e vulnerabilità note RCE/PrivEsc (DB, Remote Desktop, Mail, Condivisioni)', 1800, 'SYSTEM_FLAWS', false, NULL),

    -- 3. Web Application & API Services
    (gen_random_uuid(), 'Web & API Surface Audit', '-sV -p 80,443,8000,8080,8443,8888,3000,5000,9000 -T4 --open', 'Rilevamento mirato di server HTTP/HTTPS, API REST, backend microservizi e reverse proxy applicativi (Nginx, Apache, Traefik)', 1200, 'WEB_APP', false, NULL),
    (gen_random_uuid(), 'Full Web Stack Deep Versioning', '-sV --version-all -p 80,443,8008,8080,8443,9090,9443 -T4', 'Versioning approfondito con interrogazione intensiva dei banner HTTP e intestazioni TLS per scovare componenti obsoleti', 2400, 'WEB_APP', false, NULL),

    -- 4. Network & Asset Discovery
    (gen_random_uuid(), 'Fast Ping & Host Sweep (LAN)', '-sn -PE -PS22,80,443,445 -T4 -n', 'Ricognizione LAN ultra-rapida con ping sweep, ICMP echo e probe SYN senza port scan per mappare tutti gli host accesi', 120, 'DISCOVERY', false, NULL),

    -- 5. Compliance & Hardening Check
    (gen_random_uuid(), 'Legacy & Insecure Protocol Check', '-sV -p 21,23,69,79,110,143,512,513,514 -T3', 'Verifica conformità e sicurezza per rilevare servizi in chiaro e protocolli legacy obsoleti (Telnet, FTP, TFTP, Rsh, Rlogin, POP3)', 900, 'COMPLIANCE', false, NULL)
ON CONFLICT (name) DO UPDATE SET
    category = EXCLUDED.category,
    description = EXCLUDED.description,
    nmap_args = EXCLUDED.nmap_args,
    timeout_seconds = EXCLUDED.timeout_seconds;

