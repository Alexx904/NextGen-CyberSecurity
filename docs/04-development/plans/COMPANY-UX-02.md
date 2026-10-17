# COMPANY-UX-02 · Navigazione e organizzazione aziendale

## Incremento COMPANY-PDF-03 · 2026-10-17

Richiesta PO: separare le CVE per reparto nei PDF dell'intera azienda quando sono presenti reparti. Contratto interno `PDFReportData.GroupFindingsByDepartment` e `DepartmentReportSection.Findings`; API pubbliche e schema invariati. Service attribuisce per target/scan e identità gruppo, conserva target aziendali diretti, reparti vuoti e provenienza mancante senza mescolare IP coincidenti. Generatore conserva KPI consolidati, disegna sezioni per reparto (tecnico completo, Executive top cinque locali con totale), ripete intestazioni sulle continuazioni. Singola scansione/reparto e azienda senza reparti: flusso precedente.

Rischio locale backend, nessuna modifica frontend/repository/migrazione/dipendenza. Test mirati service/generatore con IP uguali, nomi coincidenti con etichette riservate, reparti vuoti, paginazione e compatibilità; gate finale `scripts/verify.ps1 -Profile Final -Scope Backend`. Test DB non richiesti da questo incremento: nessuna query o contratto repository cambiato. Verifica visuale PDF resta necessaria, renderer indisponibile nel runtime.

Richiesta PO 2026-10-17: unire dashboard e gestione aziende; reparti per organizzazione target/scansioni, report scaricabili per reparto o azienda con IP separati nel documento; Attack Path senza filtro CVSS e con scope reparto/azienda; bloccare navigazione globale fuori dashboard; gerarchia massima Azienda → Reparto.

## Contratti e decisioni prima del codice

- Route canonica `/`: dashboard con CRUD aziende e metriche esistenti. `/companies` e `/dashboard` diventano redirect compatibili. Nessuna seconda voce Gestione Aziende.
- Se nessuna azienda valida è selezionata, guard client impedisce montaggio di route operative e redirige `/`; sidebar disabilitata, dashboard e logout disponibili. Eccezione PO: `/ssh-servers` indipendente, accessibile anche globalmente dalla sezione Tool. Il guard attende CompanyContext e gestisce selezioni persistite non valide. È UX, non nuova autorizzazione API.
- Reparti figli diretti della radice; API create/update rifiuta padre già reparto e spostamento di un'azienda con figli sotto altra azienda. Migrazione `000009` aggiunge trigger sulle nuove scritture e serializzazione per tenant delle assegnazioni parent. Nessuna riscrittura dati. Verifica locale read-only: 0 sotto-reparti esistenti. Lettura legacy preservata; nessun dato eliminato.
- Attack Path chiama la generazione senza soglia CVSS; backend mantiene parametro compatibile. Selettore locale reparto usa target e valori del reparto, scope applicato a canvas/percorsi/KPI/chokepoint. Intera azienda mantiene le ipotesi fra reparti, reparto limita la proiezione.
- Report: endpoint e template esistenti invariati; contratto interno Go `PDFReportData.Departments []DepartmentReportSection` con nome reparto e target/IP associati, inclusi quelli senza finding. Allegato/sezione multipagina nei PDF executive e technical; nessun vincolo di due pagine sul dossier aziendale ampliato. CSV aziendale con colonna reparto additiva; CSV singola scansione invariato. Associazioni tramite target/scan e gruppi del tenant, non solo corrispondenza di IP potenzialmente duplicati.
- Scansioni mantengono scelta per reparto/azienda e filtro esistenti; verificare che non sia possibile scegliere altri tenant/aziende dalla modale.

## Implementazione e test

Integrazione confermata dal PO: `client_name` autonomo e opzionale sui server SSH, senza FK aziendale. Create stringa; update puntatore (omesso conserva, vuoto azzera), max 160 caratteri. Filtro e raggruppamento interni. Migrazione `000008` up/down con default vuoto; rollback perde le sole etichette, da esportare prima in uso reale. Test su database dedicato `nextgen_company_ux02_test_20261017`, non su dati applicativi.

App/router/layout, Dashboard e CompanyModal; TreeExplorer/GroupModal e service gruppi; pagina AttackPath e guida/scope; service report e generatori PDF/CSV; test happy/error/edge. Preservare tutte le modifiche AP-REVIEW-01 preesistenti, senza accettarle implicitamente.

Test mirati: route dirette e cambio scope globale, CRUD integrato, parent reparto rifiutato anche API/DB, spostamento con figli, reparto senza target senza fallback globale, richiesta AP senza soglia, sezioni report/CSV con IP corretti e senza finding, persistenza cliente SSH e update compatibile. Regressione Go/Vitest, build/vet, migrazioni up/down su DB dedicato, Doc Sync e self-review. Verifica PDF renderizzata se strumenti disponibili; browser da rivalutare, mai dichiarare screenshot inesistenti. Nessuna nuova dipendenza, commit, push o deploy richiesti.
