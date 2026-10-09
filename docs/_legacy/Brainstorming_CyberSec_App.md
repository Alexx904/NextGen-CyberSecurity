# NextGen CyberSecurity Platform - Brainstorming & Vision

## 1. Visione del Progetto
Creare una piattaforma di CyberSecurity all-in-one, progettata con standard di ingegneria del software di alto livello (architettura scalabile, microservizi, API-first). L'obiettivo è fornire uno strumento che non sia solo tecnico, ma che traduca i rischi informatici in metriche di business chiare.

## 2. Architettura e Stack Tecnologico
In base alle decisioni progettuali, il sistema avrà un'architettura moderna, portabile e scalabile:
*   **Modello di Distribuzione:** Web-based (SaaS o installazione On-Premise presso il cliente).
*   **Linguaggio Core e Backend:** **Golang (Go)**. Garantisce altissime prestazioni per le scansioni di rete (grazie alle goroutine per la concorrenza) e permette di avere un *singolo codice sorgente* multipiattaforma.
*   **Ambiente di Esecuzione:** **Docker Containers**. L'applicazione sarà "containerizzata", il che significa che girerà in modo identico sia su server Linux che su server Windows, azzerando i problemi di compatibilità ambientale.
*   **Database Principale:** **PostgreSQL**. Offre solidità relazionale per la gestione degli utenti/tenant e flessibilità grazie a `JSONB` per storicizzare i dati non strutturati dei report di scansione.
*   **Frontend:** Single Page Application (Web UI) moderna (es. React, Vue o Angular) per offrire una dashboard reattiva ed elegante.

## 3. Core Feature: Network Vulnerability Scanner
Il motore principale della piattaforma sarà uno scanner avanzato (ispirato a soluzioni come OpenVAS/Nessus):
*   **Discovery e Fingerprinting:** Mappatura automatica degli asset di rete.
*   **Rilevamento Vulnerabilità (CVE & CVSS):** Correlazione con DB aggiornati.
*   **Motore di Reporting Avanzato:** Generazione white-label (PDF/HTML/CSV) con sezioni Executive e Technical.

## 4. Moduli Aggiuntivi (Valore di Mercato)
*   **External Attack Surface Management (EASM):** Monitoraggio perimetro esterno.
*   **Cloud Security Posture Management (CSPM):** Verifica configurazioni AWS/Azure/GCP.
*   **Identity Leak Monitor:** Verifica credenziali compromesse (Dark Web).

## 5. Funzionalità Innovative (Priorità PoC)
Queste sono le feature chiave (scelte come prioritarie per il primo rilascio) che differenzieranno il prodotto dalla concorrenza:
*   **Attack Path Mapping (Grafo delle Vulnerabilità):** Un motore grafico che non mostra solo la singola vulnerabilità, ma crea una catena visuale dell'attacco. (Es: Il DB interno è vulnerabile -> ma è irraggiungibile da fuori -> tuttavia il server Web frontale ha una RCE -> l'attaccante passa dal Web per arrivare al DB).
*   **Continuous Compliance Score:** Traduzione dei risultati tecnici in metriche di compliance manageriali. Dashboard dedicate per **NIS2, DORA e ISO 27001**, mostrando la percentuale di allineamento e quali macchine specifiche bloccano la conformità.
*   *Funzionalità Futura:* AI-Powered Remediation Copilot (Generazione script IaC per patching).

## 6. Prossimi Passi
1.  **Definizione dei Requisiti (Requirements Engineering):**
    *   Requisiti Funzionali (FR)
    *   Requisiti Non Funzionali (NFR - Sicurezza, Scalabilità, Performance)
2.  **System Design:** Disegno dei microservizi, flussi dati e diagrammi C4.
3.  **Setup dell'infrastruttura di sviluppo:** Init del modulo Go, setup Docker Compose con Postgres.
