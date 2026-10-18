# PROJECT STATE

**Ultimo aggiornamento:** 2026-10-18 · working tree.
Memoria corrente: aggiornare solo fatti persistenti. Task/stati dettagliati in [TASK_LIST](04-development/TASK_LIST.md); evidenze precedenti nello [storico](../.agent/history/PROJECT_STATE-2026-10-18.md). Lo storico non autorizza nuove azioni.

## Project Profile

NextGen-CyberSecurity: Go 1.22+ / Chi / pgx, React 18 / TypeScript / Vite; PostgreSQL 16 e Nmap nel container. Fonti operative: [Backend](04-development/BACKEND_SETUP.md), [Frontend](04-development/FRONTEND_SETUP.md), [Test Plan](05-qa/TEST_PLAN.md#strategia).
Runner dalla root: `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/verify.ps1`, profili Quick/Final/Full/Docs. Go rapido esclude DB; PostgreSQL richiede build tag integration e TEST_DB_* esplicite su DB locale dedicato. UI Docker fornita dal PO: localhost:3080; Vite nativo: 3000.
Preferenza PO persistente: lavorare su main senza nuovi branch; commit/push/merge/deploy richiedono scope esplicito. Lavoro preesistente M1.1 preservato; nessuna autorizzazione remota per AGENT-HARNESS-01.

## Baseline test (prima di ogni intervento)

Consultare, non rilanciare automaticamente; la freschezza dipende dagli input effettivi, non solo da HEAD.
- M11-CLOSE-01 (2026-10-18): backend suite 20,816s/vet 1,189s/build 1,432s; frontend 168 test/build PASS exit 0. Prima build UI FAIL per fixture TS, poi corretta. Fonte: [matrice QA](05-qa/M11_CLOSURE_QA.md).
- M11-REPORT-01 (2026-10-18): frontend 175 test 4,995s/build 10,805s PASS exit 0; report `20261018-113105-c03e6e33.json`. Backend invariato riutilizzato nel relativo incremento. UI/PDF non verificati.
- PostgreSQL SSH/migrazione 000010: prove dedicate up/down/up PASS storiche nel [piano SSH](04-development/plans/SSH-HARDEN-01.md); non sono un nuovo PASS DB di questa working tree.
- AGENT-HARNESS-01: Docs PASS exit 0 (verificatore 0,792s, diff 0,065s); 22 prove harness PASS exit 0 (2,276s). Report `20261018-114735-c1bb6cad.json`; controlli preservazione PASS. Nessuna suite applicativa attribuita a questa revisione.

## In corso

- **M11-CLOSE-01 / M11-REPORT-01**: input firewall, timer obbligatorio e filtri archivio report implementati/verificati automaticamente. Milestone non conclusa: UI/PDF, firewall/watchdog Linux e persistenza discovery restano aperti. [Piano](04-development/plans/M11-CLOSE-01.md), [QA](05-qa/M11_CLOSURE_QA.md).
- **SSH-HARDEN-01 / E28-S1-T1**: pin host e deadline implementati con test; gate visuale e Linux/rollback/multiistanza aperti. TECH-DEBT-06/07 non chiusi dalle sole deadline locali.
- **AGENT-HARNESS-01**: implementato/verificato, attende PO; quattro skill locali, stato compatto e mapping eseguibile. [Piano/fonti/evidenze](04-development/plans/AGENT-HARNESS-01.md).

## Accettati recentemente dal PO (`[x]`)

- 2026-10-17: DOCKER-CI-01; SSH-TOOLS-01; SSH-CONNECT-02; DISCOVERY-01; TEST-OPT-01; COMPANY-UX-02; COMPANY-PDF-03; AP-REVIEW-01; REMOVE-01. Limiti visuali accettati non diventano prove autonome.
- 2026-10-16: AGENT-CLEAN-03, AGENT-CODEX-01, SCAN-UX-01, SCAN-TIME-01 e SCAN-COPY-01.
- Accettazioni più vecchie e dettagli: [CHANGELOG](../CHANGELOG.md) e storico collegato sopra. Nessuno stato di accettazione applicativo cambiato dalla compattazione.

## In attesa di accettazione utente

- CI-CONTEXT-01: actionlint rosso→verde, porta PostgreSQL risolta nello step. CI remota non verificata; [pipeline](06-devops/CI_CD_PIPELINE.md).
- README-GH-01: consegna editoriale e URL Docker 3080; pubblicazione storica registrata, nessun nuovo permesso remoto.
- AGENT-HARNESS-01: `[~]`, gate documentali e 22 prove PASS; provare il profilo Docs e consultare i nuovi ingressi skill. Beneficio token/qualità Sol non misurato. M1.1 e SSH-HARDEN restano incompleti.

## Bloccati / Domande aperte per l'utente

- Gate browser/PDF M1.1 e pinning SSH non eseguiti. Disponibilità strumenti da ricontrollare al task pertinente, non assunta permanente.
- Laboratorio Linux E28/watchdog non disponibile nell'ultima evidenza: nessun host reale autorizzato per prove distruttive. TECH-DEBT-07 richiede transazione remota, cancellazione e coordinamento multiistanza.
- SSH-CONNECT-01 è una diagnosi storica superata dalla connessione confermata dal PO; il relativo `[!]` nella task list è conservato come storia, non richiede ripetere vecchie domande.
- TECH-DEBT-05 consenso API, 08 modello Attack Path, 09 advisory dipendenze, 10 costo archivio report e debiti 01..04 restano nel registro task.

## Assunzioni confermate

Azienda→reparto a due livelli nelle nuove scritture; SSH/clienti indipendenti dallo scope aziende; Attack Path euristico 2D, non prova di compromissione; Executive con cinque priorità per reparto; discovery non recupera identità storiche senza MAC verificabile.
DOCKER-CI: Linux amd64, artifact 7 giorni, PostgreSQL separato, nessun registry/deploy. Firewall PO: iptables/iptables-nft con modello corrente, nessun motore nft nativo. Pin host fornito tramite canale indipendentemente fidato.
Assunzione AGENT-HARNESS-01: skill condivise a livello repository, nessuna modifica a config personale o dipendenze; beneficio comportamentale su Sol da misurare.

## Decisioni recenti (ultime 10, dettagli negli ADR)

- ADR-23: contesto progressivo, skill specifiche e verificatori deterministici; [decisione](03-architecture/ADR/README.md#adr-23).
- ADR-22: archivio report completo via paginazione esistente; costo lineare esplicito.
- ADR-21: timer firewall obbligatorio, input validati e commenti quotati.
- ADR-20: pin host separato e deadline SSH; nessuna garanzia remota dedotta.
- ADR-19: gate finale condiviso e PostgreSQL esplicito.
- ADR-18: dashboard unica e clienti SSH indipendenti.
- ADR-17: Attack Path euristico; supera l'interpretazione probabilistica ADR-06.
- ADR-16: blocco porta sul server del chokepoint.
- ADR-15: telemetria e ETA euristica del lotto.
- ADR-14: profili espliciti ed editor Nmap.

## Prossimi passi

Concludere gate governance e accettazione separata; proseguire M1.1 secondo dipendenze già approvate, senza riordinare milestone. Completare collaudi SSH/UI/PDF e TECH-DEBT-07 quando il laboratorio è disponibile. Nessun commit, push o deploy implicito.

## Log sessioni (ultime 5, la più recente in alto)

- 2026-10-18 · AGENT-HARNESS-01: Docs/22 prove/preservazione PASS, self-review e Doc Sync; nessun commit/push/deploy. Stato `[~]`; Python non installato perché nessun vantaggio misurato rispetto al gate PowerShell di pochi secondi.
- 2026-10-18 · M11-REPORT-01: filtri e archivio paginato; gate frontend PASS, UI/PDF aperti.
- 2026-10-18 · M11-CLOSE-01: validazione e timer firewall; gate automatici PASS, milestone aperta.
- 2026-10-18 · CI-CONTEXT-01 / README-GH-01: actionlint e consegna editoriale; evidenze nei documenti dedicati.
- 2026-10-17 · Pubblicazione/rimozione e accettazioni: dettagli nello storico; nessuna nuova autorizzazione.
