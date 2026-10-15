# Use Cases

> Responsabile: Business Analyst. Migrazione documentale: 2026-10-15. Gli stati funzionali restano quelli di [PROJECT_STATE](../PROJECT_STATE.md).

### UC-01: Configurazione Target e Policy
- **Attore:** SecOps Engineer.
- **Flusso:** L'operatore accede alla sezione Target, inserisce un IP/CIDR/range, verifica che non sia in blacklist e lo assegna a una cartella nell'albero.

### UC-02: Esecuzione Scansione On-Demand con Esclusione Host Inattivi
- **Attore:** SecOps Engineer.
- **Flusso:** L'operatore seleziona un range IP. Il backend esegue la scansione, scarta a monte gli host down, memorizza solo gli host reattivi, correla i servizi alle CVE e aggiorna la dashboard.

### UC-03: Pianificazione Ricorrente su Cartella Asset
- **Attore:** SecOps Engineer.
- **Flusso:** L'operatore programma una scansione settimanale su una cartella dell'albero di asset. Lo scheduler esegue il job ricorsivo su tutti i nodi appartenenti alla cartella.

### UC-04: Analisi Findings e Triage Vulnerabilità
- **Attore:** SecOps Engineer / Auditor.
- **Flusso:** Consultazione della matrice vulnerabilità con filtri per CVSS, visualizzazione evidenze e note di remediation.

### UC-05: Personalizzazione Branding Tenant (White-Label)
- **Attore:** Tenant Admin.
- **Flusso:** Caricamento logo, selezione colori, intestazioni e note legali in Settings con anteprima live.

### UC-06: Gestione Centralizzata Report dall'Hub `/reports`
- **Attore:** SecOps Engineer / Auditor.
- **Flusso:** L'operatore accede a `/reports`, filtra le scansioni storiche, seleziona il template desiderato (Executive o Tecnico) e scarica direttamente il file PDF o l'export CSV.

### UC-07: Organizzazione Gerarchica Asset tramite Tree Explorer
- **Attore:** SecOps Engineer.
- **Flusso:** L'operatore crea cartelle e sotto-cartelle nell'albero, organizza i target per sede geografica o perimetro logico, e monitora gli indicatori di rischio aggregati per cartella.

### UC-08: Rilevamento Rapido Dispositivi LAN (Network Discovery)
- **Attore:** SecOps Engineer / Sysadmin.
- **Flusso:** L'utente accede alla sezione Network Discovery, inserisce la subnet locale, avvia la scansione e visualizza in pochi secondi tutti i dispositivi online completi di MAC, vendor e hostname. Seleziona i dispositivi rilevati e li aggiunge a una cartella di Target con un click.

### UC-09: Gestione Regole Firewall iptables da Interfaccia Grafica
- **Attore:** SecOps / Sysadmin.
- **Flusso:** L'utente seleziona un server Linux censito, visualizza le catene iptables attive, aggiunge una regola di blocco o apertura porta via form grafico, attiva il safety timer e applica la configurazione verificando l'assenza di lockout.

### UC-10: Accesso a Terminale Web SSH Remoto
- **Attore:** Sysadmin / SecOps.
- **Flusso:** L'utente apre la sessione di terminale web direttamente dal browser per eseguire verifiche diagnostiche su un server Linux registrato con credenziali sicure.

### UC-11: Monitoraggio in Tempo Reale della Scansione e Ricognizione LAN
- **Attore:** SecOps Engineer / Sysadmin.
- **Flusso:** L'utente avvia una scansione LAN o di vulnerabilità. Senza attendere il termine del processo, visualizza in tempo reale la progressione percentuale e i singoli dispositivi o porte che compaiono istantaneamente nella griglia a schermo. Se la pagina viene ricaricata, i dati parziali già salvati su DB vengono ricaricati e il flusso di streaming si riallaccia.

### UC-12: Debugging Interattivo, Triage Errori e Feedback Loop con Agente AI
- **Attore:** Sviluppatore / Tester / Sysadmin.
- **Flusso:** Durante l'esecuzione di qualsiasi operazione nella piattaforma (scansione, discovery, firewall, login), l'utente nota un'anomalia o un badge rosso di errore. Apre il pannello laterale (click sul pulsante fluttuante o `Ctrl + Shift + D`), seleziona il filtro `ERROR`, ispeziona la richiesta fallita con status code e corpo risposta, clicca "Copia per AI (Markdown)" e incolla il blocco diagnostico nella chat con l'agente. L'agente analizza la causa radice ed emette il fix immediato.

### UC-13: Esplorazione Visiva Catene di Attacco & Remediation dei Chokepoint (Milestone 2)
- **Attore:** SecOps Engineer / CISO.
- **Flusso:** L'utente accede alla schermata `/attack-path`. Il sistema elabora l'inventario dei target attivi, le porte esposte e le vulnerabilità correlate, costruendo la mappa topologica delle minacce. L'utente visualizza visivamente i percorsi d'attacco con il cammino critico (*Critical Path*) evidenziato dal perimetro esterno agli host interni. Cliccando su un percorso o su un asset bersaglio (es. database interno o server sensibile), il pannello laterale evidenzia il *Chokepoint* principale (es. porta o servizio la cui chiusura/remediation interrompe il maggior numero di catene di compromissione), fornendo una raccomandazione di remediation a massimo impatto e ROI di sicurezza.

### UC-14: Audit di Sicurezza Presso Diverse Aziende con lo Stesso Computer (Workstation Multi-Company Workflow)
- **Attore:** Security Consultant / Auditor / Sysadmin.
- **Flusso:** L'operatore si reca presso diverse aziende o filiali con il medesimo computer portatile. Utilizza il Network Discovery per scoprire velocemente la LAN locale e importa gli host attivi in una cartella Scope dedicata al cliente (es. "Azienda Alfa", "Azienda Beta"). Nelle sezioni *Scansioni*, *Vulnerabilità* e *Report*, naviga la nuova interfaccia raggruppata per cartella: visualizza i gruppi separati con metriche chiare, apre la cartella dell'azienda desiderata e ispeziona nello specifico ciascun IP con le relative scansioni, vulnerabilità rilevate e report white-label pronti al download, mantenendo il perimetro dei diversi clienti rigorosamente isolato e visivamente ordinato.

### UC-15: Scansione Specialistica e Personalizzata di Vulnerabilità
- **Attore:** Penetration Tester / Security Engineer / IT Auditor.
- **Flusso:** L'utente apre il modale *Nuova Scansione*. Seleziona il target (o la cartella dell'azienda). Può scegliere la modalità **Semplice / Guidata** per selezionare con 1 click un profilo preimpostato categorizzato in base all'obiettivo dell'audit (Pen Test, Falle di Sistema, Web App, Discovery, Compliance) con spiegazione chiara di impatto e tempo stimato; oppure passare alla modalità **Professionale** per personalizzare porte specifiche, livello di stealth, OS detection e flag addizionali, visualizzando l'anteprima del comando e salvando facoltativamente il profilo custom. La scansione viene validata ed eseguita in modo sicuro dal backend.

### UC-17: Generazione e Presentazione Report PDF Direzionale e Tecnico per Clienti
- **Attore:** CISO / Security Consultant / Auditor.
- **Flusso:** L'operatore genera i report dall'Hub `/reports` dopo aver completato l'audit su una cartella/azienda. L'Executive Summary viene stampato o inviato in formato A4 di 2 pagine, perfettamente leggibile, con colori conformi alle direttive di sicurezza e testo di raccomandazione completo senza troncamenti. Il Technical Report presenta tutte le evidenze host per host distribuite uniformemente senza spazi vuoti e con i nomi software corretti.

### UC-18: Neutralizzazione Attiva di un Chokepoint dall'Attack Path
- **Attore:** SecOps Engineer.
- **Flusso:** L'operatore visualizza il grafo di attacco su `/attack-path`. Nota che l'host intermedio espone una porta SSH o un servizio pivot che permette all'attaccante di accedere al DB. Clicca su "Neutralizza Chokepoint": la piattaforma propone una regola `iptables -I INPUT -p tcp --dport <port> -s <untrusted> -j DROP`, si connette in SSH al server e applica la regola con timer di rollback a 60s. L'operatore convalida la connettività e conferma la regola; il grafo si ricalcola istantaneamente mostrando l'interruzione della catena.

### UC-19: Audit di Conformità NIS2 per Cliente Aziendale
- **Attore:** CISO / IT Compliance Manager.
- **Flusso:** L'operatore seleziona l'azienda cliente nel selettore globale e accede alla sezione `/compliance`. Visualizza lo stato di conformità alla NIS2 (es. 82%). Individua un gap critico sull'Art. 21 dovuto a protocolli non cifrati (Telnet/HTTP) nel Reparto Produzione. Con un click genera l'Audit-Ready PDF Dossier che documenta la gap analysis e le azioni correttive pianificate per la presentazione al comitato di audit.

### UC-20: Generazione Playbook di Remediation Assistito da AI
- **Attore:** Sysadmin / Security Analyst.
- **Flusso:** Nella schermata Vulnerabilità o Dettaglio Scansione, l'utente seleziona una CVE critica (es. RCE su OpenSSH). Clicca su "Genera Playbook AI": il copilot analizza la versione esatta del sistema operativo e fornisce la procedura di fix step-by-step con comando di patching, configurazione `sshd_config` raccomandata e test di verifica del servizio post-aggiornamento.

---

## Completamento progressivo

Gli UC conservano gli identificativi e il dettaglio storico. Per ogni UC modificato esplicitare: attore, precondizioni, trigger, happy path, errori/eccezioni, postcondizioni e riferimenti FR/test. L'assenza di UC-16 è preesistente: non rinumerare gli altri UC.
