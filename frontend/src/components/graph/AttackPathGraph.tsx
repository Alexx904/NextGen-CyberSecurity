import React, { useState, useRef, useMemo } from 'react';
import {
  Globe,
  Server,
  AlertTriangle,
  Radio,
  Network,
  ZoomIn,
  ZoomOut,
  RotateCcw,
  Flame,
} from 'lucide-react';
import {
  AttackGraph,
  AttackNode,
  AttackPath,
} from '../../types/attackPath';

interface AttackPathGraphProps {
  graph: AttackGraph;
  selectedPathId?: string | null;
  selectedNodeId?: string | null;
  onSelectNode?: (node: AttackNode | null) => void;
  onSelectPath?: (path: AttackPath | null) => void;
  className?: string;
}

interface PositionedNode extends AttackNode {
  x: number;
  y: number;
  width: number;
  height: number;
  isChokepoint: boolean;
}

const NODE_WIDTH = 210;
const NODE_HEIGHT = 64;
const COL_SPACING = 270;
const ROW_SPACING = 96;

export const AttackPathGraph: React.FC<AttackPathGraphProps> = ({
  graph,
  selectedPathId,
  selectedNodeId,
  onSelectNode,
  onSelectPath,
  className = '',
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const [zoom, setZoom] = useState(1);
  const [pan, setPan] = useState({ x: 40, y: 40 });
  const [isDragging, setIsDragging] = useState(false);
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 });

  // Insieme di ID chokepoint per highlight rapido
  const chokepointSet = useMemo(() => {
    return new Set(graph.chokepoints?.map((c) => c.node_id) || []);
  }, [graph.chokepoints]);

  // Nodi ed archi appartenenti al percorso attualmente selezionato
  const activePath = useMemo(() => {
    if (!selectedPathId) return null;
    return graph.paths?.find((p) => p.id === selectedPathId) || null;
  }, [selectedPathId, graph.paths]);

  const activePathNodeIds = useMemo(() => {
    return new Set(activePath?.node_ids || []);
  }, [activePath]);

  const activePathEdgeIds = useMemo(() => {
    return new Set(activePath?.edge_ids || []);
  }, [activePath]);

  // Calcolo topologico posizioni nodi per strati (layers)
  const { positionedNodes, nodeMap, canvasBounds } = useMemo(() => {
    const map = new Map<string, PositionedNode>();
    const nodes = graph.nodes || [];
    const edges = graph.edges || [];

    // Assegnazione layer in base alla natura e connessioni
    // Layer 0: ENTRYPOINT
    // Layer 1: HOST connessi all'entrypoint
    // Layer 2: SERVICE & VULNERABILITY di Layer 1
    // Layer 3: HOST pivotati / interni
    // Layer 4: SERVICE & VULNERABILITY interni
    const layerMap = new Map<string, number>();

    // 1. Inizializza layer per entrypoint
    nodes.forEach((n) => {
      if (n.type === 'ENTRYPOINT') {
        layerMap.set(n.id, 0);
      }
    });

    // 2. BFS per distribuire i nodi nei layer successivi
    const adj = new Map<string, string[]>();
    edges.forEach((e) => {
      const list = adj.get(e.source) || [];
      list.push(e.target);
      adj.set(e.source, list);
    });

    const queue: { id: string; layer: number }[] = [];
    nodes.forEach((n) => {
      if (n.type === 'ENTRYPOINT') {
        queue.push({ id: n.id, layer: 0 });
      }
    });

    while (queue.length > 0) {
      const { id, layer } = queue.shift()!;
      const neighbors = adj.get(id) || [];
      for (const nextId of neighbors) {
        const curLayer = layerMap.get(nextId);
        const targetNode = nodes.find((n) => n.id === nextId);

        // Nodi vulnerability/service subito dopo il loro host
        const step = targetNode?.type === 'VULNERABILITY' || targetNode?.type === 'SERVICE' ? 1 : 1;
        const newLayer = Math.max(curLayer ?? 0, layer + step);

        if (curLayer === undefined || newLayer > curLayer) {
          layerMap.set(nextId, newLayer);
          queue.push({ id: nextId, layer: newLayer });
        }
      }
    }

    // Nodi non raggiunti dall'entrypoint: assegna layer in base al tipo
    nodes.forEach((n) => {
      if (!layerMap.has(n.id)) {
        if (n.type === 'ENTRYPOINT') layerMap.set(n.id, 0);
        else if (n.type === 'HOST') layerMap.set(n.id, 1);
        else if (n.type === 'SERVICE') layerMap.set(n.id, 2);
        else if (n.type === 'VULNERABILITY') layerMap.set(n.id, 2);
        else layerMap.set(n.id, 3);
      }
    });

    // Raggruppa nodi per layer
    const columns = new Map<number, AttackNode[]>();
    nodes.forEach((n) => {
      const layer = layerMap.get(n.id) || 1;
      const list = columns.get(layer) || [];
      list.push(n);
      columns.set(layer, list);
    });

    const sortedLayers = Array.from(columns.keys()).sort((a, b) => a - b);
    let maxX = 800;
    let maxY = 600;

    const list: PositionedNode[] = [];

    sortedLayers.forEach((layerIndex, colIdx) => {
      const colNodes = columns.get(layerIndex) || [];
      const colX = 60 + colIdx * COL_SPACING;

      colNodes.forEach((node, rowIdx) => {
        const nodeY = 60 + rowIdx * ROW_SPACING;
        const posNode: PositionedNode = {
          ...node,
          x: colX,
          y: nodeY,
          width: NODE_WIDTH,
          height: NODE_HEIGHT,
          isChokepoint: chokepointSet.has(node.id),
        };
        list.push(posNode);
        map.set(node.id, posNode);

        if (colX + NODE_WIDTH + 100 > maxX) maxX = colX + NODE_WIDTH + 100;
        if (nodeY + NODE_HEIGHT + 100 > maxY) maxY = nodeY + NODE_HEIGHT + 100;
      });
    });

    return {
      positionedNodes: list,
      nodeMap: map,
      canvasBounds: { width: maxX, height: maxY },
    };
  }, [graph, chokepointSet]);

  // Gestione Pan & Zoom con mouse
  const handleMouseDown = (e: React.MouseEvent) => {
    // Solo click primario su sfondo SVG
    if (e.button !== 0) return;
    setIsDragging(true);
    setDragStart({ x: e.clientX - pan.x, y: e.clientY - pan.y });
  };

  const handleMouseMove = (e: React.MouseEvent) => {
    if (!isDragging) return;
    setPan({
      x: e.clientX - dragStart.x,
      y: e.clientY - dragStart.y,
    });
  };

  const handleMouseUp = () => {
    setIsDragging(false);
  };

  const handleZoomIn = () => setZoom((z) => Math.min(z + 0.15, 2.2));
  const handleZoomOut = () => setZoom((z) => Math.max(z - 0.15, 0.4));
  const handleResetView = () => {
    setZoom(1);
    setPan({ x: 40, y: 40 });
  };

  // Helper grafici per stili e icone in base alla tipologia di nodo
  const getNodeColor = (node: PositionedNode) => {
    if (node.type === 'ENTRYPOINT') {
      return {
        border: 'stroke-indigo-500',
        bg: 'fill-slate-900',
        text: 'text-indigo-400',
        badgeBg: 'bg-indigo-500/20 text-indigo-300 border-indigo-500/30',
        icon: Globe,
      };
    }
    if (node.type === 'VULNERABILITY') {
      const isCrit = node.severity === 'CRITICAL' || (node.cvss && node.cvss >= 9.0);
      return {
        border: isCrit ? 'stroke-rose-500' : 'stroke-amber-500',
        bg: 'fill-slate-900',
        text: isCrit ? 'text-rose-400' : 'text-amber-400',
        badgeBg: isCrit
          ? 'bg-rose-500/20 text-rose-300 border-rose-500/40'
          : 'bg-amber-500/20 text-amber-300 border-amber-500/40',
        icon: isCrit ? Flame : AlertTriangle,
      };
    }
    if (node.type === 'SERVICE') {
      return {
        border: 'stroke-cyan-500',
        bg: 'fill-slate-900',
        text: 'text-cyan-400',
        badgeBg: 'bg-cyan-500/20 text-cyan-300 border-cyan-500/30',
        icon: Radio,
      };
    }
    // Default: HOST o SUBNET
    const isCrit = node.severity === 'CRITICAL';
    const isHigh = node.severity === 'HIGH';
    return {
      border: isCrit ? 'stroke-rose-500' : isHigh ? 'stroke-amber-500' : 'stroke-slate-600',
      bg: 'fill-slate-900',
      text: isCrit ? 'text-rose-400' : isHigh ? 'text-amber-400' : 'text-slate-300',
      badgeBg: isCrit
        ? 'bg-rose-500/20 text-rose-300 border-rose-500/30'
        : 'bg-slate-800 text-slate-300 border-slate-700',
      icon: node.type === 'SUBNET' ? Network : Server,
    };
  };

  return (
    <div
      ref={containerRef}
      className={`relative w-full h-[640px] bg-slate-950 border border-slate-800 rounded-xl overflow-hidden select-none ${className}`}
      onMouseDown={handleMouseDown}
      onMouseMove={handleMouseMove}
      onMouseUp={handleMouseUp}
      onMouseLeave={handleMouseUp}
      style={{ cursor: isDragging ? 'grabbing' : 'grab' }}
    >
      {/* Controlli flottanti di Zoom e Reset */}
      <div className="absolute top-4 right-4 z-20 flex items-center gap-1.5 bg-slate-900/90 backdrop-blur-md border border-slate-700/80 rounded-lg p-1 shadow-xl">
        <button
          onClick={handleZoomIn}
          title="Zoom Avanti"
          aria-label="Zoom Avanti"
          className="p-1.5 text-slate-400 hover:text-white hover:bg-slate-800 rounded transition"
        >
          <ZoomIn className="w-4 h-4" />
        </button>
        <button
          onClick={handleZoomOut}
          title="Zoom Indietro"
          aria-label="Zoom Indietro"
          className="p-1.5 text-slate-400 hover:text-white hover:bg-slate-800 rounded transition"
        >
          <ZoomOut className="w-4 h-4" />
        </button>
        <div className="w-[1px] h-4 bg-slate-700 mx-1" />
        <button
          onClick={handleResetView}
          title="Centra e Reimposta Vista"
          aria-label="Centra e Reimposta Vista"
          className="flex items-center gap-1.5 px-2 py-1 text-xs font-mono text-slate-300 hover:text-white hover:bg-slate-800 rounded transition"
        >
          <RotateCcw className="w-3.5 h-3.5" />
          <span>{Math.round(zoom * 100)}%</span>
        </button>
      </div>

      {/* Indicatore di selezione attiva */}
      {(selectedPathId || selectedNodeId) && (
        <div className="absolute top-4 left-4 z-20 flex items-center gap-2 bg-slate-900/90 backdrop-blur-md border border-cyan-500/40 rounded-lg px-3 py-1.5 shadow-xl text-xs text-slate-200">
          <span className="w-2 h-2 rounded-full bg-cyan-400 animate-pulse" />
          <span>
            {selectedPathId ? 'Percorso Evidenziato' : 'Nodo Selezionato'}
          </span>
          <button
            onClick={() => {
              onSelectPath?.(null);
              onSelectNode?.(null);
            }}
            className="ml-2 text-cyan-400 hover:text-cyan-200 underline text-[11px]"
          >
            Azzera selezione
          </button>
        </div>
      )}

      {/* SVG Canvas Principale */}
      <svg
        className="w-full h-full"
        viewBox={`0 0 ${canvasBounds.width} ${canvasBounds.height}`}
      >
        <defs>
          {/* Marcatori frecce per gli archi orientati */}
          <marker
            id="arrow-default"
            viewBox="0 0 10 10"
            refX="9"
            refY="5"
            markerWidth="6"
            markerHeight="6"
            orient="auto-start-reverse"
          >
            <path d="M 0 1 L 10 5 L 0 9 z" fill="#64748b" />
          </marker>
          <marker
            id="arrow-active"
            viewBox="0 0 10 10"
            refX="9"
            refY="5"
            markerWidth="7"
            markerHeight="7"
            orient="auto-start-reverse"
          >
            <path d="M 0 1 L 10 5 L 0 9 z" fill="#f43f5e" />
          </marker>
          <marker
            id="arrow-pivot"
            viewBox="0 0 10 10"
            refX="9"
            refY="5"
            markerWidth="7"
            markerHeight="7"
            orient="auto-start-reverse"
          >
            <path d="M 0 1 L 10 5 L 0 9 z" fill="#f97316" />
          </marker>
          <marker
            id="arrow-network"
            viewBox="0 0 10 10"
            refX="9"
            refY="5"
            markerWidth="6"
            markerHeight="6"
            orient="auto-start-reverse"
          >
            <path d="M 0 1 L 10 5 L 0 9 z" fill="#38bdf8" />
          </marker>

          {/* Pattern a griglia di sfondo in stile CyberSec Radar */}
          <pattern id="grid" width="30" height="30" patternUnits="userSpaceOnUse">
            <path
              d="M 30 0 L 0 0 0 30"
              fill="none"
              stroke="#1e293b"
              strokeWidth="0.6"
              strokeOpacity="0.7"
            />
          </pattern>
        </defs>

        {/* Griglia di sfondo */}
        <rect width="100%" height="100%" fill="url(#grid)" />

        {/* Gruppo principale con trasformazione Pan e Zoom */}
        <g transform={`translate(${pan.x}, ${pan.y}) scale(${zoom})`}>
          {/* 1. Rendering Archi di Connessione */}
          {graph.edges?.map((edge) => {
            const sourceNode = nodeMap.get(edge.source);
            const targetNode = nodeMap.get(edge.target);
            if (!sourceNode || !targetNode) return null;

            const isEdgeInActivePath =
              activePathEdgeIds.has(edge.id) ||
              (activePathNodeIds.has(edge.source) && activePathNodeIds.has(edge.target));

            const isDimmed =
              (selectedPathId && !isEdgeInActivePath) ||
              (selectedNodeId &&
                selectedNodeId !== edge.source &&
                selectedNodeId !== edge.target);

            // Coordinate di connessione (da bordo destro a bordo sinistro)
            const x1 = sourceNode.x + sourceNode.width;
            const y1 = sourceNode.y + sourceNode.height / 2;
            const x2 = targetNode.x;
            const y2 = targetNode.y + targetNode.height / 2;

            // Curva di Bézier morbida
            const dx = Math.abs(x2 - x1);
            const cx1 = x1 + Math.max(40, dx * 0.45);
            const cx2 = x2 - Math.max(40, dx * 0.45);
            const pathD = `M ${x1} ${y1} C ${cx1} ${y1}, ${cx2} ${y2}, ${x2} ${y2}`;

            let strokeColor = '#475569';
            let marker = 'url(#arrow-default)';
            let strokeWidth = 1.6;

            if (edge.type === 'EXPLOITS') {
              strokeColor = '#f43f5e';
              marker = 'url(#arrow-active)';
              strokeWidth = 2.0;
            } else if (edge.type === 'PIVOTS_TO') {
              strokeColor = '#f97316';
              marker = 'url(#arrow-pivot)';
              strokeWidth = 2.2;
            } else if (edge.type === 'NETWORK_ACCESS') {
              strokeColor = '#38bdf8';
              marker = 'url(#arrow-network)';
              strokeWidth = 1.8;
            }

            if (isEdgeInActivePath) {
              strokeColor = '#f43f5e';
              marker = 'url(#arrow-active)';
              strokeWidth = 3.2;
            }

            return (
              <g
                key={edge.id}
                className="transition-opacity duration-300"
                opacity={isDimmed ? 0.15 : 1}
              >
                <path
                  d={pathD}
                  fill="none"
                  stroke={strokeColor}
                  strokeWidth={strokeWidth}
                  strokeDasharray={
                    edge.inferred || edge.type === 'PIVOTS_TO' || isEdgeInActivePath ? '6 4' : undefined
                  }
                  markerEnd={marker}
                  className={isEdgeInActivePath ? 'animate-pulse' : ''}
                />
                {/* Label opzionale dell'arco posizionata sul punto medio */}
                {edge.label && !isDimmed && (
                  <text
                    x={(x1 + x2) / 2}
                    y={(y1 + y2) / 2 - (edge.inferred ? 10 : 6)}
                    fill="#94a3b8"
                    fontSize="9"
                    fontFamily="monospace"
                    textAnchor="middle"
                    className="select-none pointer-events-none bg-slate-900"
                  >
                    {edge.label}
                  </text>
                )}
                {/* Badge visivo per archi inferiti con tooltip su confidenza e base euristica */}
                {edge.inferred && !isDimmed && (
                  <g className="cursor-help pointer-events-auto">
                    <title>{`Arco Inferito: ${edge.inference_basis || 'Correlazione euristica di rete / subnet'} (Confidenza: ${Math.round((edge.confidence ?? 0.8) * 100)}%)`}</title>
                    <rect
                      x={(x1 + x2) / 2 - 44}
                      y={(y1 + y2) / 2 + 2}
                      width="88"
                      height="15"
                      rx="3"
                      fill="#0f172a"
                      stroke="#f97316"
                      strokeWidth="0.8"
                      strokeDasharray="2 2"
                    />
                    <text
                      x={(x1 + x2) / 2}
                      y={(y1 + y2) / 2 + 13}
                      fill="#fb923c"
                      fontSize="8"
                      fontWeight="bold"
                      fontFamily="monospace"
                      textAnchor="middle"
                      className="select-none"
                    >
                      {`INFERRED ${Math.round((edge.confidence ?? 0.8) * 100)}%`}
                    </text>
                  </g>
                )}
              </g>
            );
          })}

          {/* 2. Rendering Nodi Topologici */}
          {positionedNodes.map((node) => {
            const isNodeSelected = selectedNodeId === node.id;
            const isNodeInPath = activePathNodeIds.has(node.id);
            const isDimmed =
              (selectedPathId && !isNodeInPath) ||
              (selectedNodeId && !isNodeSelected && !activePathNodeIds.has(node.id));

            const styles = getNodeColor(node);
            const Icon = styles.icon;

            return (
              <g
                key={node.id}
                transform={`translate(${node.x}, ${node.y})`}
                onClick={(e) => {
                  e.stopPropagation();
                  onSelectNode?.(isNodeSelected ? null : node);
                }}
                className="cursor-pointer transition-opacity duration-300 group"
                opacity={isDimmed ? 0.25 : 1}
              >
                {/* Aura di sicurezza per i nodi Chokepoint (massimo ROI di remediation) */}
                {node.isChokepoint && (
                  <rect
                    x="-6"
                    y="-6"
                    width={node.width + 12}
                    height={node.height + 12}
                    rx="14"
                    fill="none"
                    stroke="#f59e0b"
                    strokeWidth="2"
                    strokeDasharray="4 4"
                    className="animate-spin-slow opacity-80"
                  />
                )}

                {/* Rettangolo card principale del nodo */}
                <rect
                  x="0"
                  y="0"
                  width={node.width}
                  height={node.height}
                  rx="10"
                  fill="#0f172a"
                  stroke={
                    isNodeSelected
                      ? '#38bdf8'
                      : isNodeInPath
                      ? '#f43f5e'
                      : node.isChokepoint
                      ? '#f59e0b'
                      : '#334155'
                  }
                  strokeWidth={isNodeSelected || isNodeInPath ? 2.5 : 1.5}
                  className="filter drop-shadow-md group-hover:stroke-slate-400 transition-colors"
                />

                {/* Striscia laterale colorata in base alla severità/tipo */}
                <path
                  d={`M 0 10 Q 0 0 10 0 L 10 ${node.height} L 0 ${node.height} Z`}
                  fill={
                    node.type === 'ENTRYPOINT'
                      ? '#818cf8'
                      : node.severity === 'CRITICAL'
                      ? '#f43f5e'
                      : node.severity === 'HIGH'
                      ? '#f97316'
                      : node.severity === 'MEDIUM'
                      ? '#eab308'
                      : '#38bdf8'
                  }
                />

                {/* Contenuto HTML-like nel foreignObject SVG */}
                <foreignObject x="10" y="0" width={node.width - 10} height={node.height}>
                  <div className="flex items-center h-full px-2.5 gap-2 text-left">
                    <div
                      className={`p-1.5 rounded-lg border ${styles.badgeBg} flex-shrink-0`}
                    >
                      <Icon className="w-4 h-4" />
                    </div>

                    <div className="flex flex-col min-w-0 flex-1">
                      <div className="flex items-center justify-between gap-1">
                        <span
                          className="text-xs font-semibold text-slate-100 truncate"
                          title={node.label}
                        >
                          {node.label}
                        </span>
                        {node.isChokepoint && (
                          <span
                            className="text-[9px] font-mono px-1 py-0.5 rounded bg-amber-500/20 text-amber-300 border border-amber-500/40 flex-shrink-0"
                            title="Chokepoint Strategico"
                          >
                            CHOKE
                          </span>
                        )}
                      </div>

                      <div className="flex items-center gap-1.5 mt-0.5 text-[11px] font-mono text-slate-400">
                        {node.type === 'HOST' && node.ip && (
                          <span className="truncate">{node.ip}</span>
                        )}
                        {node.type === 'VULNERABILITY' && node.cvss && (
                          <span className="text-rose-400 font-bold">
                            CVSS {node.cvss.toFixed(1)}
                          </span>
                        )}
                        {node.type === 'SERVICE' && (
                          <span>
                            Port {node.port}/{node.protocol || 'tcp'}
                          </span>
                        )}
                        {node.type === 'ENTRYPOINT' && (
                          <span className="text-indigo-400">Perimeter Gateway</span>
                        )}
                      </div>
                    </div>
                  </div>
                </foreignObject>
              </g>
            );
          })}
        </g>
      </svg>

      {/* Legenda Flottante in basso a sinistra */}
      <div className="absolute bottom-4 left-4 z-20 flex flex-wrap items-center gap-3 bg-slate-900/90 backdrop-blur-md border border-slate-800 rounded-lg px-3 py-2 shadow-xl text-xs text-slate-400">
        <div className="flex items-center gap-1.5">
          <span className="w-2.5 h-2.5 rounded-full bg-indigo-500" />
          <span>Entrypoint</span>
        </div>
        <div className="flex items-center gap-1.5">
          <span className="w-2.5 h-2.5 rounded-full bg-slate-400" />
          <span>Host</span>
        </div>
        <div className="flex items-center gap-1.5">
          <span className="w-2.5 h-2.5 rounded-full bg-cyan-400" />
          <span>Servizio</span>
        </div>
        <div className="flex items-center gap-1.5">
          <span className="w-2.5 h-2.5 rounded-full bg-rose-500" />
          <span>Vulnerabilità (RCE)</span>
        </div>
        <div className="flex items-center gap-1.5">
          <span className="w-2.5 h-2.5 rounded border border-amber-400 bg-amber-400/20" />
          <span className="text-amber-300 font-medium">Chokepoint (Max ROI)</span>
        </div>
      </div>
    </div>
  );
};
