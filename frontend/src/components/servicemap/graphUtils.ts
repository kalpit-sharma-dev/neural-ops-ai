import type { DependencyMap } from '../../api/types';
import type { Node, Edge } from '@xyflow/react';
import type { ServiceNodeData } from './ServiceMapNode';

export type LayoutMode = 'layered' | 'circular' | 'force';

function hashString(s: string): number {
  let h = 0;
  for (let i = 0; i < s.length; i += 1) h = (h << 5) - h + s.charCodeAt(i);
  return Math.abs(h);
}

function buildSparkline(seed: number) {
  return Array.from({ length: 12 }, (_, i) => ({
    t: i,
    v: Math.max(0, 0.5 + Math.sin(i + seed) * 0.3 + (seed % 5) * 0.05),
  }));
}

export function computeNodeMetrics(
  svc: string,
  map: DependencyMap,
  healthScores: Record<string, number>,
  failingCounts: Record<string, number>,
) {
  const outgoing = map[svc] ?? [];
  const incoming = Object.values(map)
    .flat()
    .filter((e) => e.target === svc);
  const rps = outgoing.reduce((s, e) => s + e.callCount, 0) || incoming.reduce((s, e) => s + e.callCount, 0) || 50;
  const p99 =
    outgoing.reduce((max, e) => Math.max(max, e.p99LatencyMs), 0) ||
    incoming.reduce((max, e) => Math.max(max, e.p99LatencyMs), 0) ||
    80;
  const score = healthScores[svc] ?? 0.9;
  const errorRate = failingCounts[svc]
    ? Math.min(25, failingCounts[svc] * 0.5)
    : score < 0.85
      ? (1 - score) * 15
      : hashString(svc) % 100 / 50;
  const health: ServiceNodeData['health'] =
    score < 0.6 || errorRate > 8 ? 'down' : score < 0.85 || errorRate > 3 ? 'degraded' : 'healthy';

  return {
    errorRate,
    p99,
    rps,
    health,
    sparkline: buildSparkline(hashString(svc) % 10),
  };
}

function layeredLayout(services: string[], map: DependencyMap): Map<string, { x: number; y: number }> {
  const positions = new Map<string, { x: number; y: number }>();
  const inDegree = new Map<string, number>();
  services.forEach((s) => inDegree.set(s, 0));
  Object.entries(map).forEach(([src, edges]) => {
    if (!inDegree.has(src)) inDegree.set(src, 0);
    edges.forEach((e) => inDegree.set(e.target, (inDegree.get(e.target) ?? 0) + 1));
  });

  const layers: string[][] = [];
  const remaining = new Set(services);
  while (remaining.size > 0) {
    const layer = [...remaining].filter((s) => (inDegree.get(s) ?? 0) === 0);
    if (layer.length === 0) {
      layers.push([...remaining]);
      break;
    }
    layers.push(layer);
    layer.forEach((s) => {
      remaining.delete(s);
      (map[s] ?? []).forEach((e) => inDegree.set(e.target, Math.max(0, (inDegree.get(e.target) ?? 1) - 1)));
    });
  }

  layers.forEach((layer, li) => {
    layer.forEach((svc, si) => {
      positions.set(svc, { x: li * 260, y: si * 150 });
    });
  });
  return positions;
}

function circularLayout(services: string[]): Map<string, { x: number; y: number }> {
  const positions = new Map<string, { x: number; y: number }>();
  const radius = Math.max(200, services.length * 35);
  const cx = 400;
  const cy = 300;
  services.forEach((svc, i) => {
    const angle = (2 * Math.PI * i) / services.length - Math.PI / 2;
    positions.set(svc, { x: cx + radius * Math.cos(angle), y: cy + radius * Math.sin(angle) });
  });
  return positions;
}

function forceLayout(services: string[]): Map<string, { x: number; y: number }> {
  const positions = new Map<string, { x: number; y: number }>();
  const cols = Math.ceil(Math.sqrt(services.length));
  services.forEach((svc, i) => {
    positions.set(svc, {
      x: (i % cols) * 220 + (hashString(svc) % 40),
      y: Math.floor(i / cols) * 160 + (hashString(svc) % 30),
    });
  });
  return positions;
}

export function buildServiceGraph(
  map: DependencyMap,
  layout: LayoutMode,
  healthScores: Record<string, number>,
  failingCounts: Record<string, number>,
  overlay?: { blastRadius: string[]; firstFailing?: string; criticalPath: string[] },
): { nodes: Node<ServiceNodeData>[]; edges: Edge[] } {
  const services = new Set<string>();
  Object.entries(map).forEach(([source, edges]) => {
    services.add(source);
    edges.forEach((e) => services.add(e.target));
  });
  const serviceList = Array.from(services);

  const layoutFn =
    layout === 'circular'
      ? () => circularLayout(serviceList)
      : layout === 'force'
        ? () => forceLayout(serviceList)
        : () => layeredLayout(serviceList, map);
  const positions = layoutFn();

  const nodes: Node<ServiceNodeData>[] = serviceList.map((svc) => {
    const metrics = computeNodeMetrics(svc, map, healthScores, failingCounts);
    const pos = positions.get(svc) ?? { x: 0, y: 0 };
    return {
      id: svc,
      type: 'service',
      position: pos,
      data: {
        label: svc,
        ...metrics,
        blastRadius: overlay?.blastRadius.includes(svc),
        firstFailing: overlay?.firstFailing === svc,
        criticalPath: overlay?.criticalPath.includes(svc),
      },
    };
  });

  const edges: Edge[] = [];
  Object.entries(map).forEach(([source, list]) => {
    list.forEach((edge) => {
      const stroke =
        edge.p99LatencyMs > 300
          ? 'var(--error)'
          : edge.p99LatencyMs > 150
            ? 'var(--warning)'
            : 'var(--status-healthy)';
      const width = Math.min(6, 1 + edge.callCount / 200);
      edges.push({
        id: `${source}-${edge.target}`,
        source,
        target: edge.target,
        animated: true,
        style: { stroke, strokeWidth: width },
        label: `${edge.callCount}/s · ${edge.p99LatencyMs}ms`,
        data: {
          callCount: edge.callCount,
          p99LatencyMs: edge.p99LatencyMs,
          source,
          target: edge.target,
        },
      });
    });
  });

  return { nodes, edges };
}

export function findCriticalPath(
  map: DependencyMap,
  entryServices: string[],
  targetService: string,
): string[] {
  const queue: { svc: string; path: string[] }[] = entryServices.map((s) => ({ svc: s, path: [s] }));
  const visited = new Set<string>();

  while (queue.length > 0) {
    const { svc, path } = queue.shift()!;
    if (svc === targetService) return path;
    if (visited.has(svc)) continue;
    visited.add(svc);
    (map[svc] ?? []).forEach((e) => {
      queue.push({ svc: e.target, path: [...path, e.target] });
    });
  }
  return entryServices.length ? [entryServices[0], targetService] : [targetService];
}

export function findEntryServices(map: DependencyMap): string[] {
  const allTargets = new Set(Object.values(map).flat().map((e) => e.target));
  const sources = Object.keys(map);
  const entries = sources.filter((s) => !allTargets.has(s));
  return entries.length ? entries : sources.slice(0, 1);
}
