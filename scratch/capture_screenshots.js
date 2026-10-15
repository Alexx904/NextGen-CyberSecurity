import { spawn } from 'node:child_process';
import { writeFileSync, mkdirSync, copyFileSync } from 'node:fs';
import { join } from 'node:path';

const DOCS_SCREENSHOTS_DIR = 'c:\\Users\\aless\\Desktop\\Sviluppo\\GitHub\\NextGen-CyberSecurity\\docs\\screenshots';
const ARTIFACT_SCREENSHOTS_DIR = 'C:\\Users\\aless\\.gemini\\antigravity\\brain\\a137650a-1c83-4904-9546-40df65208649\\screenshots';

mkdirSync(DOCS_SCREENSHOTS_DIR, { recursive: true });
mkdirSync(ARTIFACT_SCREENSHOTS_DIR, { recursive: true });

async function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

async function main() {
  console.log('Avvio Google Chrome headless...');
  const chromePath = 'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe';
  const userProfileDir = 'C:\\Users\\aless\\.gemini\\antigravity\\brain\\a137650a-1c83-4904-9546-40df65208649\\scratch\\chrome-profile';
  mkdirSync(userProfileDir, { recursive: true });

  const chrome = spawn(chromePath, [
    '--headless=new',
    '--remote-debugging-port=9222',
    '--disable-gpu',
    '--no-sandbox',
    '--window-size=1600,1000',
    '--hide-scrollbars',
    `--user-data-dir=${userProfileDir}`
  ], { stdio: 'ignore' });

  let version = null;
  for (let i = 0; i < 30; i++) {
    try {
      const res = await fetch('http://127.0.0.1:9222/json/version');
      version = await res.json();
      break;
    } catch {
      await sleep(200);
    }
  }

  if (!version) {
    console.error('Chrome debugging port 9222 non disponibile');
    chrome.kill();
    process.exit(1);
  }

  const targetsRes = await fetch('http://127.0.0.1:9222/json/list');
  const targets = await targetsRes.json();
  const pageTarget = targets.find(t => t.type === 'page') || targets[0];

  const ws = new WebSocket(pageTarget.webSocketDebuggerUrl);

  let idCounter = 1;
  const callbacks = new Map();

  ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    if (msg.id && callbacks.has(msg.id)) {
      const cb = callbacks.get(msg.id);
      callbacks.delete(msg.id);
      if (msg.error) cb.reject(new Error(msg.error.message));
      else cb.resolve(msg.result);
    }
  };

  await new Promise((resolve) => ws.onopen = resolve);

  function send(method, params = {}) {
    return new Promise((resolve, reject) => {
      const id = idCounter++;
      callbacks.set(id, { resolve, reject });
      ws.send(JSON.stringify({ id, method, params }));
    });
  }

  async function evalJs(expr) {
    const res = await send('Runtime.evaluate', {
      expression: expr,
      returnByValue: true,
      awaitPromise: true
    });
    return res.result ? res.result.value : null;
  }

  async function takeScreenshot(filename) {
    const res = await send('Page.captureScreenshot', { format: 'png' });
    const buffer = Buffer.from(res.data, 'base64');
    const pathDocs = join(DOCS_SCREENSHOTS_DIR, filename);
    const pathArtifact = join(ARTIFACT_SCREENSHOTS_DIR, filename);
    writeFileSync(pathDocs, buffer);
    writeFileSync(pathArtifact, buffer);
    console.log(`[OK] Screenshot salvato: ${filename} (${buffer.length} bytes)`);
  }

  await send('Page.enable');
  await send('Runtime.enable');
  await send('DOM.enable');

  const BASE_URL = 'http://localhost:3080';

  console.log('1. Navigazione Login...');
  await send('Page.navigate', { url: `${BASE_URL}/login` });
  await sleep(1500);
  await takeScreenshot('01_login.png');

  console.log('2. Esecuzione Login...');
  await evalJs(`
    (function() {
      const emailInput = document.querySelector('input[type="email"]') || document.querySelectorAll('input')[0];
      const passInput = document.querySelector('input[type="password"]') || document.querySelectorAll('input')[1];
      const form = document.querySelector('form');
      const submitBtn = document.querySelector('button[type="submit"]') || (form && form.querySelector('button'));

      if (emailInput && passInput) {
        const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value").set;
        setter.call(emailInput, 'admin@nextgen.local');
        emailInput.dispatchEvent(new Event('input', { bubbles: true }));

        setter.call(passInput, 'AdminPass123!');
        passInput.dispatchEvent(new Event('input', { bubbles: true }));

        if (submitBtn) submitBtn.click();
        else if (form) form.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
      }
    })();
  `);
  await sleep(2500);

  console.log('3. Dashboard...');
  await takeScreenshot('02_dashboard.png');

  console.log('4. Gestione Aziende (/companies)...');
  await send('Page.navigate', { url: `${BASE_URL}/companies` });
  await sleep(2000);
  await takeScreenshot('03_companies.png');

  // Apri modale Nuova Azienda se c'è pulsante
  console.log('5. Modale Nuova Azienda...');
  await evalJs(`
    (function() {
      const btns = Array.from(document.querySelectorAll('button'));
      const btn = btns.find(b => b.innerText && (b.innerText.includes('Nuova Azienda') || b.innerText.includes('Aggiungi Azienda')));
      if (btn) btn.click();
    })();
  `);
  await sleep(1000);
  await takeScreenshot('04_company_modal.png');

  // Chiudi modale (premi Esc o annulla)
  await evalJs(`
    (function() {
      const btns = Array.from(document.querySelectorAll('button'));
      const cancelBtn = btns.find(b => b.innerText && (b.innerText.includes('Annulla') || b.innerText.includes('Chiudi')));
      if (cancelBtn) cancelBtn.click();
    })();
  `);
  await sleep(500);

  console.log('6. Gestione Target & Reparti (/targets)...');
  await send('Page.navigate', { url: `${BASE_URL}/targets` });
  await sleep(2000);
  await takeScreenshot('05_targets.png');

  // Apri modale nuovo target
  await evalJs(`
    (function() {
      const btns = Array.from(document.querySelectorAll('button'));
      const btn = btns.find(b => b.innerText && (b.innerText.includes('Aggiungi Target') || b.innerText.includes('Nuovo Target')));
      if (btn) btn.click();
    })();
  `);
  await sleep(1000);
  await takeScreenshot('06_target_modal.png');

  // Chiudi modale
  await evalJs(`
    (function() {
      const btns = Array.from(document.querySelectorAll('button'));
      const cancelBtn = btns.find(b => b.innerText && (b.innerText.includes('Annulla') || b.innerText.includes('Chiudi')));
      if (cancelBtn) cancelBtn.click();
    })();
  `);
  await sleep(500);

  console.log('7. Network Discovery (/network-discovery)...');
  await send('Page.navigate', { url: `${BASE_URL}/network-discovery` });
  await sleep(2000);
  await takeScreenshot('07_network_discovery.png');

  console.log('8. Scansioni (/scans)...');
  await send('Page.navigate', { url: `${BASE_URL}/scans` });
  await sleep(2000);
  await takeScreenshot('08_scans.png');

  // Apri modale nuova scansione
  await evalJs(`
    (function() {
      const btns = Array.from(document.querySelectorAll('button'));
      const btn = btns.find(b => b.innerText && (b.innerText.includes('Nuova Scansione') || b.innerText.includes('New Scan')));
      if (btn) btn.click();
    })();
  `);
  await sleep(1200);
  await takeScreenshot('09_new_scan_modal.png');

  // Chiudi modale nuova scansione
  await evalJs(`
    (function() {
      const btns = Array.from(document.querySelectorAll('button'));
      const cancelBtn = btns.find(b => b.innerText && (b.innerText.includes('Annulla') || b.innerText.includes('Chiudi')));
      if (cancelBtn) cancelBtn.click();
    })();
  `);
  await sleep(500);

  console.log('9. Vulnerabilità Findings (/findings)...');
  await send('Page.navigate', { url: `${BASE_URL}/findings` });
  await sleep(2000);
  await takeScreenshot('10_findings.png');

  console.log('10. Attack Path Mapping (/attack-path)...');
  await send('Page.navigate', { url: `${BASE_URL}/attack-path` });
  await sleep(2500);
  await takeScreenshot('11_attack_path.png');

  // Clicca tab Chokepoints in Attack Path se presente
  await evalJs(`
    (function() {
      const btns = Array.from(document.querySelectorAll('button'));
      const chkBtn = btns.find(b => b.innerText && b.innerText.toLowerCase().includes('chokepoint'));
      if (chkBtn) chkBtn.click();
    })();
  `);
  await sleep(1000);
  await takeScreenshot('12_attack_path_chokepoints.png');

  console.log('11. SSH Servers (/ssh-servers)...');
  await send('Page.navigate', { url: `${BASE_URL}/ssh-servers` });
  await sleep(2000);
  await takeScreenshot('13_ssh_servers.png');

  // Apri modale Nuovo Server SSH
  await evalJs(`
    (function() {
      const btns = Array.from(document.querySelectorAll('button'));
      const btn = btns.find(b => b.innerText && (b.innerText.includes('Aggiungi Server') || b.innerText.includes('Nuovo Server')));
      if (btn) btn.click();
    })();
  `);
  await sleep(1000);
  await takeScreenshot('14_ssh_modal.png');

  await evalJs(`
    (function() {
      const btns = Array.from(document.querySelectorAll('button'));
      const cancelBtn = btns.find(b => b.innerText && (b.innerText.includes('Annulla') || b.innerText.includes('Chiudi')));
      if (cancelBtn) cancelBtn.click();
    })();
  `);
  await sleep(500);

  console.log('12. Reports Hub (/reports)...');
  await send('Page.navigate', { url: `${BASE_URL}/reports` });
  await sleep(2000);
  await takeScreenshot('15_reports.png');

  console.log('13. Settings (/settings)...');
  await send('Page.navigate', { url: `${BASE_URL}/settings` });
  await sleep(2000);
  await takeScreenshot('16_settings.png');

  console.log('14. Live Debug Console Drawer (>_)...');
  await evalJs(`
    (function() {
      const btns = Array.from(document.querySelectorAll('button'));
      // Il pulsante fluttuante ha icona Terminal o testo '>_' o title 'Debug'
      const debugBtn = btns.find(b => (b.innerText && b.innerText.includes('>_')) || (b.getAttribute('title') && b.getAttribute('title').toLowerCase().includes('debug')) || b.className.includes('fixed bottom-'));
      if (debugBtn) debugBtn.click();
    })();
  `);
  await sleep(1000);
  await takeScreenshot('17_debug_console.png');

  console.log('Tutti gli screenshot sono stati acquisiti con successo!');
  ws.close();
  chrome.kill();
}

main().catch(err => {
  console.error('Errore durante la cattura screenshot:', err);
  process.exit(1);
});

