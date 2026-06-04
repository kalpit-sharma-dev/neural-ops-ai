import { Link, useRouterState } from '@tanstack/react-router';
import type { ReactNode } from 'react';
import { Badge } from '../ui/Badge';

export type EntityType = 'service' | 'host' | 'pod' | 'database' | 'container';
export type EntityTab = 'overview' | 'metrics' | 'logs' | 'traces';

interface EntityShellProps {
  entityType: EntityType;
  entityId: string;
  displayName?: string;
  health?: string;
  activeTab?: EntityTab;
  onTabChange?: (tab: EntityTab) => void;
  headerActions?: ReactNode;
  children: ReactNode;
}

const TABS: { id: EntityTab; label: string }[] = [
  { id: 'overview', label: 'Overview' },
  { id: 'metrics', label: 'Metrics' },
  { id: 'logs', label: 'Logs' },
  { id: 'traces', label: 'Traces' },
];

export function EntityShell({
  entityType,
  entityId,
  displayName,
  health,
  activeTab = 'overview',
  onTabChange,
  headerActions,
  children,
}: EntityShellProps) {
  const router = useRouterState();
  const name = displayName ?? entityId;
  const healthVariant =
    health === 'critical' || health === 'down' ? 'critical' : health === 'degraded' ? 'warning' : 'healthy';

  return (
    <div className="entity-shell">
      <header className="entity-shell__header">
        <div>
          <p className="muted" style={{ margin: 0, textTransform: 'capitalize' }}>{entityType}</p>
          <h1 style={{ margin: '4px 0' }}>{name}</h1>
          {health && <Badge variant={healthVariant}>{health}</Badge>}
        </div>
        {headerActions && <div className="entity-shell__actions">{headerActions}</div>}
      </header>
      <nav className="tab-bar entity-shell__tabs" aria-label="Entity sections">
        {TABS.map((tab) =>
          onTabChange ? (
            <button
              key={tab.id}
              type="button"
              className={`tab-bar__item ${activeTab === tab.id ? 'tab-bar__item--active' : ''}`}
              onClick={() => onTabChange(tab.id)}
            >
              {tab.label}
            </button>
          ) : (
            <Link
              key={tab.id}
              to="/entities/$type/$id"
              params={{ type: entityType, id: entityId }}
              search={{ tab: tab.id }}
              className={`tab-bar__item ${(router.location.search as { tab?: string }).tab === tab.id ? 'tab-bar__item--active' : ''}`}
            >
              {tab.label}
            </Link>
          ),
        )}
      </nav>
      <div className="entity-shell__body">{children}</div>
    </div>
  );
}
