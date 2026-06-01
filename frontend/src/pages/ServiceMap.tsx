import { useCallback, useMemo, useRef, useState, type MouseEvent } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  ReactFlowProvider,
  useReactFlow,
  type Node,
  type Edge,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { Download, Maximize2, Minus, Plus, Route } from 'lucide-react';
import toast from 'react-hot-toast';
import { fetchDependencyMap } from '../api/search';
import { fetchTopology, fetchZones } from '../api/observability';
import { topologyToDependencyMap } from '../utils/topologyMap';
import { fetchDashboardOverview } from '../api/dashboard';
import { getApiErrorMessage } from '../api/client';
import {
  buildServiceGraph,
  findCriticalPath,
  findEntryServices,
  type LayoutMode,
} from '../components/servicemap/graphUtils';
import { ServiceDetailPanel } from '../components/servicemap/ServiceDetailPanel';
import { serviceMapNodeTypes, type ServiceNodeData } from '../components/servicemap/ServiceMapNode';
import { Button } from '../components/ui/Button';
import { ErrorState, LoadingState, PageHeader } from '../components/ui/PageStates';
import { useFilterStore } from '../store/filterStore';
import { exportElementAsPng } from '../utils/exportPng';

function ServiceMapCanvas() {
  const containerRef = useRef<HTMLDivElement>(null);
  const { fitView, zoomIn, zoomOut } = useReactFlow();
  const { timeRange, setTimeRange } = useFilterStore();

  const [selected, setSelected] = useState<string | null>(null);
  const [errorsOnly, setErrorsOnly] = useState(false);
  const [layout, setLayout] = useState<LayoutMode>('layered');
  const [highlightCritical, setHighlightCritical] = useState(false);
  const [selectedIncidentId, setSelectedIncidentId] = useState<string>('');
  const [zoneFilter, setZoneFilter] = useState<string>('');

  const zonesQuery = useQuery({ queryKey: ['zones'], queryFn: fetchZones });

  const { data: topologyData } = useQuery({
    queryKey: ['topology', zoneFilter],
    queryFn: () => fetchTopology(zoneFilter || undefined),
    refetchInterval: 60_000,
  });

  const { data: mapData, isLoading, error, refetch } = useQuery({
    queryKey: ['dependency-map', topologyData?.at],
    queryFn: async () => {
      if (topologyData && topologyData.nodes.length > 0) {
        return topologyToDependencyMap(topologyData);
      }
      return fetchDependencyMap();
    },
    refetchInterval: 60_000,
  });

  const { data: dashboard } = useQuery({
    queryKey: ['dashboard-overview'],
    queryFn: fetchDashboardOverview,
    refetchInterval: 60_000,
  });

  const healthScores = useMemo(() => {
    const scores: Record<string, number> = {};
    (dashboard?.serviceHealth ?? []).forEach((s) => {
      scores[s.service] = s.score;
    });
    return scores;
  }, [dashboard]);

  const failingCounts = useMemo(() => {
    const counts: Record<string, number> = {};
    (dashboard?.topFailingServices ?? []).forEach((s) => {
      counts[s.service] = s.errorCount;
    });
    return counts;
  }, [dashboard]);

  const activeIncidents = dashboard?.activeIncidents ?? [];
  const overlayIncident = activeIncidents.find((i) => i.id === selectedIncidentId) ?? activeIncidents[0];

  const overlay = useMemo(() => {
    if (!overlayIncident || !mapData) return undefined;
    const firstFailing =
      dashboard?.topFailingServices?.[0]?.service ?? overlayIncident.service ?? '';
    const blastRadius = Object.keys(mapData).slice(0, 4);
    const entries = findEntryServices(mapData);
    const criticalPath = highlightCritical
      ? findCriticalPath(mapData, entries, firstFailing || entries[0] || '')
      : [];
    return { blastRadius, firstFailing, criticalPath };
  }, [overlayIncident, mapData, dashboard, highlightCritical]);

  const graph = useMemo(
    () =>
      mapData
        ? buildServiceGraph(mapData, layout, healthScores, failingCounts, overlay)
        : { nodes: [] as Node<ServiceNodeData>[], edges: [] as Edge[] },
    [mapData, layout, healthScores, failingCounts, overlay],
  );

  const filteredNodes = useMemo(() => {
    if (!errorsOnly) return graph.nodes;
    return graph.nodes.filter((n) => n.data.health !== 'healthy');
  }, [graph.nodes, errorsOnly]);

  const onNodeClick = useCallback((_: MouseEvent, node: Node) => {
    setSelected(node.id);
  }, []);

  const exportPng = async () => {
    if (!containerRef.current) return;
    try {
      await exportElementAsPng(containerRef.current, `service-map-${Date.now()}.png`);
      toast.success('Map exported');
    } catch {
      toast.error('Export failed');
    }
  };

  const deps = selected && mapData ? mapData[selected] ?? [] : [];

  return (
    <>
      <PageHeader
        title="Service Map"
        subtitle="Live microservice topology and dependency health"
        actions={
          <select
            value={zoneFilter}
            onChange={(e) => setZoneFilter(e.target.value)}
            className="ui-select"
            aria-label="Management zone"
          >
            <option value="">All zones</option>
            {(zonesQuery.data ?? []).map((z) => (
              <option key={z.id} value={z.id}>{z.name}</option>
            ))}
          </select>
        }
      />

      {isLoading && <LoadingState />}
      {error && <ErrorState message={getApiErrorMessage(error)} onRetry={() => refetch()} />}

      {mapData && (
        <div className={`service-map-layout ${selected ? 'service-map-layout--detail' : ''}`}>
          <div className="service-map-container" ref={containerRef}>
            <div className="service-map-controls service-map-controls--left">
              <Button variant="ghost" size="sm" onClick={() => zoomIn()}>
                <Plus size={14} />
              </Button>
              <Button variant="ghost" size="sm" onClick={() => zoomOut()}>
                <Minus size={14} />
              </Button>
              <Button variant="ghost" size="sm" onClick={() => fitView({ padding: 0.2 })}>
                <Maximize2 size={14} />
              </Button>
              <Button variant="ghost" size="sm" onClick={() => void exportPng()}>
                <Download size={14} /> PNG
              </Button>
              <div className="service-map-layout-toggle">
                {(['layered', 'force', 'circular'] as LayoutMode[]).map((mode) => (
                  <button
                    key={mode}
                    type="button"
                    className={`pill ${layout === mode ? 'pill--active' : ''}`}
                    onClick={() => setLayout(mode)}
                  >
                    {mode}
                  </button>
                ))}
              </div>
              <Button
                variant={errorsOnly ? 'primary' : 'secondary'}
                size="sm"
                onClick={() => setErrorsOnly((v) => !v)}
              >
                Errors only
              </Button>
            </div>

            <div className="service-map-controls service-map-controls--right">
              <div className="pill-group">
                {(['1h', '6h', '24h'] as const).map((tr) => (
                  <button
                    key={tr}
                    type="button"
                    className={`pill ${timeRange === tr ? 'pill--active' : ''}`}
                    onClick={() => setTimeRange(tr)}
                  >
                    {tr}
                  </button>
                ))}
              </div>
              <Button
                variant={highlightCritical ? 'primary' : 'secondary'}
                size="sm"
                onClick={() => setHighlightCritical((v) => !v)}
              >
                <Route size={14} /> Critical path
              </Button>
              {activeIncidents.length > 0 && (
                <select
                  className="filter-input"
                  value={selectedIncidentId || activeIncidents[0]?.id}
                  onChange={(e) => setSelectedIncidentId(e.target.value)}
                >
                  {activeIncidents.map((inc) => (
                    <option key={inc.id} value={inc.id}>
                      {inc.title.slice(0, 40)}
                    </option>
                  ))}
                </select>
              )}
            </div>

            <ReactFlow
              nodes={filteredNodes}
              edges={graph.edges}
              nodeTypes={serviceMapNodeTypes}
              onNodeClick={onNodeClick}
              fitView
              proOptions={{ hideAttribution: true }}
            >
              <Background color="var(--border-subtle)" gap={20} />
              <Controls showInteractive={false} />
              <MiniMap nodeColor={(n) => (n.data as ServiceNodeData).health === 'down' ? 'var(--error)' : 'var(--accent-primary)'} />
            </ReactFlow>
          </div>

          {selected && (
            <ServiceDetailPanel
              service={selected}
              nodeData={graph.nodes.find((n) => n.id === selected)?.data}
              incidents={activeIncidents}
              dependencies={deps}
              onClose={() => setSelected(null)}
            />
          )}
        </div>
      )}
    </>
  );
}

export default function ServiceMap() {
  return (
    <ReactFlowProvider>
      <ServiceMapCanvas />
    </ReactFlowProvider>
  );
}
