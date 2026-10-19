# Frontend Setup

> Responsabile: Lead Developer. Baseline verificata nei manifest e nella configurazione Vite.

Node.js 22 (minimum 22.12) is the CI/Docker build baseline; Node 24 is also supported. Install reproducibly with the npm lockfile. From `frontend`:

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

## Dependency baseline (2026-10-18)

The toolchain uses Vite 7, React plugin 5, Vitest/coverage 4.1.11 and Tailwind 4
with the official `@tailwindcss/postcss` plugin. Autoprefixer is no longer needed.
React stays on 18; React Router uses 7. Tailwind sources are declared in
`src/index.css`, which explicitly loads the custom theme from `tailwind.config.js`.
Browser minimums are Safari 16.4, Chrome 111 and Firefox 128. See the
[dependency assessment](../05-qa/TECH-DEBT-09.md) for audit and compatibility gates.
