# E28-S1-T1 · Chokepoint remediation

Stato: implemented (gate verified aperto) · Ultimo allineamento: 2026-10-17 · working tree.

## Obiettivo e riferimenti

Implementare il primo incremento backend di FR-19.1 e FR-19.2: generare una mitigazione da un chokepoint corrente del tenant e applicarla attraverso SSH con rollback remoto a 60 secondi. Dipendenze: domini [attack path](../../domains/attack_path.md) e [SSH/firewall](../../domains/ssh_and_firewall.md).

Non-goal: UI (E28-S1-T2), virtual patching, notifiche, nuove dipendenze, deployment e prove su infrastrutture remote non esplicitamente identificate come laboratorio.

## Decisione di prodotto

Il PO ha scelto il blocco porta sul server del chokepoint; isolamento host da definire separatamente. Implementati preview e remediation per SERVICE IPv4 TCP/UDP, server del tenant con IP coincidente, protezione della porta SSH. Contratto canonico nel [dominio](../../domains/active_defense_and_soar.md).

## Contratti e file previsti

- `backend/internal/domain/defense.go` e `frontend/src/types/defense.ts`: richiesta e risposta tipizzate, modalità ammesse, preview generata dal server, stato timer e identificativo del server SSH.
- `backend/internal/service/defense_service.go`: `Preview` e `Remediate`; rigenerazione del grafo del tenant, verifica che il nodo sia ancora un chokepoint, associazione al server SSH dello stesso tenant, validazione IP/porta/protocollo e comando derivato solo da dati validati.
- `backend/internal/delivery/http/handler/defense_handler.go`: endpoint `POST /api/v1/attack-paths/chokepoints/{id}/remediate`; payload limitato e validato, errori pubblici senza dettagli SSH o credenziali. Preview da definire insieme al contratto finale.
- `backend/cmd/server/main.go`: wiring e autorizzazione `RequireSecOps`, mantenendo autenticazione e isolamento tenant.
- `backend/internal/service/firewall_service.go`: riuso di snapshot e watchdog; correggere prima dell'integrazione le criticità direttamente coinvolte nel nuovo flusso.

Nessuna migrazione DB prevista. Conferma e rollback esplicito riusano gli endpoint firewall esistenti, previa verifica del loro comportamento.

## Rischi emersi dalla lettura del codice

Osservazioni statiche, ancora da riprodurre con test:

- `startSafetyTimer` controlla il timer sotto mutex ma rilascia il lock prima di creare snapshot e watchdog: due richieste possono superare il controllo.
- Il PID remoto non valido diventa zero; `Confirm` costruisce comunque il comando kill e ignora l'esito SSH.
- `Rollback` cancella il watchdog prima di verificare che il ripristino abbia avuto successo.
- Il timer può scadere mentre una mutazione è ancora in corso; definire il comportamento per evitare applicazioni successive al ripristino.
- Il blocco del servizio SSH o un isolamento totale possono interrompere il canale di gestione: il contratto deve esplicitare esclusioni e limiti; nessuna promessa di isolamento completo con una singola regola INPUT.

## Strategia QA e operativa

1. Baseline backend `go test ./...` prima delle modifiche applicative.
2. Test mirati del generatore: TCP/UDP, porta fuori range, protocolli non supportati, IPv4/IPv6, input ostile, nodo non più chokepoint, server assente o di altro tenant, mismatch host/server e porta di gestione.
3. Test del servizio con executor SSH finto: snapshot e watchdog prima della mutazione, nessuna mutazione se preparazione fallisce, timer attivo, errori di applicazione e ripristino.
4. Riproduzione rossa e verde delle criticità firewall affrontate; concorrenza, PID non valido, errori di conferma e rollback. Usare `go test -race` se la toolchain locale lo consente.
5. Test handler su autenticazione/RBAC, JSON malformato, errori di dominio e successo.
6. Regressione: `go test ./...`, `go build ./...`, `go vet ./...` in `backend`; `npm.cmd test`, `npm.cmd run build` in `frontend` dopo l'introduzione del contratto TypeScript.
7. Prova reale forward/rollback solo su host Linux di test identificato: non eseguita in questa fase. Dichiarare il limite se indisponibile.
8. Self hostile review, Doc Sync, verificatore documentale e `git diff --check`.

La cache Go predefinita ha impedito l'avvio della baseline; il tentativo successivo usa `GOCACHE` in `.agent/artifacts/e28-s1-t1/go-cache`, ignorata da Git.

## Doc Sync e gate

Creare `docs/domains/active_defense_and_soar.md` con il contratto confermato. Aggiornare DOMAIN_MAP, API_SPEC, contratto SSH ove modificato, requisiti solo per chiarimenti approvati, test plan, indicazioni operative pertinenti, TASK_LIST e PROJECT_STATE. ADR necessario se cambia il lifecycle condiviso del watchdog o viene introdotto un modello gateway.

- Ready for design: sì, scelta PO ricevuta.
- Ready for implementation: sì; contratti e test predisposti, implementazione effettuata.
- Verified: yes, 2026-10-18; shared watchdog and task-specific real SSH/TCP gate PASS.
- Accepted: no.

## Evidenze e self-review

### Resumed operational gate · 2026-10-18

M11-FIREWALL-01 now verifies the shared lifecycle on nft/legacy and real pinned
OpenSSH. The original PID/local-lock findings below are historical, superseded
by ADR-25. Host reboot/state deletion/other accounts remain explicit limitations.
To verify this task's own orchestration, run a tagged Linux test only inside the
named SSH fixture: deterministic tenant graph fixture + production DefenseService,
SSHService and FirewallService. Start a loopback HTTP endpoint on a synthetic
non-management port, prove reachable → DROP blocks fresh connections → confirmation
from a fresh service → explicit rollback restores reachability. Verify preview,
foreign tenant and stale chokepoint do not mutate rules. Backup repository remains
an in-memory test fixture; this is real SSH/firewall/TCP integration, not a new
PostgreSQL/full-graph E2E claim. No customer-host action or application code change.

Preparation: backend `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go test -c -tags=integration -o ../.agent/artifacts/M11-FIREWALL-01/service-linux.test ./internal/service`
PASS exit 0, 1.428s wall; Docker copy PASS. Inside `nextgen-ssh-lab-20261018`,
explicit TEST_FIREWALL_SSH_CONTAINER/PORT=22/IN_CONTAINER=1:
`/tmp/service-linux.test -test.run=^TestDefenseSSH -test.v -test.count=1`
PASS exit 0, test 3.01s; root Docker command 7.048s wall. Actual HTTP baseline,
block, confirmed block and restored connection all asserted. Final revision matches
baseline, pending marker absent. Test listener closes; fixture rules restored.
Default backend gate inputs unchanged: only a new integration-tagged test added,
excluded from default build; prior All backend suite/vet/build remains applicable.
No new application deploy/schema/source change for this gate. Self-review in first
person, no independent review. Historical gaps below are superseded by this result.

- `go test ./...` baseline PASS exit 0 con cache locale ignorata.
- Riproduzione PID watchdog invalido: test mirato FAIL exit 1 prima del fix; PASS exit 0 dopo.
- `go test ./internal/service ./internal/delivery/http/handler` PASS exit 0.
- `go test ./...`, `go build ./...`, `go vet ./...` PASS exit 0 in backend.
- `npm.cmd test` PASS exit 0 (145 test) dopo riesecuzione autorizzata fuori sandbox; il primo tentativo era bloccato da esbuild/accesso directory superiori.
- `npm.cmd run build` PASS exit 0 (10,94s).
- `go test -race ./internal/service -run 'TestDefense|TestBuildChokepoint|TestFirewall'` non eseguibile, exit 1: CGO richiesto e disabilitato.
- Self hostile review: input della regola derivati solo dal grafo e costanti, verifica tenant/server, RBAC sulla route, errori HTTP senza stderr remoto; lock a dimensione fissa, nessuna nuova goroutine nel servizio defense. Fix applicati a PID, concorrenza preparazione, conferma e ordine rollback. Restano host-key SSH non verificata e possibile esecuzione tardiva rispetto al watchdog, registrati come TECH-DEBT-06/07; il task rimane `[ ]`.
- Nessuna prova SSH reale né test visuale (incremento backend, UI E28-S1-T2 non implementata).
- Il PO conferma l'indisponibilità del laboratorio e richiede di mantenere aperta la verifica remota. Verificatore documentale e `git diff --check` PASS exit 0.
