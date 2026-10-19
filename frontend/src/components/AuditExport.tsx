import { useState } from 'react';
import { downloadReportFile } from '../utils/download';

export function AuditExport({ companyId, companyName, departmentId = '' }: { companyId: string; companyName: string; departmentId?: string }) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  return <details className="assessment-panel rounded-xl border border-slate-700 p-4">
    <summary className="cursor-pointer font-semibold">Dossier Audit PDF A4</summary>
    <p className="my-3 text-sm text-slate-400">Checklist, evidenze, storico e piano residuo dello scope corrente. Documento per revisione, senza certificazione o firma digitale. Il periodo è dichiarativo; i dati rappresentano lo stato al momento dell’esportazione.</p>
    <form className="grid gap-3" onSubmit={async event => {
      event.preventDefault(); setBusy(true); setError('');
      const form = new FormData(event.currentTarget);
      await downloadReportFile({ url: '/reports/audit', method: 'post', defaultFilename: 'assessment_audit.pdf', data: { scope: { company_id: companyId, department_id: departmentId || '00000000-0000-0000-0000-000000000000' }, metadata: { legal_entity: form.get('legal_entity'), legal_details: form.get('legal_details'), signatory: form.get('signatory'), period: form.get('period'), audit_date: `${form.get('audit_date')}T00:00:00Z` } }, onError: err => setError(err.message) });
      setBusy(false);
    }}>
      <label>Ragione sociale<input className="input w-full" name="legal_entity" defaultValue={companyName} required maxLength={500} /></label>
      <label>Dati legali e sede<textarea className="input w-full" name="legal_details" maxLength={2000} /></label>
      <label>Revisore / firmatario dichiarato (facoltativo)<input className="input w-full" name="signatory" maxLength={200} /></label>
      <label>Periodo di revisione dichiarato<input className="input w-full" name="period" required maxLength={200} placeholder="Es. ottobre 2026" /></label>
      <label>Data audit<input className="input w-full" name="audit_date" type="date" defaultValue={new Date().toISOString().slice(0, 10)} required /></label>
      {error && <p role="alert" className="text-red-400">{error}</p>}
      <button className="btn-primary justify-self-start" disabled={busy}>{busy ? 'Generazione…' : 'Scarica dossier Audit'}</button>
    </form>
  </details>;
}
