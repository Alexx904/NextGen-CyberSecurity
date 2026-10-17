# Network Discovery: analisi e intervento DISCOVERY-01

Ultimo allineamento: 2026-10-17 · working tree.

Accettazione PO: 2026-10-17, “OK”, con limiti dichiarati preservati.
Nessun commit, push o deploy incluso nell'accettazione.

## Scope e contratto

Playbook feedback (modifica) e orchestration; revisione svolta dallo stesso agente.
Backend soltanto, API e database invariati. Nessuna dipendenza o scansione reale.
Obiettivo: ridurre false attribuzioni e rendere limitato il costo dei probe HTTP.

## Riscontri e correzioni

| Riscontro nel codice | Intervento |
|---|---|
| MAC non esadecimali, nulli e multicast accettati | Validazione degli ottetti e unicast prima dell'uso |
| Cache solo tenant/IP, senza scadenza | Riutilizzo con MAC corrente identico, TTL 15 minuti e reset su cambio identità |
| Vendor Nmap usato dopo euristiche | Preferenza OUI, quindi vendor Nmap, quindi euristica |
| PTR fallback escluso dagli indizi | Hostname risolto disponibile prima dell'inferenza |
| Windows, nginx, Linux e Samba etichettati come produttori | Ritorno Unknown senza indizio di marca |
| Canon/Epson/Brother accorpati come Printer Device | Marche distinte |
| HTTP segue redirect e somma timeout per porta | Redirect disabilitati, budget complessivo 2s/host, chiusura connessioni |
| URL IPv6 e titoli multilinea/entità | JoinHostPort, regex multilinea e decodifica HTML |
| /31 e /32 eliminati come rete/broadcast | Indirizzi mantenuti |

## Precisione residua e ottimizzazioni successive

- OUI embedded è un catalogo parziale a 24 bit. Nmap può aggiungere assegnatari
  mancanti; nessun aggiornamento IEEE scaricato o nuova copertura dichiarata.
- Il produttore della scheda di rete può differire dal dispositivo. MAC locali
  possono appartenere a virtualizzazione oltre che a Wi-Fi privato; l'etichetta
  storica Private MAC non dimostra che il dispositivo sia mobile.
- Le marche da hostname/banner restano inferenze. Titolo HTTP e nomi sintetici
  sono ancora combinati nel campo hostname storico: non equivalgono a un PTR.
  Un futuro contratto dovrebbe distinguere hostname, display name, vendor della
  NIC, vendor inferito, fonte e confidenza, con adeguata UI e importazione.
- Stage 2 usa -F senza -sV: non promette fingerprint delle versioni. Rimane
  opzionale e limitato a 64 host. Nessun ampliamento implicito della scansione.
- HTTP resta sequenziale tra host; il budget limita il costo individuale,
  senza benchmark o percentuali di accelerazione dichiarate. Un pool limitato
  è un successivo intervento da misurare, preservando SSE e cancellazione.
- ARP rappresenta la visibilità del segmento dello scanner/container; oltre
  router/VLAN non può fornire automaticamente il MAC dell'host remoto.
- Cache storica non usata senza MAC corrente verificabile: scelta conservativa
  per evitare associazioni obsolete dopo DHCP. Nessuna garanzia contro spoofing.

## Verifica

Test regressione MAC/software eseguiti prima del fix: FAIL atteso exit 1
con otto false attribuzioni riprodotte. Cache Go iniziale non scrivibile:
GOCACHE spostata in `.agent/artifacts/go-cache`, senza pulizie.
Gate finale mediante `scripts/verify.ps1 -Profile Final -Scope Backend`:
suite Go, vet, build, verificatore documentale e diff. Evidenze misurate nel
report JSON ignorato `.agent/artifacts/testing/`.
Test HTTP con server locali effimeri; nessuna rete cliente interrogata.
Frontend non modificato; PostgreSQL non coinvolto. Review del diff manuale
su permessi, cache, timeout, compatibilità e dati non attendibili.
