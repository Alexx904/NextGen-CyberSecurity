---
description: Verifica ripetibile di test, build, typecheck, lint e stato Git
---

# Verify

1. Leggi i comandi dal Project Profile e convalidali contro manifest e CI. Non installare dipendenze o avviare servizi senza necessità/autorizzazione.
2. Registra branch, working tree e baseline nota; non confondere fallimenti preesistenti con regressioni.
3. Definisci scope e rischio secondo `AGENTS.md` sez. 8. Durante lo sviluppo usa test mirati; alla consegna esegui una sola suite completa e build/typecheck/vet per lato coinvolto. Una richiesta `/verify` senza scope include entrambi i lati e PostgreSQL. Usa `scripts/verify.ps1 -Profile Quick|Final|Full|Docs -Scope Backend|Frontend|All`; `Quick` richiede package/file mirati; `Full` impone entrambi i lati e DB. Per `Final` aggiungi `-Integration` se repository/DB sono coinvolti. I controlli documentali sono inclusi nei profili finali. UI reale e rollback migrazioni restano gate aggiuntivi quando pertinenti.
   - Non duplicare mirati e suite completa sullo stesso codice solo per formalità; riutilizza i mirati già verdi se ancora validi.
   - Non riusare risultati di un altro ambiente come prova locale. Non avviare DB applicativi o caricare `.env`; configurare `TEST_DB_*` per un database dedicato locale.
   - Job backend, frontend e integrazioni isolate possono procedere in parallelo; non eseguire migrazioni concorrenti sullo stesso DB.
4. Per ogni comando riporta directory, exit code, durata e conteggio test quando disponibile. Un comando non eseguito è `NON ESEGUITO`, non PASS.
5. In caso di errore diagnostica la causa; correggi solo se il PO ha chiesto una modifica, altrimenti produci un report evidence-first.
6. Aggiorna la baseline in `PROJECT_STATE.md` solo se la verifica è completa e rappresentativa.
