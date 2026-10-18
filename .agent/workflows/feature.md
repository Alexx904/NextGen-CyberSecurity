---
description: Implementa una capacità nello scope autorizzato fino alla consegna verificata
---

# Feature

1. Definisci obiettivo, attori, non-goal, happy/error/edge e criteri osservabili dal PO; confronta requisiti e dominio pertinenti. Chiedi solo decisioni bloccanti secondo AGENTS.
2. Applica [orchestration](orchestration.md). Prima del codice definisci contratti reali API/schema/tipi, compatibilità, UX e rischi QA/operativi. Documenti già validi soddisfano il gate; aggiorna HLD/ADR solo se cambia una decisione architetturale.
3. Registra task con dipendenze e DoD; piano separato solo per lavoro complesso. Dichiarare file, flusso, rischi e comandi è sufficiente per auto-approvare scelte interne reversibili. Rispetta working tree e preferenza Git del PO.
4. Implementa in ordine di dipendenza con test pertinenti; per DB applica [postgres-migration](../../.agents/skills/postgres-migration/SKILL.md). Dichiara variazioni sostanziali al piano.
5. Completa il gate di AGENTS sez. 8 sul codice integrato: [repo-verification](../../.agents/skills/repo-verification/SKILL.md). I sotto-task condividono il gate finale; UI e rollback richiedono prove reali quando pertinenti.
6. Esegui self-review del diff e [Doc Sync](../../.agents/skills/documentation-sync/SKILL.md). Approfondisci con [security-review](../../.agents/skills/security-review/SKILL.md) se rischio o richiesta lo giustificano.
7. Consegna secondo AGENTS sez. 10: `[~]` solo con gate obbligatori completati, altrimenti `[ ]`/`[!]` con evidenza. Aggiorna stato persistente e attese senza duplicare il report QA. Commit solo su richiesta tramite [commit](commit.md).
