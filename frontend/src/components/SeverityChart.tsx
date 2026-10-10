import React from 'react';
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip } from 'recharts';
import { ShieldCheck, AlertTriangle } from 'lucide-react';

interface SeverityChartProps {
  critical: number;
  high: number;
  medium: number;
  low: number;
  info: number;
}

const SEVERITY_COLORS = {
  Critical: '#dc2626',
  High: '#ea580c',
  Medium: '#eab308',
  Low: '#3b82f6',
  Info: '#64748b',
};

export const SeverityChart: React.FC<SeverityChartProps> = ({
  critical,
  high,
  medium,
  low,
  info,
}) => {
  const total = critical + high + medium + low + info;

  const data = [
    { name: 'Critical', value: critical, color: SEVERITY_COLORS.Critical },
    { name: 'High', value: high, color: SEVERITY_COLORS.High },
    { name: 'Medium', value: medium, color: SEVERITY_COLORS.Medium },
    { name: 'Low', value: low, color: SEVERITY_COLORS.Low },
    { name: 'Info', value: info, color: SEVERITY_COLORS.Info },
  ].filter((item) => item.value > 0);

  if (total === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-64 text-center p-6 rounded-xl border border-dashed border-cyber-border bg-slate-900/30">
        <div className="w-12 h-12 rounded-xl bg-emerald-500/10 border border-emerald-500/30 flex items-center justify-center text-emerald-400 mb-3 shadow-[0_0_15px_rgba(16,185,129,0.15)]">
          <ShieldCheck className="w-6 h-6" />
        </div>
        <div className="text-sm font-semibold text-slate-200">Nessuna Vulnerabilità Rilevata</div>
        <p className="text-xs text-slate-400 mt-1 max-w-xs">
          Le scansioni completate non hanno registrato vulnerabilità note nei servizi esposti.
        </p>
      </div>
    );
  }

  const CustomTooltip = ({ active, payload }: any) => {
    if (active && payload && payload.length) {
      const item = payload[0];
      const percentage = ((item.value / total) * 100).toFixed(1);
      return (
        <div className="p-3 rounded-lg border border-cyber-border bg-[#111726] shadow-xl text-xs space-y-1">
          <div className="flex items-center space-x-2">
            <span
              className="w-2.5 h-2.5 rounded-full"
              style={{ backgroundColor: item.payload.color }}
            />
            <span className="font-semibold text-white">{item.name}</span>
          </div>
          <div className="text-slate-300">
            Conteggio: <strong className="text-white">{item.value}</strong>
          </div>
          <div className="text-slate-400 text-[10px]">
            {percentage}% del totale vulnerabilità
          </div>
        </div>
      );
    }
    return null;
  };

  const severityItems = [
    { label: 'Critical', count: critical, color: SEVERITY_COLORS.Critical, bg: 'bg-red-500/10 text-red-400 border-red-500/30' },
    { label: 'High', count: high, color: SEVERITY_COLORS.High, bg: 'bg-orange-500/10 text-orange-400 border-orange-500/30' },
    { label: 'Medium', count: medium, color: SEVERITY_COLORS.Medium, bg: 'bg-yellow-500/10 text-yellow-400 border-yellow-500/30' },
    { label: 'Low', count: low, color: SEVERITY_COLORS.Low, bg: 'bg-blue-500/10 text-blue-400 border-blue-500/30' },
    { label: 'Info', count: info, color: SEVERITY_COLORS.Info, bg: 'bg-slate-500/10 text-slate-400 border-slate-500/30' },
  ];

  return (
    <div className="grid grid-cols-1 md:grid-cols-12 gap-6 items-center">
      {/* Donut Chart */}
      <div className="md:col-span-6 h-64 relative flex items-center justify-center">
        <ResponsiveContainer width="100%" height="100%" minWidth={100} minHeight={100}>
          <PieChart>
            <Tooltip content={<CustomTooltip />} />
            <Pie
              data={data}
              innerRadius={65}
              outerRadius={95}
              paddingAngle={3}
              dataKey="value"
              stroke="#0a0d14"
              strokeWidth={2}
            >
              {data.map((entry, index) => (
                <Cell key={`cell-${index}`} fill={entry.color} />
              ))}
            </Pie>
          </PieChart>
        </ResponsiveContainer>

        {/* Center label */}
        <div className="absolute inset-0 flex flex-col items-center justify-center pointer-events-none">
          <span className="text-2xl font-bold text-white tracking-tight">{total}</span>
          <span className="text-[11px] text-slate-400 uppercase tracking-wider">Totale CVE</span>
        </div>
      </div>

      {/* Breakdown List */}
      <div className="md:col-span-6 space-y-2.5">
        {severityItems.map((item) => {
          const percentage = total > 0 ? (item.count / total) * 100 : 0;
          return (
            <div key={item.label} className="space-y-1">
              <div className="flex items-center justify-between text-xs">
                <div className="flex items-center space-x-2">
                  <span
                    className="w-2.5 h-2.5 rounded-full"
                    style={{ backgroundColor: item.color }}
                  />
                  <span className="text-slate-300 font-medium">{item.label}</span>
                </div>
                <div className="flex items-center space-x-2">
                  <span className="text-slate-400 text-[11px]">
                    {percentage.toFixed(0)}%
                  </span>
                  <span className="font-semibold text-white min-w-6 text-right">
                    {item.count}
                  </span>
                </div>
              </div>
              <div className="w-full h-1.5 rounded-full bg-slate-800 overflow-hidden">
                <div
                  className="h-full rounded-full transition-all duration-500"
                  style={{
                    width: `${percentage}%`,
                    backgroundColor: item.color,
                  }}
                />
              </div>
            </div>
          );
        })}

        {critical + high > 0 && (
          <div className="mt-3 p-2.5 rounded-lg border border-red-900/40 bg-red-950/20 text-red-300 text-[11px] flex items-center space-x-2">
            <AlertTriangle className="w-4 h-4 shrink-0 text-red-400" />
            <span>
              <strong>{critical + high}</strong> vulnerabilità richiedono attenzione prioritaria di remediation.
            </span>
          </div>
        )}
      </div>
    </div>
  );
};

