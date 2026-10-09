# Dominio: Attack Path Mapping & Threat Chaining (Milestone 2)

> **Dominio:** Analisi Topologica delle Minacce, Correlazione Lateral Movement & Ottimizzazione Remediation  
> **File di Riferimento:** `backend/internal/domain/attack_path.go`, `backend/internal/service/attack_path_service.go`, `backend/internal/delivery/http/handler/attack_path_handler.go`, `frontend/src/pages/AttackPath.tsx`, `frontend/src/components/graph/AttackPathGraph.tsx`

---

## 1. Visione e Modello Concettuale

L'**Attack Path Mapping** supera la visualizzazione frammentata della singola vulnerabilità isolata trasformando i dati di scansione in un modello predittivo e orientato al business:
- **Dal singolo finding alla catena di compromissione:** Un database interno isolato potrebbe non essere vulnerabile direttamente dall'esterno, ma diventa compromettibile se un server web frontale in DMZ presenta un'esecuzione remota di codice (RCE) e ha visibilità di rete o credenziali di accesso al database.
- **Identificazione dei Chokepoint:** Invece di costringere il team di sicurezza a patchare centinaia di anomalie a basso impatto, l'algoritmo individua il "punto di strozzatura" (chokepoint) la cui rimozione neutralizza all'istante l'intero albero di attacco (massimo Return On Investment di sicurezza).

---

## 2. Modello Dati del Grafo (Graph Topology)

### 2.1 Tipologia Nodi (`AttackNodeType`)
- `ENTRYPOINT`: Il punto di partenza dell'attaccante (es. Internet pubblico o subnet untrusted).
- `HOST`: Macchina fisica o virtuale (IP, hostname, sistema operativo, ruolo).
- `SERVICE`: Porta e protocollo in ascolto (es. HTTP 80, SSH 22, Postgres 5432).
- `VULNERABILITY`: Finding o CVE correlata con CVSS v3.1 e vettore di attacco.

### 2.2 Tipologia Archi (`AttackEdgeType`)
- `NETWORK_ACCESS`: Raggiungibilità di rete tra l'esterno e un host/servizio o tra subnet diverse.
- `EXPLOITS`: Relazione tra un attaccante/servizio e una specifica vulnerabilità sfruttabile.
- `PIVOTS_TO`: Capacità di eseguire lateral movement da un nodo compromesso verso la rete interna.
- `TRUSTED_RELATION`: Relazione fiduciaria o credenziali condivise (es. chiave SSH o accesso DB).

### 2.3 Formula di Cumulative Path Risk Score
Ogni percorso $P = (N_0, N_1, \dots, N_k)$ riceve un punteggio di rischio aggregato normalizzato su una scala da 0 a 100:

$$\text{RiskScore}(P) = \min\left(100,\, \sum_{i \in \text{Vulns}} \text{CVSS}_i \times \text{Weight}_{\text{Exploit}} \times \left(1 + \frac{1}{\sqrt{\text{HopCount}}}\right)\right)$$

- Punteggio maggiore per percorsi corti e diretti (minor complessità per l'attaccante).
- Moltiplicatore per exploit con vettore di rete (`AV:N`) e privilegi bassi o nulli (`PR:N`).

---

## 3. Algoritmo di Individuazione dei Chokepoint

Un **Chokepoint** è un nodo $N$ che compare nel massimo numero di cammini critici da un entrypoint ai nodi target critici:
1. Enumerazione di tutti i cammini aciclici $P_1, \dots, P_m$ con `RiskScore` $\ge 50$.
2. Conteggio delle occorrenze per ciascun nodo intermedio: $\text{Frequency}(N) = \sum_{j=1}^m [N \in P_j]$.
3. Calcolo del **Mitigated Risk Impact**: somma dei punteggi di rischio dei percorsi che verrebbero interrotti rimuovendo o sanando il nodo $N$.
4. Ordinamento decrescente: i primi 3 nodi costituiscono le raccomandazioni primarie di remediation.

---

## 4. Specifiche Contratti API REST (`/api/v1/attack-paths`)

### 4.1 Generazione Grafo Topologico
- `GET /api/v1/attack-paths/graph?min_cvss=5.0`
- **Risposta (JSON):**
```json
{
  "success": true,
  "data": {
    "tenant_id": "00000000-0000-0000-0000-000000000001",
    "generated_at": "2026-10-09T22:00:00Z",
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
        "label": "Public HTTP 80"
      },
      {
        "id": "e2",
        "source": "host-192-168-1-10",
        "target": "vuln-cve-2021-41773",
        "type": "EXPLOITS",
        "label": "RCE Exploit"
      },
      {
        "id": "e3",
        "source": "host-192-168-1-10",
        "target": "host-10-0-0-50",
        "type": "PIVOTS_TO",
        "label": "Internal DB Access"
      }
    ],
    "paths": [
      {
        "id": "path-1",
        "name": "External RCE to Internal Database Pivot",
        "node_ids": ["entry-internet", "host-192-168-1-10", "vuln-cve-2021-41773", "host-10-0-0-50"],
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
        "remediation": "Aggiornare Apache HTTP Server a versione >= 2.4.51 per eliminare l'RCE e interrompere la catena di penetrazione."
      }
    ],
    "summary": {
      "total_nodes": 4,
      "total_edges": 3,
      "critical_paths": 1,
      "highest_risk_score": 86.4
    }
  }
}
```

### 4.2 Elenco Chokepoint Prioritari
- `GET /api/v1/attack-paths/chokepoints`
- Restituisce la lista ordinata dei nodi la cui bonifica azzera il massimo volume di rischio per il tenant.

