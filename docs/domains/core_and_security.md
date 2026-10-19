# Core, authentication and shared transport

> Publication scope 2026-10-18: alert-channel / outbox / migration 000012 sections describe deferred T4/T5 design and local prototype evidence. Their code and routes are excluded from this release; published migration is 000011. Alert-specific commands are future gates.


Ultimo allineamento: 2026-10-18 · working tree · AGENT-HARNESS-01.
Ownership trasversale ricostruita dai percorsi esistenti; questa revisione documentale non attesta nuovi test di autenticazione o sicurezza.

## Confini e fonti

TECH-DEBT-09 · 2026-10-18 · working tree: shared CSS uses Tailwind 4/PostCSS;
custom theme and v3 control/border defaults are retained, and renamed focus,
shadow, blur and shrink utilities are migrated across consuming domains.
Authentication/tenant logic is unchanged. Toolchain/browser requirements and
verification limits are owned by [Frontend Setup](../04-development/FRONTEND_SETUP.md#dependency-baseline-2026-10-18)
and the [dependency assessment](../05-qa/TECH-DEBT-09.md).

- Wiring: `backend/cmd/server/main.go`, `frontend/src/App.tsx`, entrypoint e layout. Se collega domini, consultare i rispettivi contratti.
- Auth/tenant: `backend/pkg/auth/`, middleware HTTP, domain user/tenant, auth/tenant service e handler; frontend AuthContext, ProtectedRoute, client API, tipi auth/tenant, Login/Settings. Intento RBAC in [PRD](../01-product/PRD.md); interfacce in [API_SPEC](../03-architecture/LLD/API_SPEC.md). Scope aziendale è distinto dal tenant e documentato nel [dominio assets](reporting_and_assets.md).
- Configurazione: `backend/internal/config/`; inventario sicuro dei nomi in `.env.example`, senza leggere valori reali. Cifratura SSH usa `backend/pkg/crypto/`; gestione fiducia, credenziali e deadline resta nel [dominio SSH](ssh_and_firewall.md).
- Eventi/errori: domain event/errors e service event_broadcaster, response JSON, hook useSSE e test realtime condivisi. SSE/scansioni in [scanning](scanning_and_discovery.md); diagnostica e redazione nel [debug console](debug_console.md).

## Invarianti da verificare quando modificati

E28-S1-T4/T5 · 2026-10-18 · working tree: server wires a cancellable alert worker
after migrations and joins shutdown before DB close. Admin-only alert routes
require tenant context and strict bounded JSON. Settings mounts the channel card;
non-admin users do not fetch channel configuration metadata. Destinations and
write-only secrets remain owned by [Active Defense](active_defense_and_soar.md#alerting-hub-e28-s1-t4t5-in-progress).
Existing target-group repository maps the alert root-company scope constraint
to a validation failure; moving an in-use company cannot silently reroute alerts.

Preservare autenticazione/ruoli e isolamento tenant prima di accesso ai dati; valutare anche consumatori REST/SSE/WS. Ticket e connessioni hanno lifetime/cancellazione propri: non inferire dal test HTTP la validità del terminale. Errori e log non devono esporre segreti. Credenziali di test e DB applicativi non sono intercambiabili.

## Verifica e Doc Sync

M11-CLOSE-01 · 2026-10-18 · working tree: `CompanyProvider` è montato nel layout
protetto dopo l'autenticazione. Il login non invia più richieste inventario
premature; al login vengono caricati gruppi/target senza richiedere refresh.
Il provider viene smontato al logout. Test `frontend/src/App.test.tsx` rosso→verde
con regressione CompanyContext; [evidenza](../05-qa/M11_CLOSURE_QA.md#browser-e-pdf--2026-10-18).

Auth/tenant/config comuni richiedono entrambi i lati secondo [AGENTS](../../AGENTS.md); repository/schema aggiungono PostgreSQL. Test rilevanti nei package auth/crypto/middleware/service e nei test client/route; scegliere i casi dai comportamenti cambiati, non dall'ownership soltanto. Nessun nuovo PASS applicativo attribuito alla mappatura.

TECH-DEBT-05 · 2026-10-19 · working tree: scan authorization actor/tenant are derived
from authenticated context, never body identity/time. Schedule/profile repositories
are wired before starting workers; live consent checks supplement existing RBAC
and tenant/scope rules. JWT and role contracts are unchanged.
[Scanning contract and evidence](scanning_and_discovery.md#3-impatto-e-autorizzazione).
