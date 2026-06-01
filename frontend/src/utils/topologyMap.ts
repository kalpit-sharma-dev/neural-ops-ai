import type { DependencyMap } from '../api/types';
import type { TopologyGraph } from '../api/observability';

/** Converts live topology API payload into the dependency-map shape used by ServiceMap. */
export function topologyToDependencyMap(topology: TopologyGraph): DependencyMap {
  const map: DependencyMap = {};
  for (const edge of topology.edges) {
    const list = map[edge.source] ?? [];
    list.push({
      target: edge.target,
      callCount: edge.callCount,
      p99LatencyMs: edge.p95Ms,
    });
    map[edge.source] = list;
  }
  for (const node of topology.nodes) {
    if (!map[node.id]) {
      map[node.id] = [];
    }
  }
  return map;
}
