import { format } from 'date-fns';
import { Badge } from '../ui/Badge';
import { StatusDot } from '../ui/StatusDot';
import {
  useFilterStore,
  type Environment,
  type TimeRangePreset,
} from '../../store/filterStore';

const SEVERITIES = ['DEBUG', 'INFO', 'WARN', 'ERROR', 'FATAL'] as const;
const ENVIRONMENTS: Environment[] = ['ALL', 'PROD', 'STAGING', 'DEV'];
const TIME_PRESETS: { value: TimeRangePreset; label: string }[] = [
  { value: '1h', label: '1h' },
  { value: '6h', label: '6h' },
  { value: '24h', label: '24h' },
  { value: '7d', label: '7d' },
];

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
  } = useFilterStore();

  const options =
    serviceOptions.length > 0
      ? serviceOptions
      : ['payment-api', 'upi-service', 'auth-service', 'ledger-service', 'notification-service'];

  return (
    <aside className="log-filters">
      <div className="filter-group">
        <h4>Services</h4>
        {options.map((svc) => {
          const score = serviceHealth[svc] ?? 0.9;
          return (
            <label key={svc} className="checkbox-row">
              <input type="checkbox" checked={services.includes(svc)} onChange={() => toggleService(svc)} />
              <StatusDot status={healthFromScore(score)} />
              <span>{svc}</span>
            </label>
          );
        })}
      </div>

      <div className="filter-group">
        <h4>Severity</h4>
        {SEVERITIES.map((sev) => (
          <label key={sev} className="checkbox-row">
            <input type="checkbox" checked={severities.includes(sev)} onChange={() => toggleSeverity(sev)} />
            <Badge variant={severityVariant(sev) as 'info'}>{sev}</Badge>
            <span className="filter-count">{severityCounts[sev] ?? 0}</span>
          </label>
        ))}
      </div>

      <div className="filter-group">
        <h4>Environment</h4>
        <div className="pill-group">
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
      </div>

      <div className="filter-group">
        <h4>Time range</h4>
        <div className="pill-group">
          {TIME_PRESETS.map(({ value, label }) => (
            <button
              key={value}
              type="button"
              className={`pill ${timeRange === value ? 'pill--active' : ''}`}
              onClick={() => setTimeRange(value)}
            >
              {label}
            </button>
          ))}
        </div>
      </div>

      <div className="filter-group">
        <h4>Host / Pod</h4>
        <input
          className="filter-input"
          placeholder="Host filter"
          value={host}
          onChange={(e) => setHost(e.target.value)}
        />
        <input
          className="filter-input"
          placeholder="Pod filter"
          value={pod}
          onChange={(e) => setPod(e.target.value)}
          style={{ marginTop: 8 }}
        />
      </div>

      <div className="filter-group">
        <h4>Advanced</h4>
        <label className="checkbox-row">
          <input type="checkbox" checked={hasStackTrace} onChange={(e) => setHasStackTrace(e.target.checked)} />
          Has stack trace
        </label>
        <label className="checkbox-row">
          <input
            type="checkbox"
            checked={hasAIExplanation}
            onChange={(e) => setHasAIExplanation(e.target.checked)}
          />
          Has AI explanation
        </label>
      </div>

      <div className="filter-group">
        <h4>Quick filters</h4>
        <div className="pill-group">
          <button type="button" className="pill pill--active" onClick={setErrorsOnly}>
            Errors only
          </button>
          <button
            type="button"
            className={`pill ${myServicesOnly ? 'pill--active' : ''}`}
            onClick={() => setMyServicesOnly(!myServicesOnly)}
          >
            My services
          </button>
        </div>
      </div>
    </aside>
  );
}

export function formatLogTimestamp(ts: string) {
  return format(new Date(ts), 'yyyy-MM-dd HH:mm:ss.SSS');
}
