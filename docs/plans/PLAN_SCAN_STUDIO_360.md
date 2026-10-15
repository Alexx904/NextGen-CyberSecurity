# Implementation Plan · Epic 27.1: Semplificazione Centro Scansioni & Scanning Studio 360° (FEAT-SCAN-STUDIO)

> **Stato:** APPROVATO DAL PRODUCT OWNER (In attesa di esecuzione in sessione successiva)  
> **Data Redazione:** 2026-10-15  
> **Autore:** TechLead Engineer Agent  
> **Documenti Correlati:** [docs/3_TASK_LIST.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/3_TASK_LIST.md) | [docs/GUIDA_FUNZIONALE_PIATTAFORMA.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/GUIDA_FUNZIONALE_PIATTAFORMA.md) | [docs/domains/scanning_and_discovery.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/domains/scanning_and_discovery.md)

---

## 1. Obiettivo e Motivazione Architetturale

L'attuale modale di avvio scansione (`NewScanModal.tsx`) presenta un menu a tendina con **14 profili preconfigurati** suddivisi in 6 categorie.  
Come evidenziato dall'analisi di sicurezza e dal Product Owner, questo approccio genera **sovraccarico cognitivo (*paradox of choice*)**, ridondanze operative e rigidità nell'uso quotidiano.

### Direttiva del Product Owner
1. **Semplificazione dell'interfaccia:** eliminare l'elenco statico di 14 profili dispersivi.
2. **Master Presets Intelligenti (1 Clic):** 3-4 scenari operativi principali chiari e standardizzati che coprono il 95% delle necessità operative.
3. **Pannello "Fine-Tuning a 360°" (Libertà Completa):** possibilità di personalizzare scope porte, sonde attive, velocità e digitare comandi avanzati per power user nello stesso schermo.
4. **Calcolo Dinamico dell'Invasività:** il sistema rileva in automatico se i parametri scelti rendono la scansione `INTRUSIVE` e attiva il Gate Obbligatorio di Consenso solo quando serve.

---

## 2. Nuova Architettura UX/UI: "Smart Scanning Studio"

La modale `NewScanModal.tsx` viene unificata in un'unica schermata divisa in tre aree logiche:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ 🎯 NUOVA SCANSIONE DI SICUREZZA                                      [ X ]  │
├─────────────────────────────────────────────────────────────────────────────┤
│ Bersaglio: [ Seleziona Reparto o Singolo IP dell'Azienda Attiva ▼ ]         │
│                                                                             │
│ [ Scheda: Scansione Immediata ]     [ Scheda: Schedulazione Ricorrente ]    │
│                                                                             │
│ ── 1. SCENARI CONSIGLIATI (1 CLIC) ──────────────────────────────────────── │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐           │
│  │ 🟢 DISCOVERY     │  │ 🔵 VULN AUDIT    │  │ 🔴 FULL PENTEST  │           │
│  │    RAPIDA        │  │    (CONSIGLIATO) │  │    DEEP 360°     │           │
│  │ Host vivi + Top  │  │ Top 1000 porte + │  │ Tutte 65.535     │           │
│  │ 100 porte        │  │ versioni demoni  │  │ porte + CVE OS   │           │
│  │ ~1 min · SAFE    │  │ ~5 min · SAFE    │  │ ~45 min·INTRUSIVE│           │
│  └──────────────────┘  └──────────────────┘  └──────────────────┘           │
│  ┌──────────────────────────────────────────────────────────────┐           │
│  │ 🌐 WEB & API SURFACE FOCUS (Porte 80, 443, 8080, 8443... SAFE)│           │
│  └──────────────────────────────────────────────────────────────┘           │
│                                                                             │
│ ── 2. PERSONALIZZAZIONE A 360° (FINE-TUNING GRANULARE) ──────────────────── │
│  Scope Porte:                                                               │
│    (•) Top 100 (-F)   ( ) Top 1000   ( ) Tutte 65k (-p-)   ( ) Custom: [   ]│
│                                                                             │
│  Sonde e Rilevamento Attivo:                                                │
│    [x] Rilevamento Versioni Servizi (-sV)                                   │
│    [ ] Rilevamento Sistema Operativo (-O)                                   │
│    [ ] Script Vulnerabilità Note (--script=vuln)                            │
│    [ ] Includi Servizi Critici UDP (-sU)                                    │
│                                                                             │
│  Velocità ed Evasione Firewall:                                             │
│    [ Stealth / Lento (-T2) ]   [ Bilanciato (-T3) ]   [ Veloce (-T4) ]      │
│                                                                             │
│  Argomenti Avanzati Power User (Opzionale):                                 │
│    [ Inserisci flag Nmap custom con validazione sicura...                 ] │
│                                                                             │
│ ── 3. ANTEPRIMA & SICUREZZA ─────────────────────────────────────────────── │
│  Live Command Preview:                                                      │
│  $ nmap -sS -sV -T4 --open 192.168.1.50                                     │
│                                                                             │
│  ⚠️ [x] Dichiaro autorizzazione formale scritta (compare solo se INTRUSIVE) │
│                                                                             │
│                                           [ Annulla ]  [ 🚀 Lancia Scansione ]│
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Specifiche dei 4 Master Presets Standard

| Preset | Nome Interno / ID | Argomenti Nmap Generati | Livello Invasività | Descrizione Operativa |
| :--- | :--- | :--- | :--- | :--- |
| **Discovery Rapida** | `PRESET_DISCOVERY` | `-T4 -F` | `SAFE` | Mappatura host vivi e 100 porte più comuni. Durata: <60s. Impatto zero. |
| **Vulnerability Audit** | `PRESET_VULN_AUDIT` | `-sV -T4 --open` | `SAFE` | **Preset predefinito raccomandato.** Top 1000 porte standard + estrazione versioni servizi per correlazione automatica CVE e calcolo Posture Score. |
| **Full Pentest Deep 360°** | `PRESET_FULL_DEEP` | `-sS -sV -O -p- --script=vuln -T4 --open` | `INTRUSIVE` | Audit esaustivo su 65.535 porte TCP, OS detection e script di exploit. Attiva il Gate Obbligatorio di Consenso. |
| **Web & API Focus** | `PRESET_WEB_API` | `-sV -p 80,443,8000,8080,8443,8888,3000,5000,9000 -T4 --open` | `SAFE` | Ispezione perimetrale di server web, API REST, reverse proxy e container. |

---

## 4. Dettaglio Controlli del Pannello "Fine-Tuning 360°"

1. **Scope Porte:**
   - Bottoni di scelta rapida:
     - `Top 100`: inserisce `-F`.
     - `Top 1000`: default Nmap standard (nessun flag porta esplicito).
     - `Tutte le 65.535`: inserisce `-p-` (fa scattare `INTRUSIVE`).
     - `Custom`: apre input text validato (es. `80,443,8080` oppure `1-1024`).
2. **Sonde e Profondità di Ispezione:**
   - Checkbox *Rilevamento Versioni Servizi*: aggiunge `-sV`.
   - Checkbox *Riconoscimento Sistema Operativo*: aggiunge `-O`.
   - Checkbox *Script Vulnerabilità Note*: aggiunge `--script=vuln` (fa scattare `INTRUSIVE`).
   - Checkbox *Servizi Critici UDP*: aggiunge `-sU -p U:53,69,123,161,514,623,5353`.
3. **Velocità & Discrezione:**
   - Bottoni radio o toggle a 3 stati:
     - `Stealth (-T2)`: per eludere firewall e non allarmare IDS.
     - `Bilanciato (-T3)`: velocità standard Nmap.
     - `Veloce (-T4)`: raccomandato per reti moderne e test rapidi.
4. **Campo Libero per Power User (Argomenti Custom):**
   - Campo input testo libero per aggiungere flag specifici.
   - Validazione preventiva sul client tramite la funzione di sicurezza `IsSafeArgument` (nessun metacarattere shell, nessun flag vietato come `-oN`, `-iL`, `--interactive`).
5. **Calcolo Dinamico dell'Invasività:**
   - Funzione helper `computeInvasiveness(args: string[])`:
     - Se `args` contiene `-p-`, `--script=vuln`, `-sA`, o timeout > 60m ➔ `INTRUSIVE`.
     - Altrimenti ➔ `SAFE`.
   - Quando `INTRUSIVE`, il pulsante *"Lancia Scansione"* si disabilita e compare la checkbox obbligatoria di autorizzazione formale.

---

## 5. File da Modificare nella Sessione di Implementazione

1. **`frontend/src/components/NewScanModal.tsx`**:
   - Refactor del componente: rimozione del selettore a tendina a 14 opzioni disperse.
   - Integrazione card di selezione per i 4 Master Presets.
   - Integrazione controlli granulari Fine-Tuning 360° (Port Scope, Sonde, Timing, Custom Arg).
   - Calcolo reattivo di `currentNmapArgs` e `currentInvasiveness`.
   - Live Command Preview sempre visibile e aggiornato.
2. **`frontend/src/components/NewScanModal.test.tsx`**:
   - Aggiornamento / estensione dei test Vitest:
     - Render iniziale con preset "Vulnerability Audit" selezionato di default.
     - Switch a "Discovery Rapida" e "Full Pentest Deep 360°".
     - Verifica che il Full Pentest mostri il Gate di Consenso obbligatorio.
     - Test dei controlli di Fine-Tuning: selezione porte custom, toggle sonde, slider timing.
     - Invio della scansione con i payload corretti all'API backend.
3. **`docs/GUIDA_FUNZIONALE_PIATTAFORMA.md` & `docs/domains/scanning_and_discovery.md`**:
   - Aggiornamento della sezione Feature 6 con spiegazione della nuova interfaccia a 2 livelli ("Scanning Studio 360°").
   - Nuovi screenshot della modale aggiornata.

---

## 6. Piano di Esecuzione Step-by-Step (Per la prossima chat)

1. **Step 1:** Riscrivere `NewScanModal.tsx` implementando i 4 Master Presets + selettori granulari 360° + Live Preview + calcolo dinamico invasività.
2. **Step 2:** Aggiornare e lanciare i test mirati: `npx vitest run src/components/NewScanModal.test.tsx`.
3. **Step 3:** Eseguire la suite di verifica zero-assumption completa: `npm test` e `npm run build` (0 errori).
4. **Step 4:** Catturare il nuovo screenshot con Chrome headless e aggiornare `docs/screenshots/09_new_scan_modal.png`.
5. **Step 5:** Aggiornare [docs/GUIDA_FUNZIONALE_PIATTAFORMA.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/GUIDA_FUNZIONALE_PIATTAFORMA.md) e [docs/domains/scanning_and_discovery.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/domains/scanning_and_discovery.md).
6. **Step 6:** Aggiornare [docs/3_TASK_LIST.md](file:///c:/Users/aless/Desktop/Sviluppo/GitHub/NextGen-CyberSecurity/docs/3_TASK_LIST.md) contrassegnando `[~]` per revisione del PO e aggiornare `docs/PROJECT_STATE.md`.

