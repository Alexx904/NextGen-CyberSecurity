import React from 'react';

export const AttackPathGuide: React.FC = () => (
  <details className="rounded-xl border border-cyan-500/25 bg-slate-900 p-4 text-sm text-slate-300">
    <summary className="cursor-pointer font-semibold text-cyan-200">Guida all’uso di Attack Path</summary>
    <div className="grid gap-6 mt-5 lg:grid-cols-2 leading-relaxed">
      <div>
        <h2 className="font-semibold text-white">Dalla scansione alla priorità</h2>
        <ol className="list-decimal pl-5 mt-3 space-y-3">
          <li><strong>Prepara i dati.</strong> Censisci IP o reti nei reparti dell’azienda. Esegui una scansione autorizzata con rilevamento servizi e vulnerabilità. Più recenti sono i dati, più utile è l’analisi.</li>
          <li><strong>Scegli azienda e reparto.</strong> Il selettore globale limita l’azienda; “Intera azienda” include tutti i reparti, mentre il selettore locale limita il grafo a un reparto. Nodi, percorsi, KPI e chokepoint seguono lo scope. Un reparto senza target resta vuoto. Tutte le CVE sono incluse: non c’è filtro CVSS.</li>
          <li><strong>Parti da un percorso.</strong> Selezionalo nella scheda Percorsi e leggi la catena da sinistra a destra. Ogni freccia corrisponde a un passaggio del modello. “Mappa completa” mostra il contesto; “Adatta vista” ripristina l’inquadratura.</li>
          <li><strong>Ispeziona le prove.</strong> Clicca un nodo, oppure selezionalo con Tab e Invio. Controlla IP, protocollo, porta, CVSS e metadati del finding, inclusi metodo, confidenza e data quando disponibili. Passa sulle frecce per leggere il motivo dell’inferenza.</li>
          <li><strong>Verifica e intervieni.</strong> Consulta Chokepoints, confronta impatto operativo e percorsi interessati, copia la raccomandazione e concorda il cambiamento. Dopo la correzione ripeti la scansione e confronta i finding. Questa schermata non applica modifiche remote.</li>
        </ol>
      </div>
      <div className="space-y-4">
        <div><h2 className="font-semibold text-white">Come interpretare le frecce</h2>
          <p className="mt-2"><strong className="text-cyan-300">Continua:</strong> associazione host/servizio presente nei dati. Non garantisce accessibilità da Internet.</p>
          <p className="mt-2"><strong className="text-violet-300">Tratteggiata:</strong> accesso iniziale ipotetico o CVE correlata, con sfruttabilità da verificare.</p>
          <p className="mt-2"><strong className="text-orange-300">Laterale /24:</strong> vicinanza degli indirizzi IPv4, senza prova della maschera reale. Non dimostra routing, controllo dell’host, credenziali valide o superamento di ACL/firewall.</p>
        </div>
        <div><h2 className="font-semibold text-white">Esempio di lettura</h2>
          <p className="mt-2">Origine ipotetica → 10.0.0.10 → 443/tcp → CVE → 10.0.0.20 significa: verificare prima accesso e vulnerabilità del servizio, poi se un eventuale controllo della sorgente permetterebbe accesso al secondo host. Non significa che i due host siano già compromessi.</p>
        </div>
        <div><h2 className="font-semibold text-white">Score, chokepoint e assenza di percorsi</h2>
          <p className="mt-2">L’indice euristico non misura la probabilità di successo. I punti aggregati dei chokepoint possono superare 100 e non vanno sommati fra nodi, perché gli stessi percorsi possono sovrapporsi. Una raccomandazione non garantisce l’eliminazione di tutti gli attacchi.</p>
          <p className="mt-2">Nessun percorso può significare dati mancanti, filtro CVSS, scope vuoto o copertura parziale. L’inventario SSH non prova fiducia fra host. Le osservazioni storiche non certificano lo stato attuale.</p>
        </div>
        <p className="text-xs text-slate-400">Riferimenti: <a className="text-cyan-300 underline" href="https://www.first.org/cvss/user-guide.html" target="_blank" rel="noreferrer">FIRST CVSS</a> · <a className="text-cyan-300 underline" href="https://attack.mitre.org/tactics/TA0008/" target="_blank" rel="noreferrer">MITRE ATT&amp;CK: Lateral Movement</a>.</p>
      </div>
    </div>
  </details>
);
