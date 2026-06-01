import { useMemo, useState } from 'react';
import { format, formatDistanceToNow } from 'date-fns';
import { AlertTriangle, GitBranch, LineChart, Rocket } from 'lucide-react';
import type { IncidentEvent } from '../../api/types';
import { Badge } from '../ui/Badge';
import { Button } from '../ui/Button';

const EVENT_TYPES = ['deployment', 'error', 'metric', 'alert'] as const;

function eventIcon(type: string) {
  const t = type.toLowerCase();
  if (t.includes('deploy')) return Rocket;
  if (t.includes('metric')) return LineChart;
  if (t.includes('alert')) return AlertTriangle;
  return GitBranch;
}

function eventColor(type: string): string {
  const t = type.toLowerCase();
  if (t.includes('deploy')) return 'timeline-event--deployment';
  if (t.includes('metric')) return 'timeline-event--metric';
  if (t.includes('alert')) return 'timeline-event--alert';
  return 'timeline-event--error';
}

interface TimelineViewProps {
  events: IncidentEvent[];
  narrative?: string;
  onMarkRootCause?: (eventId: string) => void;
}

export function TimelineView({ events, narrative, onMarkRootCause }: TimelineViewProps) {
  const [filters, setFilters] = useState<Record<string, boolean>>({
    deployment: true,
    error: true,
    metric: true,
    alert: true,
  });

  const filtered = useMemo(() => {
    return events.filter((ev) => {
      const t = ev.eventType.toLowerCase();
      if (t.includes('deploy')) return filters.deployment;
      if (t.includes('metric')) return filters.metric;
      if (t.includes('alert')) return filters.alert;
      return filters.error;
    });
  }, [events, filters]);

  return (
    <div className="timeline-view">
      <div className="timeline-filters">
        {EVENT_TYPES.map((type) => (
          <label key={type} className="checkbox-row">
            <input
              type="checkbox"
              checked={filters[type]}
              onChange={(e) => setFilters((f) => ({ ...f, [type]: e.target.checked }))}
            />
            {type}
          </label>
        ))}
      </div>

      <div className="timeline-list">
        {filtered.map((ev, idx) => {
          const Icon = eventIcon(ev.eventType);
          return (
            <div key={ev.id ?? idx} className={`timeline-event ${eventColor(ev.eventType)}`}>
              <div className="timeline-event__icon">
                <Icon size={16} />
              </div>
              <div className="timeline-event__body">
                <div className="timeline-event__header">
                  <Badge variant="info">{ev.eventType}</Badge>
                  <time title={format(new Date(ev.timestamp), 'PPpp')}>
                    {formatDistanceToNow(new Date(ev.timestamp), { addSuffix: true })}
                  </time>
                </div>
                <p>{ev.description}</p>
                {ev.service && <span className="muted">{ev.service}</span>}
                {onMarkRootCause && ev.id && (
                  <Button variant="ghost" size="sm" onClick={() => onMarkRootCause(ev.id!)}>
                    Mark as root cause
                  </Button>
                )}
              </div>
            </div>
          );
        })}
      </div>

      {narrative && <p className="muted timeline-narrative">{narrative}</p>}
    </div>
  );
}
