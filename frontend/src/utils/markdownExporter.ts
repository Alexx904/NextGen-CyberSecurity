/**
 * NextGen CyberSecurity - Diagnostics & AI Markdown Exporter
 * Genera report diagnostici strutturati in formato Markdown per analisi e fixing assistito via AI.
 * Include esportatore JSON e helper per clipboard/download.
 */

import { DebugEntry, DebugLoggerStats } from '../types/debug';

/**
 * Genera un report Markdown strutturato pronto per la chat con assistenti AI
 */
export function generateAiMarkdownReport(
  entries: DebugEntry[],
  stats: DebugLoggerStats,
  currentUrl?: string
): string {
  const activeRoute =
    currentUrl ||
    (typeof window !== 'undefined' && window.location
      ? window.location.pathname + window.location.search
      : 'N/A');

  const userAgent =
    typeof navigator !== 'undefined' && navigator.userAgent
      ? navigator.userAgent
      : 'Unknown User Agent / Node Environment';

  const nowIso = new Date().toISOString();
  const errorEntries = entries.filter((e) => e.level === 'error');

  let md = `# 🛡️ NextGen CyberSecurity - Diagnostic Session Report\n\n`;

  // 1. Riepilogo Sessione
  md += `## 1. Riepilogo Sessione\n`;
  md += `- **Data/Ora Report:** \`${nowIso}\`\n`;
  md += `- **Rotta Attiva:** \`${activeRoute}\`\n`;
  md += `- **Ambiente Client / Browser:** \`${userAgent}\`\n`;
  md += `- **Totale Eventi Bufferizzati:** ${stats.total}\n`;
  md += `- **Errori Rilevati:** **${stats.errors}** | **Avvisi (Warn):** ${stats.warnings} | **Info:** ${stats.info}\n\n`;

  // 2. Tabella Sinottica degli Errori
  md += `## 2. Tabella Sinottica degli Errori\n\n`;
  if (errorEntries.length === 0) {
    md += `> ✅ **Nessun errore critico rilevato nel buffer corrente.** Tutti i sistemi e le chiamate API nominali.\n\n`;
  } else {
    md += `| # | Orario | Sorgente | Titolo / Endpoint | Status | Latenza |\n`;
    md += `| :---: | :--- | :---: | :--- | :---: | :---: |\n`;

    errorEntries.forEach((err, idx) => {
      const timeStr = err.timestamp.split('T')[1]?.replace('Z', '') || err.timestamp;
      const statusStr = err.status !== undefined ? String(err.status) : '-';
      const durationStr = err.durationMs !== undefined ? `${err.durationMs}ms` : '-';
      const cleanTitle = err.title.replace(/\|/g, '\\|');
      md += `| ${idx + 1} | \`${timeStr}\` | \`${err.source.toUpperCase()}\` | ${cleanTitle} | \`${statusStr}\` | ${durationStr} |\n`;
    });
    md += `\n`;

    // 3. Schede Dettagliate per ogni Errore
    md += `## 3. Schede Dettagliate degli Errori (Deep-Dive)\n\n`;
    errorEntries.forEach((err, idx) => {
      md += `### Errore #${idx + 1}: ${err.title}\n`;
      md += `- **ID Evento:** \`${err.id}\`\n`;
      md += `- **Timestamp:** \`${err.timestamp}\`\n`;
      md += `- **Sorgente:** \`${err.source}\`\n`;
      if (err.url) md += `- **URL / Endpoint:** \`${err.url}\`\n`;
      if (err.method) md += `- **Metodo HTTP:** \`${err.method}\`\n`;
      if (err.status) md += `- **Status Code:** \`${err.status}\`\n`;
      if (err.durationMs !== undefined) md += `- **Durata Round-Trip:** \`${err.durationMs}ms\`\n`;
      if (err.error) md += `- **Messaggio di Errore:** \`${err.error}\`\n`;

      if (err.requestData !== undefined) {
        md += `\n**Corpo Richiesta (Sanificato):**\n\`\`\`json\n${JSON.stringify(
          err.requestData,
          null,
          2
        )}\n\`\`\`\n`;
      }

      if (err.responseData !== undefined) {
        md += `\n**Risposta Server (Sanificata):**\n\`\`\`json\n${JSON.stringify(
          err.responseData,
          null,
          2
        )}\n\`\`\`\n`;
      }

      if (err.stack) {
        md += `\n**Stack Trace:**\n\`\`\`text\n${err.stack.trim()}\n\`\`\`\n`;
      }

      if (err.metadata && Object.keys(err.metadata).length > 0) {
        md += `\n**Metadati Aggiuntivi:**\n\`\`\`json\n${JSON.stringify(
          err.metadata,
          null,
          2
        )}\n\`\`\`\n`;
      }

      md += `\n---\n\n`;
    });
  }

  // Sezione Recenti Eventi di Contesto (ultimi 5 log indipendentemente dal livello)
  const recentEntries = entries.slice(-5);
  if (recentEntries.length > 0) {
    md += `## 4. Cronologia Ultimi Eventi Recenti (Contesto)\n\n`;
    recentEntries.forEach((entry, idx) => {
      const timeStr = entry.timestamp.split('T')[1]?.replace('Z', '') || entry.timestamp;
      md += `${idx + 1}. \`[${timeStr}]\` **[${entry.level.toUpperCase()}]** [${entry.source}] ${entry.title}\n`;
    });
    md += `\n`;
  }

  return md;
}

/**
 * Esporta l'intero buffer di log in formato JSON formattato
 */
export function exportLogsToJson(entries: DebugEntry[]): string {
  return JSON.stringify(entries, null, 2);
}

/**
 * Helper per copiare una stringa negli appunti di sistema con fallback sicuro
 */
export async function copyTextToClipboard(text: string): Promise<boolean> {
  if (typeof navigator !== 'undefined' && navigator.clipboard && navigator.clipboard.writeText) {
    try {
      await navigator.clipboard.writeText(text);
      return true;
    } catch {
      // Fallback
    }
  }

  // Fallback con elemento textarea invisibile per browser con permessi restrittivi
  if (typeof document !== 'undefined') {
    try {
      const textArea = document.createElement('textarea');
      textArea.value = text;
      textArea.style.position = 'fixed';
      textArea.style.left = '-9999px';
      textArea.style.top = '0';
      document.body.appendChild(textArea);
      textArea.focus();
      textArea.select();
      const success = document.execCommand('copy');
      document.body.removeChild(textArea);
      return success;
    } catch {
      return false;
    }
  }

  return false;
}

/**
 * Helper per avviare il download di un file nel browser
 */
export function triggerFileDownload(content: string, filename: string, mimeType = 'text/plain'): void {
  if (typeof document === 'undefined') return;

  const blob = new Blob([content], { type: mimeType });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}
