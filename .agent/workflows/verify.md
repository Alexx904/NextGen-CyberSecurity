---
description: Verifica ripetibile di test, build, typecheck, lint e stato Git
---

# Verify

1. Leggi i comandi dal Project Profile e convalidali contro manifest e CI. Non installare dipendenze o avviare servizi senza necessità/autorizzazione.
2. Registra branch, working tree e baseline nota; non confondere fallimenti preesistenti con regressioni.
3. Esegui nell'ordine: test mirati del cambiamento, suite backend, build/vet backend, suite frontend, build/typecheck frontend e controlli documentali.
4. Per ogni comando riporta directory, exit code, durata e conteggio test quando disponibile. Un comando non eseguito è `NON ESEGUITO`, non PASS.
5. In caso di errore diagnostica la causa; correggi solo se il PO ha chiesto una modifica, altrimenti produci un report evidence-first.
6. Aggiorna la baseline in `PROJECT_STATE.md` solo se la verifica è completa e rappresentativa.
