# Test Plan

> Responsabile: QA Engineer. Strategia preventiva; risultati eseguiti e baseline sono in [PROJECT_STATE](../PROJECT_STATE.md).

## Strategia

Per ogni feature collegare FR/UC/storia → criterio osservabile → test → esito. Coprire happy/error/edge, isolamento tenant, autorizzazioni, input ostili e lifecycle risorse nei percorsi pertinenti. Per bug dimostrare riproduzione rossa e test verde; per refactor fissare prima gli invarianti.

| Livello | Comandi e scope | Evidenza richiesta |
|---|---|---|
| Backend | Da `backend`: test package, poi `go test ./...`, `go vet ./...`, `go build ./...` | Exit code e package; indicare skip e prerequisiti reali |
| Frontend | Da `frontend`: test file, poi `npm test`, `npm run build` | Exit code, file/test e typecheck |
| Integrazione/E2E | API, DB di test, migrazioni up/down, flussi browser interessati | Ambiente, dati, output e screenshot reali per UI |
| Documentazione | `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/verify-agent-docs.ps1`, confronto dei contenuti migrati, `git diff --check` | Esito e perimetro; nessuna suite applicativa per soli Markdown e verificatore |

Lo script `scripts/e2e_test.ps1` esiste ma esegue operazioni sulla piattaforma: usarlo solo dopo avere identificato destinazione e credenziali di test. Un test saltato internamente non prova l'integrazione con un servizio reale.

Salvare le evidenze visuali temporanee in `.agent/artifacts/<task-id>/` (ignorato da Git). In `docs/` restano strategia, criteri e risultati testuali, senza screenshot o allegati binari.

## Test case per feature

ID, FR/UC, precondizioni/dati, passi, risultato atteso, livello, comando, esito (PASS/FAIL/NON ESEGUITO) ed evidenza. L'assenza di un ambiente obbligatorio impedisce il gate verified.

## Bug report

ID/task, severità, ambiente/versione, atteso, osservato, passi minimi, evidenza sanificata, test di riproduzione, causa radice, fix e regressione.

## Gate finale

Test pertinenti e regressione verdi; build/typecheck/vet verdi; nessun blocco critico; Doc Sync; review del diff. La review è indipendente quando un secondo agente è disponibile e autorizzato; altrimenti dichiarare self-review e relativi limiti.
