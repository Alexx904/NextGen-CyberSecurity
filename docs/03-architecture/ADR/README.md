# Architecture Decision Records

> Responsabile: Software Architect. ADR storici preservati; nuove decisioni significative ricevono ID incrementale, contesto, decisione e conseguenze.

## 7. Architecture Decision Records (ADR)

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
