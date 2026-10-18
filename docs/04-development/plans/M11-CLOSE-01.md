# M11-CLOSE-01 · Chiusura Milestone 1.1

Ultimo allineamento: 2026-10-18 · working tree.

Richiesta PO: completare M1.1 mantenendo le milestone approvate. Nessuna
funzionalità M3/M4/M5 viene inclusa implicitamente. Working tree iniziale pulita.

## Contratto e primo incremento

FR-07..FR-13 restano il perimetro: alive-only, report, gerarchia aggiornata
azienda/reparto, discovery, SSH/firewall, SSE e diagnostica. Le accettazioni
precedenti non sostituiscono i collaudi mancanti.

La verifica del firewall ha individuato input shell non protetti e la possibilità
di disabilitare il safety timer contro il contratto del dominio. Il primo
incremento corregge questi difetti senza nuove dipendenze o migrazioni:

- Source/destination: solo IPv4 letterali o CIDR IPv4, oppure vuoto; nessun DNS
  o IPv6 nel comando iptables IPv4.
- Interfacce: vuoto, `*` (nessun filtro), o nomi di massimo 15 caratteri ASCII
  alfanumerici con `_`, `.`, `-`, `:` e suffisso wildcard `+` opzionale.
- Commento: massimo 255 byte UTF-8, senza caratteri di controllo, preservato
  tramite quoting shell di un singolo argomento.
- AddRule/DeleteRule attivano sempre il timer anche se il vecchio flag è false;
  campo API mantenuto per compatibilità. Timeout 0 usa 60s, ammessi 10..300s.
- Rifiuto degli input prima di backup o esecuzione SSH.

## Verifica e dipendenze residue

Rischio: backend e UI firewall; test rossi su injection, timer e timeout,
poi regressione service/domain, suite finale All con build/vet. Nessun DB
modificato da questo incremento. Collaudo Linux isolato necessario per watchdog,
disconnessione e coordinamento multiistanza (TECH-DEBT-07); non usare target reali.
Verifica browser con screenshot se il runtime espone un browser.

Prima della chiusura globale serve una matrice FR-07..FR-13 con evidenze reali,
gap residui e accettazione PO. Non promuovere la milestone per il solo fix firewall.

Matrice canonica della ricognizione: [QA M1.1](../../05-qa/M11_CLOSURE_QA.md).

## M11-REPORT-01 · Contratto UI · 2026-10-18

Decisione PO: firewall iptables/iptables-nft, non nftables nativo. Roadmap invariata.
Report: nessuna nuova API/SQL. Caricare sequenzialmente tutte le pagine dello storico
prima di aggregazioni e filtri; dati incompleti o pagine duplicate sono errore
visibile, non risultati completi. Richieste superate non aggiornano la UI.
Filtri locali: giorni UTC inclusivi sulla data completamento (fallback creazione),
profilo, target, cartella esatta; cartella include solo la voce selezionata mentre
lo scope aziendale include i reparti. Ricerca/severità esatta coerenti nelle viste.
Periodo invertito: errore visibile e nessun risultato. Vista cronologica: 10 per
pagina, navigazione esplicita e reset pagina al cambio filtro/scope. Opzioni dei
filtri ricavate dallo storico completo nello scope corrente. I download consolidati
continuano a riguardare l'intera azienda/reparto, con testo esplicito vicino ai filtri.
Test: pagine oltre la prima, errore parziale, filtri combinati, date limite,
isolamento scope, navigazione; gate frontend unico. Browser runtime da verificare.
