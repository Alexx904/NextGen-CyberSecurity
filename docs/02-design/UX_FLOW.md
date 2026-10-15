# UX Flow

> Responsabile: UI/UX Designer. Navigazione verificata in `frontend/src/App.tsx`; flussi di dominio derivati dalle specifiche esistenti.

```mermaid
flowchart TD
    Login["/login"] --> Auth{"Sessione valida"}
    Auth -->|Sì| Dashboard["/ · Dashboard"]
    Auth -->|No| Login
    Dashboard --> Companies["/companies · Aziende"]
    Companies --> Targets["/targets · Reparti e target"]
    Dashboard --> Discovery["/network-discovery"]
    Discovery --> Targets
    Targets --> Scans["/scans"]
    Scans --> Detail["/scans/:id"]
    Detail --> Findings["/findings"]
    Findings --> Attack["/attack-path"]
    Detail --> Reports["/reports"]
    Dashboard --> SSH["/ssh-servers"]
    Dashboard --> Settings["/settings"]
```

Le frecce dopo il login descrivono il percorso operativo, non redirect automatici o link obbligatori. Il selettore azienda determina il contesto trasversale.

Per modifiche UX aggiornare il solo flusso interessato, includendo empty/loading/error, autorizzazioni, conferme e annullamento. Verificare il risultato con browser e screenshot quando disponibile; il diagramma non sostituisce la prova visuale.
