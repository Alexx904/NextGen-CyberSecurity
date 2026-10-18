# Core, authentication and shared transport

Ultimo allineamento: 2026-10-18 · working tree · AGENT-HARNESS-01.
Ownership trasversale ricostruita dai percorsi esistenti; questa revisione documentale non attesta nuovi test di autenticazione o sicurezza.

## Confini e fonti

- Wiring: `backend/cmd/server/main.go`, `frontend/src/App.tsx`, entrypoint e layout. Se collega domini, consultare i rispettivi contratti.
- Auth/tenant: `backend/pkg/auth/`, middleware HTTP, domain user/tenant, auth/tenant service e handler; frontend AuthContext, ProtectedRoute, client API, tipi auth/tenant, Login/Settings. Intento RBAC in [PRD](../01-product/PRD.md); interfacce in [API_SPEC](../03-architecture/LLD/API_SPEC.md). Scope aziendale è distinto dal tenant e documentato nel [dominio assets](reporting_and_assets.md).
- Configurazione: `backend/internal/config/`; inventario sicuro dei nomi in `.env.example`, senza leggere valori reali. Cifratura SSH usa `backend/pkg/crypto/`; gestione fiducia, credenziali e deadline resta nel [dominio SSH](ssh_and_firewall.md).
- Eventi/errori: domain event/errors e service event_broadcaster, response JSON, hook useSSE e test realtime condivisi. SSE/scansioni in [scanning](scanning_and_discovery.md); diagnostica e redazione nel [debug console](debug_console.md).

## Invarianti da verificare quando modificati

Preservare autenticazione/ruoli e isolamento tenant prima di accesso ai dati; valutare anche consumatori REST/SSE/WS. Ticket e connessioni hanno lifetime/cancellazione propri: non inferire dal test HTTP la validità del terminale. Errori e log non devono esporre segreti. Credenziali di test e DB applicativi non sono intercambiabili.

## Verifica e Doc Sync

Auth/tenant/config comuni richiedono entrambi i lati secondo [AGENTS](../../AGENTS.md); repository/schema aggiungono PostgreSQL. Test rilevanti nei package auth/crypto/middleware/service e nei test client/route; scegliere i casi dai comportamenti cambiati, non dall'ownership soltanto. Nessun nuovo PASS applicativo attribuito alla mappatura.
