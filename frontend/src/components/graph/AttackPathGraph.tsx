import React, { useMemo, useRef, useState, useId } from 'react';
import { ZoomIn, ZoomOut, Maximize2, Network, Route } from 'lucide-react';
import type { AttackGraph, AttackNode, AttackPath } from '../../types/attackPath';
import { layoutAttackGraph } from './attackPathModel';

interface AttackPathGraphProps {
  graph: AttackGraph;
  selectedPathId?: string | null;
  selectedNodeId?: string | null;
  onSelectNode?: (node: AttackNode | null) => void;
  onSelectPath?: (path: AttackPath | null) => void;
  className?: string;
}

const colors = { ENTRYPOINT: '#818cf8', HOST: '#94a3b8', SUBNET: '#a78bfa', SERVICE: '#22d3ee', VULNERABILITY: '#fb7185' };
const names = { ENTRYPOINT: 'ORIGINE IPOTETICA', HOST: 'HOST', SUBNET: 'RETE', SERVICE: 'SERVIZIO', VULNERABILITY: 'CVE CORRELATA' };
const width = 210;
const height = 74;

export const AttackPathGraph: React.FC<AttackPathGraphProps> = ({
  graph, selectedPathId, selectedNodeId, onSelectNode, onSelectPath, className = '',
}) => {
  const [pathOnly, setPathOnly] = useState(true);
  const [zoom, setZoom] = useState(1);
  const [pan, setPan] = useState({ x: 0, y: 0 });
  const drag = useRef<{ x: number; y: number; px: number; py: number } | null>(null);
  const svg = useRef<SVGSVGElement>(null);
  const markerId = useId().replace(/:/g, '');
  const activePath = graph.paths.find(p => p.id === selectedPathId);
  const focusedPath = pathOnly ? activePath : null;
  const positions = useMemo(() => layoutAttackGraph(graph, focusedPath), [graph, focusedPath]);
  const map = useMemo(() => new Map(positions.map(n => [n.id, n])), [positions]);
  const pathNodes = new Set(activePath?.node_ids || []);
  const pathEdges = new Set(activePath?.edge_ids || []);
  const chokepoints = new Set(graph.chokepoints.map(c => c.node_id));
  const visibleEdges = graph.edges.filter(e => map.has(e.source) && map.has(e.target) && (!focusedPath || pathEdges.has(e.id)));
  const bounds = {
    width: Math.max(500, ...positions.map(n => n.x + width + 40)),
    height: Math.max(focusedPath ? 310 : 400, ...positions.map(n => n.y + height + 60)),
  };
  const reset = () => { setZoom(1); setPan({ x: 0, y: 0 }); };

  return (
    <div className={`rounded-xl border border-slate-700 bg-slate-950 overflow-hidden ${className}`}>
      <div className="flex flex-wrap items-center justify-between gap-2 p-3 border-b border-slate-800 bg-slate-900">
        <div className="flex items-center gap-1 text-xs">
          <button type="button" aria-pressed={pathOnly} onClick={() => { setPathOnly(true); reset(); }}
            className={`px-3 py-2 rounded-lg flex items-center gap-2 ${pathOnly ? 'bg-cyan-500/20 text-cyan-200' : 'text-slate-400'}`}>
            <Route size={14} />Vista percorso
          </button>
          <button type="button" aria-pressed={!pathOnly} onClick={() => { setPathOnly(false); reset(); }}
            className={`px-3 py-2 rounded-lg flex items-center gap-2 ${!pathOnly ? 'bg-cyan-500/20 text-cyan-200' : 'text-slate-400'}`}>
            <Network size={14} />Mappa completa
          </button>
        </div>
        <div className="flex items-center gap-2 text-slate-300">
          <button type="button" aria-label="Zoom Indietro" onClick={() => setZoom(z => Math.max(.4, z - .2))}><ZoomOut size={17} /></button>
          <span className="text-xs font-mono w-10 text-center">{Math.round(zoom * 100)}%</span>
          <button type="button" aria-label="Zoom Avanti" onClick={() => setZoom(z => Math.min(3, z + .2))}><ZoomIn size={17} /></button>
          <button type="button" aria-label="Adatta vista" title="Adatta vista" onClick={reset} className="ml-2"><Maximize2 size={17} /></button>
        </div>
      </div>
      <div className="px-4 py-2 text-xs text-slate-400 border-b border-slate-800 flex justify-between gap-3">
        <span>{focusedPath ? 'Leggi da sinistra a destra · scorri in orizzontale per seguire i passaggi' : 'Contesto completo · seleziona un percorso per ridurre gli incroci'}</span>
        {(activePath || selectedNodeId) && <button type="button" onClick={() => { onSelectPath?.(null); onSelectNode?.(null); reset(); }} className="text-cyan-300 shrink-0">Azzera selezione</button>}
      </div>
      <div className="overflow-x-auto" tabIndex={0} role="region" aria-label="Area scorrevole del grafo">
      <svg ref={svg} role="group" aria-label="Grafo dei percorsi di analisi"
        style={{ minWidth: focusedPath ? bounds.width : 960 }}
        className={`w-full touch-none ${focusedPath ? 'h-[340px]' : 'h-[540px]'}`}
        viewBox={`0 0 ${bounds.width} ${bounds.height}`}
        onPointerDown={e => {
          if (e.button !== 0 || (e.target as Element).closest('[role="button"]')) return;
          drag.current = { x: e.clientX, y: e.clientY, px: pan.x, py: pan.y };
          e.currentTarget.setPointerCapture(e.pointerId);
        }}
        onPointerMove={e => {
          if (!drag.current) return;
          const rect = svg.current!.getBoundingClientRect();
          // getScreenCTM accounts for viewBox letterboxing, zoom and viewport scaling.
          const matrix = svg.current?.getScreenCTM();
          const sx = matrix?.a || rect.width / bounds.width;
          const sy = matrix?.d || rect.height / bounds.height;
          setPan({ x: drag.current.px + (e.clientX - drag.current.x) / sx, y: drag.current.py + (e.clientY - drag.current.y) / sy });
        }}
        onPointerUp={() => { drag.current = null; }}
        onPointerCancel={() => { drag.current = null; }}>
        <title>Associazioni dei finding e ipotesi di accesso. Non prova di compromissione.</title>
        <defs>
          <marker id={markerId} viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse">
            <path d="M 0 1 L 10 5 L 0 9 z" fill="context-stroke" />
          </marker>
          <pattern id={`${markerId}-grid`} width="28" height="28" patternUnits="userSpaceOnUse">
            <circle cx="1" cy="1" r=".8" fill="#334155" />
          </pattern>
        </defs>
        <rect width="100%" height="100%" fill={`url(#${markerId}-grid)`} />
        <g transform={`translate(${pan.x},${pan.y}) scale(${zoom})`}>
          {visibleEdges.map(edge => {
            const a = map.get(edge.source)!;
            const b = map.get(edge.target)!;
            const active = pathEdges.has(edge.id);
            const backward = b.x <= a.x;
            const x1 = a.x + width, y1 = a.y + height / 2, x2 = b.x, y2 = b.y + height / 2;
            const d = backward
              ? `M ${x1} ${y1} C ${x1 + 80} ${y1 + 90}, ${x2 - 80} ${y2 + 90}, ${x2} ${y2}`
              : `M ${x1} ${y1} C ${x1 + 45} ${y1}, ${x2 - 45} ${y2}, ${x2} ${y2}`;
            const color = edge.type === 'PIVOTS_TO' ? '#fb923c' : edge.inferred ? '#a78bfa' : '#22d3ee';
            return (
              <g key={edge.id} opacity={!focusedPath && activePath && !active ? .22 : 1}>
                <title>{edge.label}{edge.inference_basis ? ': ' + edge.inference_basis : ''}</title>
                <path data-edge-id={edge.id} d={d} fill="none" stroke={color} strokeWidth={active ? 3 : 1.7}
                  strokeDasharray={edge.inferred ? '7 5' : undefined} markerEnd={`url(#${markerId})`} />
                {(focusedPath || !activePath && graph.edges.length < 16) && (
                  <g>
                    <rect x={(x1 + x2) / 2 - 59} y={(y1 + y2) / 2 - 37} width="118" height="22" rx="5" fill="#0f172a" />
                    <text x={(x1 + x2) / 2} y={(y1 + y2) / 2 - 23} fill={color} textAnchor="middle" fontSize="10">
                      {edge.type === 'PIVOTS_TO' ? 'IPOTESI /24' : edge.inferred ? 'DA VERIFICARE' : 'OSSERVATO'}
                    </text>
                  </g>
                )}
              </g>
            );
          })}
          {positions.map((node, i) => {
            const selected = node.id === selectedNodeId;
            const cp = chokepoints.has(node.id);
            const color = colors[node.type];
            return (
              <g key={node.id} role="button" tabIndex={0} aria-label={`Ispeziona ${node.label}`} aria-pressed={selected}
                onClick={() => onSelectNode?.(node)}
                onKeyDown={e => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onSelectNode?.(node); } }}
                opacity={!focusedPath && activePath && !pathNodes.has(node.id) ? .25 : 1}
                className="cursor-pointer outline-hidden focus:stroke-white">
                <title>{node.label}{node.ip ? ' · ' + node.ip : ''}{cp ? ' · Chokepoint del modello' : ''}</title>
                {focusedPath && <text x={node.x} y={node.y - 15} fontSize="11" fill="#94a3b8">{i + 1}. {names[node.type]}</text>}
                <rect x={node.x} y={node.y} width={width} height={height} rx="12" fill="#0f172a"
                  stroke={selected ? '#ffffff' : cp ? '#fbbf24' : color} strokeWidth={selected ? 3 : 1.5} />
                <rect x={node.x} y={node.y + 13} width="4" height={height - 26} rx="2" fill={color} />
                <text x={node.x + 15} y={node.y + 20} fill={color} fontSize="9" letterSpacing="1">{names[node.type]}{cp ? ' · CHOKE' : ''}</text>
                <text x={node.x + 15} y={node.y + 40} fill="#f1f5f9" fontSize="12" fontWeight="600">
                  {node.label.length > 26 ? node.label.slice(0, 25) + '…' : node.label}
                </text>
                <text x={node.x + 15} y={node.y + 59} fill="#94a3b8" fontSize="10">
                  {node.type === 'SERVICE' ? `${node.port}/${node.protocol}` : node.type === 'VULNERABILITY' ? `CVSS ${(node.cvss || 0).toFixed(1)} · non prova di exploit` : node.ip || 'Nessuna esposizione presunta'}
                </text>
              </g>
            );
          })}
        </g>
      </svg>
      </div>
      <div className="flex flex-wrap items-center gap-x-5 gap-y-2 px-4 py-3 border-t border-slate-800 text-xs text-slate-300">
        <span><span className="text-cyan-300">━━</span> Associazione osservata</span>
        <span><span className="text-violet-300">┄┄</span> Ipotesi / correlazione da verificare</span>
        <span><span className="text-orange-300">┄➜</span> Possibile accesso laterale</span>
        <span className="text-amber-300">▣ Chokepoint del modello</span>
      </div>
    </div>
  );
};
