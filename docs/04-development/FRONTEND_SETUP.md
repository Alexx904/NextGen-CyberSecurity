# Frontend Setup

> Responsabile: Lead Developer. Baseline verificata nei manifest e nella configurazione Vite.

Node.js 20 è la versione CI; installazione ripetibile con lockfile npm. Dalla directory `frontend`:

```powershell
npm ci
npm run dev
```

Vite espone la UI sulla porta 3000 e inoltra `/api` al backend locale sulla porta 8080. Configurazione in `vite.config.ts`; non assumere che questo proxy configuri anche un deployment di produzione.

- `npm test`: Vitest in modalità run, ambiente happy-dom.
- `npm run build`: TypeScript e bundle Vite.
- `npm run preview`: anteprima del bundle.
- `npm run test:watch`: sviluppo interattivo.

React Router definisce le route in `src/App.tsx`. AuthContext, CompanyContext e DebugContext gestiscono i contesti trasversali; le pagine sono caricate con React.lazy/Suspense. Componenti, hook e client API devono mantenere la coerenza dei tipi e i confini del dominio.

Nessuno script lint frontend è dichiarato nel manifest corrente: non riportare un lint PASS senza un comando esistente. Vedi [UI](../02-design/UI_SPECIFICATIONS.md) e [test plan](../05-qa/TEST_PLAN.md).
