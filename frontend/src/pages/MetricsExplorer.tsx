import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { Bar, BarChart, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';
import { fetchMetricCatalog, queryMetric, queryPromQL } from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { PromQLEditor } from '../components/metrics/PromQLEditor';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { ErrorState, LoadingState, PageHeader } from '../components/ui/PageStates';
import { Select } from '../components/ui/Select';
import { chartCartesianDefaults } from '../lib/chartTheme';

type ChartType = 'line' | 'bar' | 'stat' | 'table';

export default function MetricsExplorer() {
  const [metric, setMetric] = useState('latency_p95');
  const [service, setService] = useState('payment-service');
  const [chartType, setChartType] = useState<ChartType>('line');
  const [promql, setPromql] = useState('rate(http_requests_total[5m])');
  const [promqlError, setPromqlError] = useState('');
  const [usePromql, setUsePromql] = useState(false);

  const catalogQuery = useQuery({ queryKey: ['metric-catalog'], queryFn: fetchMetricCatalog });
  const seriesQuery = useQuery({
    queryKey: ['metric-query', metric, service],
    queryFn: () => queryMetric(metric, service),
    enabled: !!metric && !usePromql,
  });

  const promqlQuery = useQuery({
    queryKey: ['metric-promql', promql],
    queryFn: () => queryPromQL(promql),
    enabled: usePromql && promql.length > 2,
  });

  const activeSeries = usePromql ? promqlQuery.data : seriesQuery.data;
  const activeLoading = usePromql ? promqlQuery.isLoading : seriesQuery.isLoading;
  const activeError = usePromql ? promqlQuery.error : seriesQuery.error;
  const refetch = usePromql ? promqlQuery.refetch : seriesQuery.refetch;

  const chartData = (activeSeries?.points ?? []).map((p) => ({
    time: new Date(p.timestamp).toLocaleTimeString(),
    value: p.value,
  }));

  const lastValue = chartData[chartData.length - 1]?.value;
  const chartDefaults = chartCartesianDefaults();

  const runPromql = () => {
    if (!promql.trim()) {
      setPromqlError('Query cannot be empty');
      return;
    }
    setPromqlError('');
    setUsePromql(true);
    void promqlQuery.refetch();
  };

  return (
    <div>
      <PageHeader title="Metrics Explorer" subtitle="Browse catalog metrics or run PromQL" />
      <div style={{ display: 'flex', gap: 12, marginBottom: 16, flexWrap: 'wrap', alignItems: 'center' }}>
        <div className="pill-group">
          {(['line', 'bar', 'stat', 'table'] as ChartType[]).map((t) => (
            <button
              key={t}
              type="button"
              className={`pill ${chartType === t ? 'pill--active' : ''}`}
              onClick={() => setChartType(t)}
            >
              {t}
            </button>
          ))}
        </div>
        <Button variant={usePromql ? 'ghost' : 'secondary'} size="sm" onClick={() => setUsePromql(false)}>
          Catalog
        </Button>
        <Button variant={usePromql ? 'secondary' : 'ghost'} size="sm" onClick={() => setUsePromql(true)}>
          PromQL
        </Button>
      </div>

      {!usePromql && (
        <div style={{ display: 'flex', gap: 12, marginBottom: 24 }}>
          <Select value={metric} onChange={(e) => setMetric(e.target.value)} aria-label="Metric">
            {(catalogQuery.data ?? [{ name: 'latency_p95' }]).map((m) => (
              <option key={m.name} value={m.name}>{m.name}</option>
            ))}
          </Select>
          <Select value={service} onChange={(e) => setService(e.target.value)} aria-label="Service">
            <option value="payment-service">payment-service</option>
            <option value="api-gateway">api-gateway</option>
            <option value="auth-service">auth-service</option>
          </Select>
        </div>
      )}

      {usePromql && (
        <div style={{ marginBottom: 24 }}>
          <PromQLEditor value={promql} onChange={setPromql} onRun={runPromql} error={promqlError} />
          <Button variant="primary" style={{ marginTop: 8 }} onClick={runPromql}>Run query</Button>
        </div>
      )}

      {activeLoading && <LoadingState />}
      {activeError && <ErrorState message={getApiErrorMessage(activeError)} onRetry={() => refetch()} />}
      {activeSeries && (
        <Card title={usePromql ? promql : `${metric} — ${service}`}>
          {chartType === 'stat' && (
            <p style={{ fontSize: 32, fontWeight: 600, margin: 0 }}>
              {lastValue != null ? lastValue.toFixed(2) : '—'}
              {activeSeries.unit && <span className="muted" style={{ fontSize: 16 }}> {activeSeries.unit}</span>}
            </p>
          )}
          {chartType === 'table' && (
            <table className="data-table">
              <thead><tr><th>Time</th><th>Value</th></tr></thead>
              <tbody>
                {chartData.slice(-20).map((row) => (
                  <tr key={row.time}><td>{row.time}</td><td>{row.value.toFixed(4)}</td></tr>
                ))}
              </tbody>
            </table>
          )}
          {chartType === 'line' && (
            <ResponsiveContainer width="100%" height={280}>
              <LineChart data={chartData}>
                <XAxis dataKey="time" {...chartDefaults.axis} />
                <YAxis {...chartDefaults.axis} />
                <Tooltip contentStyle={chartDefaults.tooltipStyle} />
                <Line type="monotone" dataKey="value" stroke="var(--accent-primary)" dot={false} strokeWidth={2} />
              </LineChart>
            </ResponsiveContainer>
          )}
          {chartType === 'bar' && (
            <ResponsiveContainer width="100%" height={280}>
              <BarChart data={chartData}>
                <XAxis dataKey="time" {...chartDefaults.axis} />
                <YAxis {...chartDefaults.axis} />
                <Tooltip contentStyle={chartDefaults.tooltipStyle} />
                <Bar dataKey="value" fill="var(--accent-primary)" />
              </BarChart>
            </ResponsiveContainer>
          )}
        </Card>
      )}
    </div>
  );
}
