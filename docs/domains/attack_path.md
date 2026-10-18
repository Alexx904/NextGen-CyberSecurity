# Dominio: Attack Path Mapping & Threat Chaining

> Ultimo allineamento: 2026-10-17 · working tree · AP-REVIEW-01.
> Fonti: `backend/internal/domain/attack_path.go`, servizio e handler omonimi; `frontend/src/types/attackPath.ts`, API, pagina e componenti in `components/graph/`.
> [Piano e finding](../04-development/plans/AP-REVIEW-01.md); [report QA](../05-qa/ATTACK_PATH_REVIEW.md).

## Scopo e limite di validazione

E28-S1-T2 (2026-10-18, working tree) adds an operational UI only for current
SERVICE chokepoints: explicit matching SSH server, backend preview and impact
acknowledgement, followed by the existing firewall safety manager. Auditor stays
read-only, unsupported nodes disabled. No client-generated arbitrary rule, automatic
server selection or automatic confirmation. Model risk/path counts are not reduced
merely because a firewall rule was applied. Contract in
[Active Defense](active_defense_and_soar.md); browser gate pending.

Analisi euristica per ordinare verifiche e individuare nodi comuni ai percorsi modellati. Non prova accessibilità da Internet, sfruttabilità, RCE, credenziali condivise o lateral movement reale. Un inventario SSH non descrive trust fra host. Il modello non contiene routing, ACL, identità/privilegi né vettori CVSS: non deve presentarli come conoscenze acquisite.

[MITRE TA0008](https://attack.mitre.org/tactics/TA0008/) descrive il movimento e controllo di sistemi remoti; il modello attuale genera ipotesi, non verifica quelle capacità. [FIRST](https://www.first.org/cvss/user-guide.html) distingue severità CVSS e rischio.

## Dati e copertura

- Lettura paginata per tenant: fino a 2.000 target e 10.000 finding, pagine 200/500. Ordinamenti repository con ID come spareggio stabile. Non è uno snapshot transazionale durante scansioni concorrenti.
- Host da target IP/FQDN e finding. CIDR e range non sono espansi in host inesistenti: sono utili per lo scope dei finding già osservati.
- Servizi distinti per host, porta e protocollo TCP/UDP. Finding chiusi/filtrati espliciti esclusi; stato assente significa osservazione storica senza conferma attuale.
- Per endpoint si considera lo scan con osservazione più recente; duplicati CVE dello stesso scan sono accorpati. Una porta assente da una nuova scansione non cancella automaticamente osservazioni vecchie: occorre verificare età, profilo e copertura.
- Metadati dei nodi CVE: CVE, host, metodo e confidenza del finding, scan ID e data. Nessun banner o segreto SSH caricato.
- Al massimo 50 percorsi: prima le associazioni CVE ordinate per CVSS decrescente, poi le ipotesi laterali. Limite prima della proiezione aziendale. Grafo e score non rappresentano tutto il rischio quando troncati.

## Semantica del grafo

Tipi storici preservati: ENTRYPOINT, HOST, SERVICE, VULNERABILITY, SUBNET; archi NETWORK_ACCESS, EXPLOITS, PIVOTS_TO, TRUSTED_RELATION. SUBNET e TRUSTED_RELATION non vengono inventati da questo generatore.

- Radice con ID storico `entry-internet`, etichetta **Origine di analisi (accesso ipotetico)**. Non implica Internet.
- Radice → host: ipotesi di accesso da verificare, `inferred=true`.
- Host → servizio: associazione nei finding, non prova della raggiungibilità da altre macchine.
- Servizio → CVE: correlazione del finding, `inferred=true`, non esecuzione di un exploit.
- CVE con CVSS ≥ 7 → altro host nella stessa /24 IPv4 convenzionale: ipotesi laterale debole. Non assume la maschera reale, RCE o permessi. Nessun arco arbitrario fra subnet diverse o derivato dall'inventario SSH. FQDN/IPv6 non danno inferenze di vicinanza.
- Nessuna percentuale di confidenza di rete calibrata disponibile: il campo storico rimane compatibile, ma la UI non mostra percentuali arbitrarie.

Ogni percorso contiene nodi esistenti ed esattamente un arco orientato per ogni coppia consecutiva: `len(edge_ids) = len(node_ids)-1 = hop_count`. I passaggi sono relazioni logiche, non hop di routing. La validazione del dominio controlla adiacenza e ID degli archi.

## Indice euristico e chokepoint

Il campo API storico `cumulative_risk` resta un indice 0–100 per compatibilità. Non è probabilità, EPSS, rischio economico o un CVSS composto ufficiale. Formula legacy effettiva:

- Per ciascuna vulnerabilità con CVSS positivo: `w = min(0.95, max(0.05, CVSS/10 × 0.85))`.
- Base: `100 × (1 - product(1-w))`; senza vulnerabilità, base 15.
- Moltiplicatore: `1 + 0.5/sqrt(max(1, passaggi))`.
- Minimo 85 con CVSS ≥ 9; altrimenti minimo 65 con CVSS ≥ 7; massimo 100, una cifra decimale.
- Severità dell'indice: CRITICAL ≥ 75, HIGH ≥ 55, MEDIUM ≥ 35, LOW altrimenti. Non confondere con severità CVSS dei nodi (9/7/4).

Chokepoint: conteggio dei percorsi contenenti ciascun nodo intermedio e somma dei loro indici, con ordine deterministico per indice, conteggio e ID. La somma può superare 100; non è una riduzione di rischio misurata. Più chokepoint possono coprire gli stessi percorsi: non sommare fra nodi. Raccomandazioni da valutare per impatto operativo, non garanzia di neutralizzazione.

## Contratto API

`GET /api/v1/attack-paths/graph?min_cvss=7`: autenticazione JWT e scope tenant. `min_cvss` finito in [0,10], altrimenti 400; assente=0. Filtra i nodi CVE e i relativi percorsi, mantiene il contesto host/servizi. Errori: 401 senza tenant, 500 su errore di generazione. `GET /chokepoints` rigenera lo stesso modello con soglia 0.

Payload esistente preservato; aggiunta `analysis`:
```json
{
  "model": "heuristic-review-v2",
  "warnings": ["Accesso e sfruttabilità da verificare"],
  "paths_truncated": false,
  "inventory_truncated": false,
  "target_count": 17,
  "finding_count": 20,
  "max_paths": 50
}
```

ID nodi generati includono protocollo e componenti codificate; sono identificatori dello snapshot, non riferimenti persistenti. Dopo rigenerazione la selezione viene riallineata. La remediation [Active Defense](active_defense_and_soar.md) rigenera il modello e rivalida il chokepoint SERVICE; non modifica finding o risultati delle scansioni. Il gate Linux/SSH rimane aperto.

## UI e scope

COMPANY-UX-02 · 2026-10-17: filtro CVSS rimosso dalla pagina, richiesta graph senza soglia; parametro backend preservato. Selettore “Intera azienda / Reparto” usa valori target del reparto per grafo/KPI/chokepoint, senza fallback quando vuoto. Intera azienda permette ipotesi fra reparti; singolo reparto esclude catene esterne al reparto. Azienda obbligatoria tramite guard route. Guida aggiornata; layout effect allinea selezione senza sovrascrivere click successivi al caricamento.

Proiezione aziendale locale di nodi, archi, percorsi, KPI e chokepoint; nessun fallback globale su scope vuoto. Un percorso con qualsiasi nodo fuori scope viene escluso. Scope IP/FQDN/IPv6 letterali, CIDR IPv4 e range IPv4 (estesi o ultimo ottetto); niente risoluzione implicita DNS. Non è un confine autorizzativo aziendale: API autorizzata per tenant.

Vista 2D percorso: sequenza sinistra→destra, scorrimento orizzontale per conservare leggibilità. Mappa completa: BFS con vertici visitati una volta, terminante anche su cicli. Pan con Pointer Events, zoom, reset, selezione mouse/tastiera, legenda osservato/ipotesi e tooltip motivazione. Gli archi evidenziati sono quelli in `edge_ids`, non ogni arco fra due nodi selezionati.

Guida contestuale in `AttackPathGuide.tsx`, fuori da `docs/` che rimane documentazione agentica. Selezione non invia richieste; filtro/aggiornamento sì, con risposte obsolete scartate e rimozione dati precedenti su errore.

3D non aggiunto: decisione di leggibilità della sequenza, non risultato di un confronto utenti o benchmark visuale. Verifica browser e responsive reale ancora aperta per indisponibilità del runtime.



