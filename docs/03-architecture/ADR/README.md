# Architecture Decision Records

## ADR-30 - Automatic scan-based regulatory context replaces manual compliance (2026-10-19)

- Status: PO product direction approved; implementation verification tracked separately.
- Decision: supersede ADR-29's active manual UI/API/scorecard/Audit workflow with
  deterministic context derived from findings in existing REST/SSE and PDF flows.
  No manual assessment, identity registration, questionnaire or SLA input is needed.
  The former route redirects to Findings; manual API routes/export are retired.
- Evidence boundaries: preserve scan/finding identity, port state, source confidence
  and detection method. CVE relevance and transport review are technical mappings,
  never proof of violation, encryption absence, exploitability or certification.
  Empty mappings do not establish safety. Catalogue updates require a release.
- Compatibility: retained migrations/data and dormant implementation avoid destructive
  rollback. No external provider, new schema or dependency. ADR-29 remains historical;
  its immutable-history principles remain applicable to retained records.
- [Scope, sources and verification](../../04-development/plans/M4-AUTOMATIC-CONTEXT.md).

> Responsabile: Software Architect. ADR storici preservati; nuove decisioni significative ricevono ID incrementale, contesto, decisione e conseguenze.

## 7. Architecture Decision Records (ADR)

### ADR-29 - Evidence-backed compliance and persistent issue streams (2026-10-19)

- Status: implemented and verified in working tree, awaiting PO acceptance; refines ADR-28 for M4 without superseding M5-M8 plans.
- Context: scan targets can be ranges/dynamic addresses; observations alone cannot identify a persistent host, establish organizational compliance or close remediation.
- Decision: explicitly register stable scoped assets and associate observations; retain immutable source snapshots and ordered issue commands. Serialize writes plus expected versions, replay shared domain rules, preserve completed-cycle history on recurrence. Explicit configurable SLA, no regulatory default. Migration 14 adds four tables and rejects evidence/event mutation; referenced audit history restricts inventory deletion.
- Controls: versioned curated relevance rules, evidenced manual assessments, unknown-inclusive denominators, explicit exclusions and conservative company rollup. Latest incomplete/missing coverage cannot earn positive technical credit. Mappings are not legal applicability or certification.
- Reporting: existing report package/routes/UI provide Audit A4 from one repeatable-read dataset, cutoff/MVCC ID, source references and limitations. Period is declared review metadata, not a historical filter; no stored snapshot/replay promise, no separate report archive.
- Consequences: conservative manual identity resolution, full scoped history loading and referenced-history deletion restrictions are explicit tradeoffs. No automated executor or external notification. [Contract, gates and rollback](../../04-development/plans/M4-COMPLIANCE.md).

### ADR-28 · Integrated assessment lifecycle and reporting (2026-10-19)

- Status: direction approved by PO; implementation pending. Documentation only, no runtime/API/schema change.
- Context: M4-M8 overlap implemented remediation, company/department consolidation and Executive/Technical exports; per-scan findings lack a complete cross-scan remediation lifecycle.
- Decision: extend one company assessment workflow and one reporting pipeline. Design persistent issue identity linked to observations, lifecycle and evidence first in M4. Reuse this data for compliance, AI, applicability, validation and retest; Executive/Technical/Audit profiles use consistent explicit scope/cutoff. Preserve separate detection confidence, execution outcomes and metric meanings.
- Boundaries: SSH/Firewall remains independent (ADR-18 preserved); UI-TOOLS-01 owns final routes. M7/M8 share infrastructure with separate permissions and engagement gates. ADR-27 scan consent is not validation/exploitation authorization. Notifications remain deferred.
- Consequences: no parallel report/finding store or automatic remediation executor. Define identity, MTTR, report consistency, retention, auth/scope, compatibility and migration up/down before implementation. Resolve relevant Attack Path coverage debt before consuming inferred paths. Module availability does not imply applicability/success; failed proof does not establish safety.
- Supersession: refines separate future dossier/AI/pentest reporting descriptions in PRD/FR-20..24. Existing accepted behavior, IDs and evidence remain unchanged.

### ADR-27 · Inventory scan authorization snapshots (2026-10-19)

- Context: UI-only elevated-impact consent could be bypassed through the API or
  scheduler; private addresses do not establish network ownership.
- Decision: PO requires consent for every inventory scan. Confirm exact target
  IDs/values, options, timeout and recurrence; derive actor/time server-side.
  Compare before enqueue and execution. Migration 000013 records source and
  append-only evidence transactionally, with no cascading inventory references.
  Recurring authorization survives restart only while its configuration remains
  unchanged. Reload schedules at trigger/execution; metadata writes never restore
  disabled state. Legacy schedules require explicit replacement, never inferred
  consent. Next-run timestamps and cron use the same server-local clock.
- Consequences: API clients must supply the snapshot; existing queued work without
  consent fails before Nmap. Audit survives inventory deletion, with no automatic
  purge in this increment. Database owners can remove protections. Consent is not
  legal ownership verification; in-flight work is not recalled by later edits.
  Standalone discovery remains a separate workflow. [Contract and QA](../../04-development/plans/TECH-DEBT-05.md).

### ADR-26 · Transactional alert outbox and restricted egress (2026-10-18)

- Context: in-memory notification callbacks could lose persisted findings on a
  crash. Tenant-admin destinations also create SSRF and credential disclosure
  risks, including secret URL paths in client diagnostics.
- Decision: migration 000012 adds encrypted channel configuration, observation
  baselines and a durable outbox. Source triggers queue events in the source
  transaction; tenant advisory locks serialize observations across host batches.
  Job claims use SKIP LOCKED, leases and tokens, six attempts and bounded
  exponential backoff with provider delay. Channel changes cancel unclaimed
  old messages; workers recheck current tenant/company/revision before sending.
  Admin-only destination management and explicit synthetic-test consent; no
  plaintext configuration returned or logged. Public HTTPS/TLS SMTP only,
  validated DNS at dial time, no redirects or environment proxies.
- Consequences: at-least-once delivery, possible duplicates and in-flight sends
  after disable are explicit. Transactional enqueue failure fails the source
  write. Tenant write serialization, historical baseline cost, retained delivery
  audit and queue capacity need operational measurement. Private relays require
  a separately reviewed policy. Root companies with channels cannot be silently
  reparented/deleted. Existing history is not replayed to a newly added channel.
  Teams uses current Workflows cards. TLS-certificate expiry remains unimplemented
  until an observed certificate model exists; it is not inferred from banners.
  [Contract and QA plan](../../04-development/plans/E28-S1-T4-alerting.md).

### ADR-25 · Host-owned firewall transaction (2026-10-18)

- Context: local timers and separate remote commands allowed overlap across
  backend instances and late changes after local SSH cancellation.
- Decision: snapshot, ready detached watchdog and mutation share a remote flock
  per managed UID. Private host state binds a random token to tenant/server;
  confirm/rollback/expiry use the same lock and token, without killing PIDs.
  Persist the returned backup before acknowledging; DB failure keeps protection.
  Snapshot travels on SSH stdin. Revision-checked editor prevents stale writes;
  move preserves saved rule data, replacement explicitly replaces form fields.
- Supersedes the process-local lifecycle of ADR-16 and historical PID cancellation
  in the SSH domain. Extends ADR-20/21 without weakening pins or mandatory timers.
- Consequences: remote tools/sudo prerequisites are checked without installation;
  same UID is the coordination boundary. Host reboot, deleted temporary state,
  independent administrators/accounts remain limitations. No new schema/library.
  Linux iptables-nft and real OpenSSH disconnect/late-mutation checks are separate
  from unit tests; browser gate remains required. [Plan](../../04-development/plans/M11-FIREWALL-01.md).

### ADR-24 · Snapshot discovery durevoli e host streaming (2026-10-18)

- Contesto: `sessionResults` viveva solo nel processo; `StreamNmapXML` esisteva
  ma il runner consegnava host/finding dopo la fine di Nmap. Replay letto prima
  della sottoscrizione; topic discovery privo del tenant.
- Decisione: due tabelle normalizzate con FK tenant/sessione e upsert dispositivo
  per IP; scrittura prima di SSE, errore DB interrompe senza successo. Runner
  collega stdout al parser tramite pipe e attende il join prima di ritornare;
  errore callback cancella il processo. Finding gestiti nella stessa callback,
  parsing finale senza duplicati. Subscribe prima del replay, topic tenant/sessione.
- Compatibilità: payload invariati, 409 per ID sessione già usato, ultima sessione
  anche incompleta; UI espone snapshot parziali ed errori di recupero. Nessuna
  ripresa automatica del job dopo restart; pub/sub live ancora in memoria locale.
- Conseguenze: migrazione 000011 con down distruttivo del solo storico discovery;
  backup delle nuove tabelle prima di rollback su dati da conservare. Nessuna nuova
  dipendenza applicativa; test DB dedicati, streaming subprocess e gate All.
  [Contratto e prove](../../04-development/plans/M11-PERSIST-01.md).

<a id="adr-23"></a>

### ADR-23 · Contesto progressivo e procedure specializzate (2026-10-18)

- Contesto: bootstrap esteso, regole ripetute e stato storico mescolato al corrente; glob non corrispondenti agli handler effettivi e comando migrazione privo del tag integration.
- Decisione: AGENTS compatto con gate essenziali; quattro skill repository e playbook di ingresso; stato corrente separato da snapshot storico; mapping Markdown verificabile e controlli di manifest, riferimenti e liste. [Piano e fonti](../../04-development/plans/AGENT-HARNESS-01.md).
- Supersessione parziale di ADR-10/12: lettura per argomento anziché soglia righe, handoff comune, Doc Sync solo sulle fonti cambiate e aggiornamento stato per fatti persistenti. Conservati SDLC, ownership, contratti, sicurezza e accettazioni.
- Conseguenze: nessuna modifica al modello/config personale o ai gate applicativi; dipendenze invariate. Riduzione caratteri del bootstrap misurabile, qualità/routing/costo su Sol richiedono valutazione comportamentale. Python assente: manifest semplici validati da PowerShell, nessuna esecuzione del validatore Python dichiarata.

### ADR-22 · Filtri sullo storico report completo (2026-10-18)

- Contesto: UI troncava lo storico a 100/10 elementi; API non espone filtro profilo.
  Il cronologico non applicava lo scope aziendale. FR-08 richiede filtri completi.
- Decisione: riusare la paginazione API esistente e caricare sequenzialmente tutte
  le pagine prima di aggregare/filtrare nel frontend. Stesso insieme per entrambe
  le viste, giorni UTC inclusivi e severità esatta. Il cronologico pagina localmente
  a 10 righe. Nessuna nuova API/SQL/migrazione; scope tenant resta lato backend.
- Errori: pagine duplicate, incomplete o con conteggio variato producono errore e
  retry esplicito. Risposte superate non aggiornano lo stato. Non è uno snapshot DB
  transazionale: ordinamento concorrente a conteggio invariato resta un limite.
- Tradeoff: costo iniziale proporzionale allo storico del tenant, memoria/browser
  e latenza da misurare su volumi reali; futuro filtro profilo e aggregazione server
  potranno sostituire il caricamento completo. Nessun limite silenzioso ai record.
- Decisione PO firewall nello stesso incremento: iptables/iptables-nft con modello
  corrente; nessun motore nft nativo. Collaudo della compatibilità ancora richiesto.

### ADR-21 · Protezione obbligatoria delle mutazioni firewall (2026-10-18)

- Contesto: il dominio richiede anti-lockout su ogni mutazione; flag API/UI
  permettevano di disattivarlo. Alcuni campi finivano nella shell senza validazione.
- Decisione: mantenere firme e campi legacy ma attivare sempre il timer per
  add/delete. Timeout 0 usa 60s, intervallo ammesso 10..300s come UI esistente.
  Validare IPv4/CIDR e interfacce prima di SSH; preservare i commenti come un
  argomento shell quotato, max 255 byte UTF-8 senza controlli.
- Conseguenze: le richieste legacy false ora richiedono backup/watchdog; eventuali
  errori preparatori bloccano la modifica. Nessuna nuova dipendenza o migrazione;
  nessuna garanzia multiistanza o rollback Linux dedotta. Questi gate e i gap
  funzionali M1.1 restano nella [matrice QA](../../05-qa/M11_CLOSURE_QA.md).

### ADR-20 · Pinning esplicito dell'identità host SSH (2026-10-17)

- Contesto: il fingerprint storico identifica la chiave client; il trasporto ignorava la host key. Timeout del solo dial non limita handshake e comandi.
- Decisione: `host_key_fingerprint` distinto, SHA256 canonico OpenSSH, configurato tramite un canale fidato. Connessioni senza pin o con mismatch rifiutate prima dell'autenticazione; nessun TOFU o apprendimento automatico. Rotazione via update SecOps/admin esistente; endpoint modificato senza pin esplicito perde fiducia. Reset status e last_connected_at quando cambia la fiducia.
- Compatibilità: migrazione 000010 up/down additiva; inventario legacy preservato ma connessioni bloccate fino a enrollment. Down perde i soli pin; re-upgrade richiede riconfigurazione. Non migrare fingerprint client in pin host.
- Deadline: handshake massimo 10s, comandi/test connessione massimo 30s con budget precedente del chiamante rispettato; cancellazione chiude la connessione dedicata. Terminale interattivo non limitato a 10s dopo handshake.
- Conseguenze: nessuna garanzia dedotta sulla terminazione del processo remoto. TECH-DEBT-07, transazione sull'host, coordinamento multiistanza e gate Linux E28 restano aperti. Piano [SSH-HARDEN-01](../../04-development/plans/SSH-HARDEN-01.md).

<a id="adr-19"></a>

### ADR-19 · Verifica proporzionata e PostgreSQL obbligatorio esplicito (2026-10-17)

- Contesto: suite/build richieste per ogni sotto-task aumentavano le ripetizioni; i test PostgreSQL potevano saltare la prova dopo tentativi di connessione falliti e la CI non forniva il DB.
- Decisione autorizzata dal PO: test mirati durante lo sviluppo, un gate finale completo per lato coinvolto; entrambi i lati per contratti comuni, sicurezza o impatto incerto. Riutilizzo solo di evidenze con input/ambiente invariati. Profilo Full include PostgreSQL; Quick non sostituisce la verifica finale.
- Implementazione: `scripts/verify.ps1` misura comandi e registra JSON temporanei; build tag `integration` separa i test PostgreSQL. Configurazione esclusivamente `TEST_DB_*`, host locale e nome dedicato `_test`/`_test_<id>`; nessun fallback applicativo e nessuno skip su indisponibilità DB. CI con terzo job indipendente PostgreSQL 16 effimero, health check e cache Go; esecuzioni obsolete dello stesso evento/ref cancellate.
- Conseguenze: `go test ./...` ora prova il percorso rapido, non il DB; chiamanti/CI devono richiedere le integrazioni esplicitamente. Nessuna migrazione/schema/API applicativa modificata. Timeout funzionali e asserzioni preservati. Profiling prima di refactor temporali; nessun risparmio percentuale dichiarato senza benchmark. Browser/PDF e rollback pertinenti restano gate separati. Strategia e limiti nel [Test Plan](../../05-qa/TEST_PLAN.md).

### ADR-18 · Dashboard unica e clienti SSH indipendenti (2026-10-17)

- Decisione PO: dashboard/CRUD aziende unificati; feature aziendali con selezione obbligatoria, SSH/Firewall esplicitamente indipendente e sempre disponibile. I clienti SSH sono etichette autonome, senza FK verso le aziende.
- Gerarchia nuova Azienda → Reparto protetta da UI/service e trigger PostgreSQL con lock per tenant sulle assegnazioni parent. Nessuna riscrittura di alberi legacy. Migrazioni 000008 (clienti SSH) e 000009 (vincolo di scrittura) con down.
- Report aggregati da radice e discendenti, attribuzione IP via target/scan, inventario PDF per reparto e colonna CSV additiva. PDF aziendale può superare due pagine. Scope reparto AP locale senza soglia CVSS UI; backend compatibile.
- Guard frontend è politica di navigazione, non autorizzazione API aggiuntiva. Rollback 000008 perde le etichette cliente: esportarle prima. Prove e limiti in [QA](../../05-qa/COMPANY_WORKFLOW_QA.md).

### ADR-17 · Attack Path come modello euristico verificabile (2026-10-17)

- Contesto: revisione PO di utilità cyber e grafo; la baseline confondeva esposizione, correlazione CVE e compromissione, con percorsi disconnessi e layout ciclico non terminante.
- Decisione: origine di analisi ipotetica, nessun trust da inventario SSH né pivot cross-subnet inventato. Relazioni /24 dichiarate ipotesi deboli; percorsi validati su archi orientati e score storico esplicitamente euristico. Provenienza e limiti visibili. Scope aziendale proiettato sull'intero grafo; JWT tenant invariato.
- UI: 2D percorso/completa, guida contestuale e tastiera; niente dipendenza 3D. La scelta riduce occlusioni e preserva la sequenza, senza dichiarare superiorità dimostrata da user testing.
- Compatibilità: API e firma servizio preservate, `analysis` additivo, identità dei nodi specifica per protocollo e snapshot. Nessuna migrazione. Modello corretto ma validazione reale di reachability/exploit e prova visuale restano distinte. Dettagli nel [dominio](../../domains/attack_path.md) e [report QA](../../05-qa/ATTACK_PATH_REVIEW.md).

### ADR-16 · Remediation sul server del chokepoint (2026-10-17)

- Contesto: E28-S1-T1 collega analisi euristica e mutazioni remote. Il PO sceglie blocco porta sul server del chokepoint; isolamento host separato.
- Decisione: preview ed esecuzione rivalidano il chokepoint SERVICE nel grafo del tenant; server SSH esplicito e IP coincidente, IPv4 TCP/UDP, protezione porta SSH. Riuso snapshot/watchdog e API firewall, DTO additivi, nessuna migrazione. Serializzazione del lifecycle locale e conservazione watchdog su errori.
- Conseguenze: DNS/IPv6, isolamento e gateway non supportati. Limiti di host-key, tempi remoti e deployment multiistanza registrati nel Tech Debt. Nessuna prova remota dichiarata. Contratto nel [dominio](../../domains/active_defense_and_soar.md).

### ADR-15 · Telemetria Nmap e monitor del lotto (2026-10-16)

- Contesto: il PO richiede durate per preset, progresso totale ed ETA dell'intero gruppo, non del solo target. Le precedenti percentuali di fase erano fisse.
- Decisione: statistiche XML interne ogni 5s, callback runner opzionale e snapshot volatili nell'orchestratore; GET scans per ID multipli con verifica tenant per ogni membro. Aggregazione client soltanto sugli ID del lancio, polling a blocchi senza sovrapposizione, ripristino lotto per tenant nella sessione. SSE mantenuto per dettaglio singolo.
- Conseguenze: niente connessioni SSE per ogni job del lotto, nessuna migrazione DB né nuova dipendenza. DTO/SSE estesi con campi opzionali. Percentuali e ETA totali restano euristiche; 100% di job terminati distingue successi/errori. Statistiche assenti/scadute non generano tempi inventati. In deployment multi-processo gli snapshot sono locali al processo: con routing su altri processi la UI può non avere ETA, senza fingere che sia disponibile; il deployment corrente usa singolo backend.
- Fonte e dettagli: [SCAN_PROGRESS_PLAN](../../04-development/SCAN_PROGRESS_PLAN.md).

<a id="adr-14"></a>

### ADR-14 · Profili espliciti ed editor Nmap (2026-10-15)

- **Contesto:** Tre modalità sovrapposte, nomi che promettono penetration test/compliance e preset selezionati per nomi non equivalenti agli argomenti dichiarati.
- **Decisione:** Due modalità: Profili con opzioni canoniche ed editor di argomenti entro allowlist. Riutilizzo soltanto per argomenti e timeout equivalenti. Target e XML restano gestiti dal backend; validazione semantica e invasività custom centralizzate sul server. Distinguere discovery di asset e scansioni persistite, con sovrapposizione dichiarata sulle porte frequenti.
- **Compatibilità:** Nessuna modifica DTO/schema né migrazione dei nomi/ID storici, per non alterare scansioni e scheduler esistenti. I nomi storici non sono più esposti nel catalogo della nuova modale. Nessuna dipendenza aggiunta.
- **Conseguenze:** Meno controlli duplicati e promesse tecnicamente corrette. Non è una CLI illimitata, un penetration test o una certificazione. Attestazione server di consenso resta un gap preesistente (TECH-DEBT-05); verifica visuale bloccata dal runtime senza browser.
- **Riferimento:** [Analisi e piano](../../04-development/NMAP_SCAN_REVIEW.md), [dominio scanner](../../domains/scanning_and_discovery.md).

- **ADR-01 (Modularizzazione Documentazione per Domini):** Scomposizione dei dettagli architetturali in `docs/domains/` mantenendo `2_ARCHITECTURE.md` snello per contratti globali ed ERD.
- **ADR-02 (Preservazione Storica Legacy):** Archiviazione storica dei documenti preesistenti in `docs/_legacy/` con tracciamento mapping.
- **ADR-03 (Branching Strategy):** Utilizzo di branch tematici dedicati per feature (`feat/`) e manutenzione (`chore/agent-onboarding`).
- **ADR-04 (Stream-Ticket Monouso & Hardening Perimetro Scansioni SEC-P0):**
  - **Contesto:** I token JWT a lungo termine inviati in query string (`?token=...`) per SSE e WebSocket venivano esposti nei log proxy/Nginx e nella cronologia browser. Inoltre, l'esecuzione di Nmap necessitava di barriere perimetrali contro l'iniezione di flag o target interni.
  - **Decisione:**
    1. Implementato `StreamTicketStore` in memoria, thread-safe con TTL 30s e auto-invalidation atomica al primo utilizzo. Endpoint `POST /api/v1/auth/stream-ticket` protetto da Bearer header.
    2. Separatore POSIX `--` anteposto sistematicamente al target nell'invocazione di Nmap in `BuildSafeArguments`.
    3. Deny-list preventiva in `CheckTargetScope` contro loopback, link-local, bridge Docker e host della piattaforma, con risoluzione DNS anti-rebinding.
    4. Watchdog di rollback firewall remoto via `nohup` distaccato per garanzia di ripristino anche in caso di caduta connessione SSH.
  - **Conseguenze:** Eliminazione del rischio di cred leakage nei log; sicurezza perimetrale dello scanner rafforzata; piena retrocompatibilità garantita.
- **ADR-05 (Scanner Engine Overhaul, Tassonomia di Invasività e Migrazione PDF SEC-P1):**
  - **Contesto:** Necessità di calibrazione specialistica delle scansioni UDP, eliminazione porte errate nei profili legacy (es. porta 69 TFTP su socket TCP), distinzione netta dell'impatto dei profili sulla stabilità di rete del cliente, e abbandono del package deprecato `jung-kurt/gofpdf`.
  - **Decisione:**
    1. Aggiunto profilo specialistico UDP (`ProfileSpecializedUDP`, porte `U:53,69,123,161,514,623,5353`).
    2. Tassonomia di invasività (`SAFE`, `INTRUSIVE`, `DANGEROUS`) con UI badge e gate di autorizzazione checkbox per profili ad alto impatto.
    3. Arricchimento modello `Finding` con `confidence`, `detection_method`, `cpe`, `port_state` (migrazione SQL `000007`).
    4. Hardening container in `docker-compose.yml`: mantenuta solo `CAP_NET_RAW`, rimosso `CAP_NET_ADMIN`, aggiunta env `NMAP_PRIVILEGED=1`.
    5. Migrazione libreria PDF al fork attivamente manutenuto `github.com/go-pdf/fpdf v0.9.0`.
  - **Conseguenze:** Zero falsi allarmi su profili TCP per porte UDP; controllo operativo totale prima di lanciare scansioni intrusive; compilazione PDF affidabile.
- **ADR-06 (Attack Path Probabilistic Risk Formula & Scan Differential SEC-P2):**
  - **Stato:** interpretazione probabilistica e inferenze del grafo superate da ADR-17; Scan Diff resta una decisione separata. Il testo seguente è storico, non il contratto corrente.
  - **Contesto:** Il rischio cumulativo calcolato tramite somma algebrica pura gonfiava artificialmente percorsi con molte vulnerabilità basse; gli archi pivot non esprimevano il grado di incertezza euristica; mancava il tracciamento dei finding risolti tra scansioni temporali dello stesso bersaglio.
  - **Decisione:**
    1. Formula del rischio cumulativo riscritta secondo la teoria dell'affidabilità probabilistica: $P_{\text{combined}} = 1 - \prod (1 - p_i)$, con fattore hop e pavimento a 85 per CVE critiche.
    2. Bound computazionale DFS a 50 percorsi massimi (`maxPaths = 50`) per contrastare l'esplosione esponenziale.
    3. Archi inferiti (`inferred: true`, `confidence`, `inference_basis`) con visualizzazione tratteggiata e badge di confidenza percentuale nel frontend.
- **ADR-08 (Scope Aziendale Gerarchico e CompanyContext Globale React):**
  - **Contesto:** L'utente gestisce molteplici aziende clienti MSSP, ciascuna articolata internamente in diversi reparti con IP distinti. In precedenza mancava una persistenza globale dell'azienda selezionata, con conseguente duplicazione disorientante di menu a discesa perimetrali in ogni pagina.
  - **Decisione:**
    1. Distinzione architetturale netta dei nodi `target_groups`: nodi con `parent_id IS NULL` rappresentano le **Aziende**, mentre i nodi con `parent_id != NULL` rappresentano i **Reparti** interni.
    2. Creazione di `CompanyContext` React globale (persistente in `localStorage`) che calcola ricorsivamente la chiusura dei reparti discendenti e dei target IP perimetrali.
    3. Selettore universale in sidebar desktop e mobile drawer con indicazione dell'azienda attiva o modalità panoramica "Tutte le Aziende".
    4. Nuova rotta e dashboard dedicata `/companies` con card metriche, color tag, operazioni CRUD e attivazione immediata dello scope.
    5. Propagazione automatica del filtraggio e rimozione dei dropdown ridondanti su Dashboard, Targets (solo reparti dell'azienda attiva), Scans, Nuova Scansione, Findings (ispezione per reparto ed IP), Reports e Attack Path.
  - **Conseguenze:** Esperienza utente fluida e lineare; totale coerenza visiva e operativa; eliminazione completa di disallineamenti tra feature.
- **ADR-09 (Roadmap Strategica Milestones 3, 4, 5: Active Defense, Compliance NIS2/DORA e AI Copilot):**
  - **Contesto:** A valle del completamento di Epic 27 e della Milestone 2, il Product Owner ha validato e selezionato tre direttrici evolutive prioritarie per posizionare la piattaforma ai vertici dell'innovazione di settore.
  - **Decisione:**
    1. **Milestone 3 (Active Defense & SOAR Light):** Chiusura del ciclo tra identificazione e risoluzione del rischio con neutralizzazione 1-click dei chokepoints via firewall SSH (con safety rollback), virtual patching (Fail2ban/CrowdSec/ModSecurity) e hub di alerting multi-canale (Slack, Teams, Telegram, Webhooks, SMTP).
    2. **Milestone 4 (Continuous Compliance Frameworks):** Valutazione e scorecard automatizzata per NIS2 (Art. 21), DORA e ISO 27001 con tracciamento SLA MTTR e generazione di dossier audit-ready formali in PDF.
    3. **Milestone 5 (AI Remediation Copilot & Assistant):** Supporto generativo per script di mitigazione su misura per OS/container, sintesi narrative executive e query contestuali "Ask Infrastructure" con isolamento rigoroso dei dati sensibili (Zero Data Leakage).
  - **Conseguenze:** Definizione chiara delle prossime epiche di sviluppo; massima aderenza al valore per il cliente finale; sinergia completa con i moduli preesistenti.
- **ADR-10 (Standard Operativo Software House: Context Layering, Mappa Full-Stack & Quality Gates):**
  - **Contesto:** Con la crescita delle dimensioni del repository e delle specificità funzionali (oltre 40KB di requisiti e molteplici viste frontend), l'agente necessita di un protocollo rigoroso per preservare il budget di token, prevenire allucinazioni, garantire la retrocompatibilità DB e mantenere il 100% di allineamento documentale full-stack.
  - **Decisione:**
    1. **Context Layering:** Strutturazione dell'accesso ai file `.md` su 4 livelli gerarchici (Livello 0 per RAM attiva `PROJECT_STATE.md` + sprint attivo in `3_TASK_LIST.md`, Livello 1 per singolo dominio on-demand, Livello 2 per specifiche globali lette a slice/grep, Livello 3 no-context per manuali utente e legacy).
    2. **Mappa Documenti Full-Stack:** Copertura obbligatoria di tutte le componenti applicative frontend (`frontend/src/pages/**`, `components/**`), backend Go e migrazioni SQL (`migrations/**`).
    3. **Contract-First & DB Rollback:** Sincronizzazione preventiva dei tipi DTO TypeScript/Go prima dell'implementazione e verifica di integrità di applicazione e rollback per ogni migrazione SQL (`up`/`down`).
    4. **Quality Gates & DoD a 4 punti:** Criterio formale per ogni task: (1) Unit/Component PASS, (2) Integration/E2E PASS, (3) Build/Typecheck/Lint PASS (exit code 0), (4) Doc Sync Full-Stack completato.
    5. **Tech Debt Backlog:** Tracciamento sistematico in `3_TASK_LIST.md` del debito tecnico e dei refactoring pianificati.
  - **Conseguenze:** Standard di livello Software House Enterprise; zero token bloat; totale coerenza tra frontend, backend, database e documentazione.
- **ADR-11 (Runtime Agentico Codex e Playbook Portabili):**
  - **Contesto:** La governance v6.2 era stata progettata per Antigravity e assumeva tool e slash command non nativi in Codex (`view_file`, browser subagent, workflow auto-registrati). Questo rendeva alcune istruzioni non eseguibili o ambigue.
  - **Decisione:**
    1. `AGENTS.md` diventa l'unico entry point automaticamente scoperto da Codex, con precedenza delle istruzioni, bootstrap, tutela della working tree e uso capability-aware degli strumenti.
    2. `.agent/workflows/*.md` resta una libreria portabile di playbook richiamata da `.agent/README.md`; gli alias `/feature`, `/verify`, ecc. sono intenti conversazionali, non comandi runtime garantiti.
    3. I workflow non ampliano permessi: push, merge, deploy, tag e azioni distruttive richiedono autorizzazione esplicita. Browser e subagent hanno fallback dichiarati quando indisponibili o non consentiti.
    4. La verifica è evidence-first: working tree preesistente preservata, comandi con exit code reale, controlli Markdown dedicati e distinzione tra PASS, FAIL e NON ESEGUITO.
    5. La governance agentica segue un Doc Sync dedicato senza forzare l'associazione a un dominio applicativo.
  - **Conseguenze:** Istruzioni realmente applicabili da Codex, minori assunzioni implicite, maggiore sicurezza operativa e parità tra comandi dichiarati e playbook disponibili.

- **ADR-12 (2026-10-15 · Documentazione SDLC e orchestrazione per feature):**
  - **Contesto:** Il PO richiede documenti organizzati per fase e ruoli specializzati, conservando memoria operativa e domini.
  - **Decisione:** Adottare `01-product`–`06-devops`, changelog in root, ruoli su richiesta e gate proporzionati per feature; assegnare una fonte autorevole per informazione. Validazione tecnica al Tech Lead, scelte di prodotto al PO. I vecchi percorsi diventano rimandi senza duplicare contenuti.
  - **Conseguenze:** Ricerca mirata, ownership esplicita, passaggi verificabili e migrazione senza rinumerare FR/UC o alterare accettazioni esistenti.

- **ADR-13 (2026-10-15 · Documentazione riservata agli agenti):**
  - **Contesto:** Il PO richiede la rimozione degli screenshot e una cartella `docs/` contenente solo documenti necessari ad `AGENTS.md`.
  - **Decisione:** Rimuovere manuale utente, screenshot e relativi riferimenti operativi; conservare Markdown di prodotto, design, architettura, domini, sviluppo, QA, DevOps, indice e memoria. Salvare evidenze visuali temporanee in `.agent/artifacts/<task-id>/`, escluso da Git.
  - **Conseguenze:** Il verificatore controlla directory ammesse, file root canonici e assenza di artefatti non Markdown. L'onboarding non ricrea archivi o stub dentro `docs/`. Le prove visuali restano possibili fuori dalla documentazione permanente.
