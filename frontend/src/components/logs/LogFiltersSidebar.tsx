import { format } from 'date-fns';
import { Badge } from '../ui/Badge';
import { StatusDot } from '../ui/StatusDot';
import {
  useFilterStore,
  type Environment,
} from '../../store/filterStore';
import { LOG_SIDEBAR_TIME_PRESETS, presetShortLabel } from '../../lib/timeRangePresets';

const SEVERITIES = ['DEBUG', 'INFO', 'WARN', 'ERROR', 'FATAL'] as const;
const ENVIRONMENTS: Environment[] = ['ALL', 'PROD', 'STAGING', 'DEV'];

export function formatLogTimestamp(iso: string): string {
  try {
    return format(new Date(iso), 'PPpp');
  } catch {
    return iso;
  }
}

function severityVariant(sev: string) {
  if (sev === 'FATAL' || sev === 'CRITICAL') return 'critical';
  if (sev === 'ERROR') return 'error';
  if (sev === 'WARN') return 'warning';
  return 'info';
}

function healthFromScore(score: number): 'healthy' | 'degraded' | 'down' {
  if (score >= 0.85) return 'healthy';
  if (score >= 0.6) return 'degraded';
  return 'down';
}

interface LogFiltersSidebarProps {
  serviceOptions: string[];
  severityCounts: Record<string, number>;
  serviceHealth: Record<string, number>;
}

export function LogFiltersSidebar({
  serviceOptions,
  severityCounts,
  serviceHealth,
}: LogFiltersSidebarProps) {
  const {
    environment,
    timeRange,
    services,
    severities,
    host,
    pod,
    hasStackTrace,
    hasAIExplanation,
    myServicesOnly,
    setEnvironment,
    setTimeRange,
    setHost,
    setPod,
    toggleService,
    toggleSeverity,
    setHasStackTrace,
    setHasAIExplanation,
    setMyServicesOnly,
    setErrorsOnly,
    clearServices,
  } = useFilterStore();

  return (
    <aside className="log-filters">
      <section className="log-filters__section">
        <h3 className="log-filters__title">Time range</h3>
        <p className="muted log-filters__hint">Quick picks — use top bar for full list</p>
        <div className="pill-group log-filters__pills">
          {LOG_SIDEBAR_TIME_PRESETS.map((value) => (
            <button
              key={value}
              type="button"
              className={`pill ${timeRange === value ? 'pill--active' : ''}`}
              onClick={() => setTimeRange(value)}
            >
              {presetShortLabel(value)}
            </button>
          ))}
        </div>
      </section>

      <section className="log-filters__section">
        <h3 className="log-filters__title">Environment</h3>
        <div className="pill-group log-filters__pills">
          {ENVIRONMENTS.map((env) => (
            <button
              key={env}
              type="button"
              className={`pill ${environment === env ? 'pill--active' : ''}`}
              onClick={() => setEnvironment(env)}
            >
              {env}
            </button>
          ))}
        </div>
      </section>

      <section className="log-filters__section">
        <div className="log-filters__row">
          <h3 className="log-filters__title">Services</h3>
          {services.length > 0 && (
            <button type="button" className="log-filters__clear" onClick={clearServices}>
              Clear
            </button>
          )}
        </div>
        <label className="checkbox-row">
          <input type="checkbox" checked={myServicesOnly} onChange={(e) => setMyServicesOnly(e.target.checked)} />
          My services
        </label>
        <ul className="log-filters__list">
          {serviceOptions.map((svc) => (
            <li key={svc}>
              <label className="checkbox-row">
                <input type="checkbox" checked={services.includes(svc)} onChange={() => toggleService(svc)} />
                <StatusDot status={healthFromScore(serviceHealth[svc] ?? 0.9)} />
                <span>{svc}</span>
              </label>
            </li>
          ))}
        </ul>
      </section>

      <section className="log-filters__section">
        <h3 className="log-filters__title">Severity</h3>
        <button type="button" className="log-filters__link" onClick={setErrorsOnly}>
          Errors only
        </button>
        <ul className="log-filters__list">
          {SEVERITIES.map((sev) => (
            <li key={sev}>
              <label className="checkbox-row">
                <input type="checkbox" checked={severities.includes(sev)} onChange={() => toggleSeverity(sev)} />
                <Badge variant={severityVariant(sev) as 'error'}>{sev}</Badge>
                <span className="muted">({severityCounts[sev] ?? 0})</span>
              </label>
            </li>
          ))}
        </ul>
      </section>

      <section className="log-filters__section">
        <h3 className="log-filters__title">Host / Pod</h3>
        <input
          className="ui-input"
          placeholder="Host contains…"
          value={host}
          onChange={(e) => setHost(e.target.value)}
        />
        <input
          className="ui-input"
          placeholder="Pod contains…"
          value={pod}
          onChange={(e) => setPod(e.target.value)}
          style={{ marginTop: 8 }}
        />
      </section>

      <section className="log-filters__section">
        <h3 className="log-filters__title">Advanced</h3>
        <label className="checkbox-row">
          <input type="checkbox" checked={hasStackTrace} onChange={(e) => setHasStackTrace(e.target.checked)} />
          Has stack trace
        </label>
        <label className="checkbox-row">
          <input type="checkbox" checked={hasAIExplanation} onChange={(e) => setHasAIExplanation(e.target.checked)} />
          Has AI explanation
        </label>
      </section>

      <p className="muted log-filters__footer">Use toolbar for 2d–90d ranges</p>
    </aside>
  );
}
