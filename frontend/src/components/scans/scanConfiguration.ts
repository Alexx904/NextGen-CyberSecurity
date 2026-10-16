export interface ScanPreset {
  id: string;
  name: string;
  description: string;
  args: string;
  timeout: number;
}

export const SCAN_PRESETS: ScanPreset[] = [
  { id: 'services', name: 'Porte TCP e versioni (1000)', description: 'Individua le porte aperte tra le 1000 TCP più frequenti e identifica i servizi esposti e le relative versioni.', args: '-sS -sV --top-ports 1000 -T3', timeout: 3600 },
  { id: 'quick', name: 'Porte TCP frequenti (100)', description: 'Esegue un controllo rapido delle 100 porte TCP più frequenti per individuare i servizi esposti.', args: '-sS -F -T3', timeout: 900 },
  { id: 'full', name: 'Tutte le porte TCP (1–65535)', description: 'Esamina tutte le porte TCP da 1 a 65535 per individuare servizi esposti anche su porte non standard e rilevarne le versioni.', args: '-sS -sV -p- -T3', timeout: 14400 },
  { id: 'web', name: 'Servizi su porte web', description: 'Ricerca servizi sulle principali porte utilizzate dai server web e ne identifica software e versioni.', args: '-sS -sV -p 80,443,3000,5000,8000,8080,8443,8888,9000 -T3', timeout: 1200 },
  { id: 'vuln', name: 'Controlli vulnerabilità NSE', description: 'Identifica i servizi sulle 1000 porte TCP più frequenti ed esegue gli script NSE della categoria vuln per ricercare vulnerabilità note.', args: '-sS -sV --top-ports 1000 --script=vuln -T3', timeout: 3600 },
  { id: 'udp', name: 'Porte UDP selezionate', description: 'Verifica lo stato delle porte UDP associate a DNS, TFTP, NTP, SNMP, Syslog, IPMI e mDNS.', args: '-sU -p U:53,69,123,161,514,623,5353 -T3', timeout: 1200 },
  { id: 'legacy', name: 'Servizi TCP su porte legacy', description: 'Ricerca servizi sulle porte TCP comunemente associate a FTP, Telnet, Finger, POP3, IMAP e servizi di accesso remoto legacy, rilevandone le versioni.', args: '-sS -sV -p 21,23,79,110,143,512,513,514 -T3', timeout: 900 },
  { id: 'ack', name: 'Filtraggio TCP ACK', description: 'Analizza le risposte a sonde TCP ACK per distinguere le porte filtrate da quelle non filtrate e valutare il filtraggio di rete.', args: '-sA -Pn --top-ports 1000 -T3', timeout: 1800 },
];

// Illustrative planning ranges for one responsive IP on a stable LAN at T3, not benchmarks/SLA.
export const PROFILE_DURATION_ESTIMATES: Record<string, readonly [number, number]> = {
  services: [60, 300], quick: [10, 60], full: [600, 3600], web: [15, 120],
  vuln: [300, 1800], udp: [60, 600], legacy: [15, 120], ack: [30, 180],
};

export function formatDuration(seconds: number): string {
  if (seconds < 60) return `${Math.max(1, Math.ceil(seconds))} s`;
  if (seconds < 3600) return `${Math.ceil(seconds / 60)} min`;
  return `${Math.ceil(seconds / 360) / 10} h`;
}

export function profileDurationLabel(id: string): string {
  const estimate = PROFILE_DURATION_ESTIMATES[id];
  return estimate ? `${formatDuration(estimate[0])} – ${formatDuration(estimate[1])}` : 'Non disponibile';
}

export function normalizeNmapInput(input: string): string {
  return input.trim().replace(/^nmap(?:\s+|$)/, '').trim().replace(/\s+/g, ' ');
}

// Only normalize token spelling, never reorder arguments whose order can matter.
export function canonicalArgs(input: string): string {
  return normalizeNmapInput(input).replace(/-p\s+([^ ]+)/g, '-p$1').replace(/--script\s+(vuln|vulners)(?= |$)/g, '--script=$1');
}

export function isIntrusive(args: string): boolean {
  return /(?:^|\s)(?:--script(?:=|\s)(?:vuln|vulners)|-p(?:-|\s*[01]-65535)|-T[45]|--version-all|-sA|--min-rate)(?=\s|$)/.test(args);
}

// Preliminary feedback only; the backend remains authoritative for the allowlist.
export function commandError(input: string): string | null {
  const args = normalizeNmapInput(input);
  if (!args) return 'Inserisci le opzioni Nmap.';
  if (/[;&|$><`\\"'(){}*?!]/.test(args)) return 'Metacaratteri e comandi shell non sono ammessi.';
  const tokens = args.split(' ');
  const flags = new Set(tokens);
  let tcp = 0;
  let ports = 0;
  for (let i = 0; i < tokens.length; i++) {
    const token = tokens[i];
    if (/^-s[STANFX]$/.test(token)) tcp++;
    if (token === '-F') ports++;
    if (token === '-p' || /^-p.+/.test(token)) {
      ports++;
      const value = token === '-p' ? tokens[++i] : token.slice(2);
      if (value === '-') continue;
      if (!value || !value.split(',').every((entry) => {
        const range = entry.replace(/^[TU]:/, '').split('-');
        return range.length <= 2 && range.every((n) => /^\d{1,5}$/.test(n) && Number(n) <= 65535) && (range.length === 1 || Number(range[0]) <= Number(range[1]));
      })) return 'Porte non valide: usa numeri 0–65535 o intervalli crescenti.';
      continue;
    }
    if (['--top-ports', '--max-retries', '--min-rate', '--max-rate', '--version-intensity'].includes(token)) {
      const value = tokens[++i];
      if (!value || !/^\d+$/.test(value)) return `Valore numerico richiesto per ${token}.`;
      if (token === '--top-ports') {
        ports++;
        if (Number(value) < 1 || Number(value) > 65535) return '--top-ports: usa un valore tra 1 e 65535.';
      }
      if (token === '--version-intensity' && !/^[0-9]$/.test(value)) return 'Intensità versione: usa 0–9.';
      continue;
    }
    if (token === '--script') {
      if (!['vuln', 'vulners'].includes(tokens[++i])) return 'Sono ammessi solo gli script vuln e vulners.';
      continue;
    }
    if (/^-T[0-5]$/.test(token) || /^-P[SU]\d{1,5}(,\d{1,5})*$/.test(token)) continue;
    if (['-F', '-sV', '-O', '-Pn', '-sn', '-PR', '-PE', '-PP', '-sS', '-sT', '-sU', '-sA', '-sN', '-sF', '-sX', '-n', '-R', '-v', '-vv', '-q', '-d', '-6', '--version-light', '--version-all', '--osscan-limit', '--osscan-guess', '--open', '--packet-trace', '--script=vuln', '--script=vulners'].includes(token)) continue;
    return `Opzione non supportata: ${token}. Target e output sono gestiti dalla piattaforma.`;
  }
  if (tcp > 1 || ports > 1) return 'Scegli una sola tecnica TCP e un solo selettore di porte.';
  if (flags.has('-sn') && (tcp > 0 || ports > 0 || flags.has('-sU') || flags.has('-sV') || flags.has('-O'))) return '-sn cerca solo host attivi: rimuovi le opzioni di scansione porte, versioni e OS.';
  if (flags.has('-sA') && (flags.has('--open') || flags.has('-sV') || flags.has('-O'))) return 'ACK misura filtered/unfiltered: rimuovi --open, -sV e -O.';
  return null;
}
