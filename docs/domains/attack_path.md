# Dominio: Attack Path Mapping & Threat Chaining (Milestone 2)

> **Dominio:** Analisi Topologica delle Minacce, Correlazione Lateral Movement, Visual Graph Explorer & Ottimizzazione Remediation  
> **File di Riferimento:** `backend/internal/domain/attack_path.go`, `backend/internal/service/attack_path_service.go`, `backend/internal/delivery/http/handler/attack_path_handler.go`, `frontend/src/types/attackPath.ts`, `frontend/src/api/attackPath.ts`, `frontend/src/pages/AttackPath.tsx`, `frontend/src/components/graph/AttackPathGraph.tsx`

---

## 1. Visione e Modello Concettuale

L'**Attack Path Mapping** (`FR-14`, `UC-13`) supera la visualizzazione frammentata della singola vulnerabilità isolata trasformando i dati di scansione in un modello predittivo e orientato al business:
- **Dal singolo finding alla catena di compromissione:** Un database interno isolato potrebbe non essere vulnerabile direttamente dall'esterno, ma diventa compromettibile se un server web frontale in DMZ presenta un'esecuzione remota di codice (RCE) e ha visibilità di rete o credenziali di accesso al database.
- **Identificazione dei Chokepoint:** Invece di costringere il team di sicurezza a patchare centinaia di anomalie a basso impatto, l'algoritmo individua il "punto di strozzatura" (chokepoint) la cui rimozione neutralizza all'istante l'intero albero di attacco (massimo Return On Investment di sicurezza).

---

## 2. Modello Dati del Grafo (Graph Topology)

### 2.1 Tipologia Nodi (`AttackNodeType`)
- `ENTRYPOINT`: Il punto di partenza dell'attaccante (es. Internet pubblico o perimetro untrusted).
- `SUBNET`: Segmento o rete locale di transito (es. `192.168.1.0/24`).
- `HOST`: Macchina fisica o virtuale (IP, hostname, sistema operativo, ruolo).
- `SERVICE`: Porta e protocollo in ascolto (es. HTTP 80, SSH 22, Postgres 5432).
- `VULNERABILITY`: Finding o CVE correlata con CVSS v3.1 e vettore di attacco.

### 2.2 Tipologia Archi (`AttackEdgeType`)
- `NETWORK_ACCESS`: Raggiungibilità di rete tra l'esterno e un host/servizio o tra subnet diverse.
- `EXPLOITS`: Relazione tra un attaccante/servizio e una specifica vulnerabilità sfruttabile.
- `PIVOTS_TO`: Capacità di eseguire lateral movement da un nodo compromesso verso la rete interna.
- `TRUSTED_RELATION`: Relazione fiduciaria o credenziali condivise (es. chiave SSH o accesso DB).

### 2.3 Formula di Cumulative Path Risk Score (`domain.CalculateCumulativeRisk`)
Ogni percorso $P = (N_0, N_1, \dots, N_k)$ riceve un punteggio di rischio aggregato (0.0 - 100.0):

$$\text{EffectiveHops} = \max(1.0, \, \text{HopCount})$$

$$\text{HopFactor} = 1.0 + \frac{1.0}{\sqrt{\text{EffectiveHops}}}$$

$$\text{RawRisk} = \left(\sum_{i \in \text{Vulns}} \text{CVSS}_i\right) \times \text{HopFactor} \times 5.0$$

- Se il percorso non contiene vulnerabilità esplicite (cammino puramente topologico), viene assunto $\sum \text{CVSS} = 2.0$.
- **Critical Vulnerability Safeguard:** Se il cammino include almeno una vulnerabilità critica ($\text{CVSS} \ge 9.0$), il rischio ha un pavimento minimo garantito di **80.0**.
- **Normalizzazione:** $\text{RiskScore}(P) = \min(100.0, \, \text{RawRisk})$.
- **Classificazione Severità Cammino:**
  - `CRITICAL`: $\text{RiskScore} \ge 75.0$
  - `HIGH`: $55.0 \le \text{RiskScore} < 75.0$
  - `MEDIUM`: $35.0 \le \text{RiskScore} < 55.0$
  - `LOW`: $\text{RiskScore} < 35.0$

---

## 3. Algoritmo di Individuazione dei Chokepoint (`domain.CalculateChokepoints`)

Un **Chokepoint** è un nodo intermedio $N$ la cui eliminazione interrompe il massimo numero di percorsi di attacco:
1. Analisi di tutti i cammini aciclici validi con almeno 3 nodi (escludendo il primo entrypoint e l'asset target finale).
2. Per ciascun vertice intermedio (vulnerabilità, host ponte o porta esposta), calcolo di:
   - $\text{MitigatedPaths}(N)$: conteggio dei percorsi unici interrotti.
   - $\text{MitigatedRisk}(N)$: somma cumulativa dei punteggi di rischio dei cammini neutralizzati.
3. Assegnazione di una prescrizione di remediation automatica contestualizzata per tipo di nodo (patch per CVE, isolamento/regole firewall per host, chiusura porta o whitelist per servizi).
4. Ordinamento decrescente per `MitigatedRisk` e prioritizzazione per le prime raccomandazioni.

---

## 4. Frontend Interactive Visual Graph Explorer (`AttackPathGraph.tsx`, `AttackPath.tsx`)

### 4.1 Layout Topologico a Livelli (SVG Canvas)
- **Motore di Distribuzione Spaziale:**
  - Colonna 0 ($X=0$): Nodi `ENTRYPOINT` (Internet).
  - Colonna 1 ($X=270$): Nodi `HOST` edge perimetrali e nodi `SUBNET`.
  - Colonna 2 ($X=540$): Nodi `SERVICE` e porte aperte.
  - Colonna 3 ($X=810$): Nodi `VULNERABILITY` (CVE critiche/alte).
  - Colonna 4 ($X=1080$): Target interni ad alto impatto (DB, Host interni pivotati).
- **Controlli Canvas:** Zoom In, Zoom Out, Reset (scala 1:1), e Pan bidirezionale con drag-and-drop.

### 4.2 Interattività e Highlighting Dinamico
- **Selezione Cammino:** Cliccando su un percorso nel pannello laterale, la catena di nodi e archi corrispondente viene illuminata con bordo colorato ed effetto pulsante (`animate-pulse`), mentre i nodi estranei vengono attenuati (opacità ridotta a 0.25).
- **Visualizzazione Chokepoint (Bullseye Animation):** I nodi identificati come chokepoint mostrano una doppia aura concentrica e l'icona fuoco/radio animata per attrarre l'attenzione visiva del SecOps.
- **Pannello a Schede Laterale:**
  - **Scheda "Percorsi":** Elenco dei percorsi con severity badge, hop count, cumulative risk score e target finale.
  - **Scheda "Chokepoints":** Classifica dei colli di bottiglia con contatore percorsi interrotti, rischio mitigato e prescrizione di remediation copiabile con 1-click.
  - **Scheda "Dettaglio Nodo":** Ispezione tecnica puntuale del vertice cliccato (IP, porta, protocollo, severità, CVSS, CVE ID e metadati).
- **Filtro CVSS Reattivo:** Dropdown per visualizzare nodi con soglia minima (Tutte, $\ge 4.0$, $\ge 7.0$, $\ge 9.0$).
- **Performance & Code Splitting:** Caricamento con `React.lazy` e chunk Vite isolato da soli 28 kB.

---

## 5. Specifiche Contratti API REST (`/api/v1/attack-paths`)

### 5.1 Generazione Grafo Topologico
- `GET /api/v1/attack-paths/graph?min_cvss=5.0`
- **Risposta (JSON):**
```json
{
  "success": true,
  "data": {
    "tenant_id": "00000000-0000-0000-0000-000000000001",
    "generated_at": "2026-10-10T15:00:00Z",
    "nodes": [
      {
        "id": "entry-internet",
        "type": "ENTRYPOINT",
        "label": "Internet Perimeter",
        "severity": "CRITICAL"
      },
      {
        "id": "host-192-168-1-10",
        "type": "HOST",
        "label": "srv-web-dmz (192.168.1.10)",
        "ip": "192.168.1.10",
        "severity": "CRITICAL"
      },
      {
        "id": "vuln-cve-2021-41773",
        "type": "VULNERABILITY",
        "label": "Apache RCE (CVE-2021-41773)",
        "cvss": 7.5,
        "severity": "CRITICAL"
      },
      {
        "id": "host-10-0-0-50",
        "type": "HOST",
        "label": "srv-db-internal (10.0.0.50)",
        "ip": "10.0.0.50",
        "severity": "HIGH"
      }
    ],
    "edges": [
      {
        "id": "e1",
        "source": "entry-internet",
        "target": "host-192-168-1-10",
        "type": "NETWORK_ACCESS",
        "label": "Public HTTP 80",
        "weight": 1.0
      },
      {
        "id": "e2",
        "source": "host-192-168-1-10",
        "target": "vuln-cve-2021-41773",
        "type": "EXPLOITS",
        "label": "RCE Exploit",
        "weight": 2.0
      },
      {
        "id": "e3",
        "source": "host-192-168-1-10",
        "target": "host-10-0-0-50",
        "type": "PIVOTS_TO",
        "label": "Internal DB Access",
        "weight": 1.5
      }
    ],
    "paths": [
      {
        "id": "path-1",
        "name": "External RCE to Internal Database Pivot",
        "node_ids": ["entry-internet", "host-192-168-1-10", "vuln-cve-2021-41773", "host-10-0-0-50"],
        "edge_ids": ["e1", "e2", "e3"],
        "cumulative_risk": 86.4,
        "severity": "CRITICAL",
        "hop_count": 3,
        "target_asset": "srv-db-internal (10.0.0.50)"
      }
    ],
    "chokepoints": [
      {
        "node_id": "host-192-168-1-10",
        "label": "srv-web-dmz (192.168.1.10)",
        "mitigated_paths": 1,
        "mitigated_risk": 86.4,
        "remediation": "Isolare l'host 'srv-web-dmz (192.168.1.10)' o restringere le regole firewall per bloccare il lateral movement verso la rete interna."
      }
    ],
    "summary": {
      "total_nodes": 4,
      "total_edges": 3,
      "total_paths": 1,
      "critical_paths": 1,
      "highest_risk_score": 86.4,
      "top_chokepoint": "srv-web-dmz (192.168.1.10)"
    }
  }
}
```

### 5.2 Elenco Chokepoint Prioritari
- `GET /api/v1/attack-paths/chokepoints`
- Restituisce la slice ordinata dei nodi la cui remediation azzera il massimo volume di rischio mitigato per il tenant.


