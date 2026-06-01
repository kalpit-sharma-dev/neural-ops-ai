import { Link } from '@tanstack/react-router';
import type { ReactNode } from 'react';
import { Badge } from '../ui/Badge';

export type EntityType = 'service' | 'host' | 'pod' | 'database' | 'container';

interface EntityShellProps {
  entityType: EntityType;
  entityId: string;
  displayName?: string;
  health?: string;
  children: ReactNode;
}

function buildTabs(entityType: EntityType, entityId: string) {
  if (entityType === 'service') {
    return [
      { label: 'Overview', to: '/entities/$type/$id', params: { type: entityType, id: entityId } },
      { label: 'Traces', to: '/traces', search: { service: entityId } },
      { label: 'Logs', to: '/logs', search: { service: entityId } },
      { label: 'Metrics', to: '/metrics', search: { service: entityId } },
      { label: 'Service map', to: '/service-map' },
    ];
  }
  return [
    { label: 'Overview', to: '/entities/$type/$id', params: { type: entityType, id: entityId } },
  ];
}

export function EntityShell({ entityType, entityId, displayName, health, children }: EntityShellProps) {
  const tabs = buildTabs(entityType, entityId);
  const name = displayName ?? entityId;
  const healthVariant = health === 'critical' || health === 'down' ? 'critical' : health === 'degraded' ? 'warning' : 'healthy';

  return (
    <div className="entity-shell">
      <header className="entity-shell__header">
        <div>
          <p className="muted" style={{ margin: 0, textTransform: 'capitalize' }}>{entityType}</p>
          <h1 style={{ margin: '4px 0' }}>{name}</h1>
          {health && <Badge variant={healthVariant}>{health}</Badge>}
        </div>
      </header>
      <nav className="tab-bar entity-shell__tabs">
        {tabs.map((tab) => (
          <Link
            key={tab.label}
            to={tab.to}
            params={tab.params}
            search={'search' in tab ? tab.search : undefined}
            className="tab-bar__item"
          >
            {tab.label}
          </Link>
        ))}
      </nav>
      <div className="entity-shell__body">{children}</div>
    </div>
  );
}
