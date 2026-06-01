import { Link } from '@tanstack/react-router';
import type { DependencyEdge, IncidentSummary } from '../../api/types';
import { logsSearch } from '../../utils/logsSearch';
import { Badge } from '../ui/Badge';
import { Button } from '../ui/Button';
import { StatusDot } from '../ui/StatusDot';
import type { ServiceNodeData } from './ServiceMapNode';

interface ServiceDetailPanelProps {
  service: string;
  nodeData?: ServiceNodeData;
  incidents: IncidentSummary[];
  dependencies: DependencyEdge[];
  onClose: () => void;
}

export function ServiceDetailPanel({
  service,
  nodeData,
  incidents,
  dependencies,
  onClose,
}: ServiceDetailPanelProps) {
  const related = incidents.filter(
    (i) => i.service === service || i.title.toLowerCase().includes(service.toLowerCase()),
  );

  return (
    <aside className="service-detail-panel">
      <div className="service-detail-panel__header">
        <h3>{service}</h3>
        <Link to="/entities/$type/$id" params={{ type: 'service', id: service }} className="muted" style={{ fontSize: 12 }}>
          Entity →
        </Link>
        <Button variant="ghost" size="sm" onClick={onClose}>
          Close
        </Button>
      </div>

      {nodeData && (
        <div className="service-detail-panel__stats">
          <StatusDot status={nodeData.health} pulse={nodeData.health !== 'healthy'} />
          <div>
            <div>Error rate: {nodeData.errorRate.toFixed(1)}%</div>
            <div>p99: {Math.round(nodeData.p99)}ms</div>
            <div>Throughput: {Math.round(nodeData.rps)} req/s</div>
          </div>
        </div>
      )}

      <section>
        <h4>Recent incidents</h4>
        {related.length === 0 && <p className="muted">No active incidents</p>}
        {related.map((inc) => (
          <Link key={inc.id} to="/incidents/$id" params={{ id: inc.id }} className="service-detail-incident">
            <Badge variant="info">{inc.severity}</Badge>
            {inc.title}
          </Link>
        ))}
      </section>

      <section>
        <h4>Dependencies</h4>
        <ul className="insight-list">
          {dependencies.map((dep) => (
            <li key={dep.target}>
              → {dep.target} ({dep.callCount}/s, p99 {dep.p99LatencyMs}ms)
            </li>
          ))}
          {dependencies.length === 0 && <li className="muted">No outgoing dependencies</li>}
        </ul>
      </section>

      <Link to="/logs" search={logsSearch({ service })}>
        <Button variant="secondary" size="sm">
          View logs
        </Button>
      </Link>
    </aside>
  );
}
