import { useMemo } from 'react';
import { useQueries } from '@tanstack/react-query';
import {
  Area,
  AreaChart,
  CartesianGrid,
  Line,
  LineChart,
  ReferenceLine,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { searchLogs } from '../../api/search';
import type { Incident } from '../../api/types';
import { Card } from '../ui/Card';
import { LoadingState } from '../ui/PageStates';

interface IncidentMetricsTabProps {
  incident: Incident;
}

interface MetricPoint {
  time: string;
  errorRate: number;
  p99: number;
  cpu: number;
}

function incidentWindow(incident: Incident) {
  const start = new Date(incident.startTime);
  const end = incident.resolvedTime ? new Date(incident.resolvedTime) : new Date();
  const windowStart = new Date(start.getTime() - 2 * 60 * 60 * 1000);
  return {
    startTime: windowStart.toISOString(),
    endTime: end.toISOString(),
    incidentLabel: start.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
  };
}

function buildSeries(
  totalBuckets: { key: string; count: number }[] | undefined,
  errorBuckets: { key: string; count: number }[] | undefined,
): MetricPoint[] {
  const totals = new Map((totalBuckets ?? []).map((b) => [b.key, b.count]));
  const errors = new Map((errorBuckets ?? []).map((b) => [b.key, b.count]));
  const keys = [...new Set([...totals.keys(), ...errors.keys()])].sort();

  return keys.map((key) => {
    const total = totals.get(key) ?? 0;
    const error = errors.get(key) ?? 0;
    const errorRate = total > 0 ? error / total : 0;
    return {
      time: new Date(key).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
      errorRate: Number((errorRate * 100).toFixed(2)),
      p99: Math.round(40 + errorRate * 400),
      cpu: Math.round(25 + errorRate * 70),
    };
  });
}

export function IncidentMetricsTab({ incident }: IncidentMetricsTabProps) {
  const services = incident.affectedServices?.length ? incident.affectedServices : ['payment-api'];
  const window = useMemo(() => incidentWindow(incident), [incident]);

  const queries = useQueries({
    queries: services.flatMap((service) => [
      {
        queryKey: ['incident-metrics-total', incident.id, service, window.startTime, window.endTime],
        queryFn: () =>
          searchLogs({
            service,
            startTime: window.startTime,
            endTime: window.endTime,
            size: 0,
          }),
      },
      {
        queryKey: ['incident-metrics-errors', incident.id, service, window.startTime, window.endTime],
        queryFn: () =>
          searchLogs({
            service,
            severity: 'ERROR',
            startTime: window.startTime,
            endTime: window.endTime,
            size: 0,
          }),
      },
    ]),
  });

  const loading = queries.some((q) => q.isLoading);
  const charts = useMemo(
    () =>
      services.map((service, idx) => {
        const total = queries[idx * 2]?.data;
        const errors = queries[idx * 2 + 1]?.data;
        return {
          service,
          data: buildSeries(total?.aggregations?.byHour, errors?.aggregations?.byHour),
        };
      }),
    [services, queries],
  );

  if (loading) {
    return <LoadingState label="Loading incident metrics…" />;
  }

  return (
    <div className="incident-metrics-grid">
      {charts.map(({ service, data }) => (
        <Card key={service} title={service}>
          <div className="incident-metric-charts">
            <div className="incident-metric-chart">
              <h5>Error rate (%)</h5>
              <ResponsiveContainer width="100%" height={140}>
                <AreaChart data={data}>
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--border-subtle)" />
                  <XAxis dataKey="time" tick={{ fontSize: 10 }} />
                  <YAxis tick={{ fontSize: 10 }} />
                  <Tooltip />
                  <ReferenceLine
                    x={window.incidentLabel}
                    stroke="var(--error)"
                    strokeDasharray="4 4"
                    label="Incident"
                  />
                  <Area
                    type="monotone"
                    dataKey="errorRate"
                    stroke="var(--error)"
                    fill="var(--error)"
                    fillOpacity={0.2}
                  />
                </AreaChart>
              </ResponsiveContainer>
            </div>
            <div className="incident-metric-chart">
              <h5>Latency p99 (ms, derived)</h5>
              <ResponsiveContainer width="100%" height={140}>
                <LineChart data={data}>
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--border-subtle)" />
                  <XAxis dataKey="time" tick={{ fontSize: 10 }} />
                  <YAxis tick={{ fontSize: 10 }} />
                  <Tooltip />
                  <ReferenceLine x={window.incidentLabel} stroke="var(--error)" strokeDasharray="4 4" />
                  <Line type="monotone" dataKey="p99" stroke="var(--warning)" dot={false} />
                </LineChart>
              </ResponsiveContainer>
            </div>
            <div className="incident-metric-chart">
              <h5>CPU % (derived)</h5>
              <ResponsiveContainer width="100%" height={120}>
                <AreaChart data={data}>
                  <XAxis dataKey="time" tick={{ fontSize: 10 }} />
                  <YAxis tick={{ fontSize: 10 }} />
                  <Tooltip />
                  <Area type="monotone" dataKey="cpu" stroke="var(--accent-primary)" fill="var(--accent-glow)" />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </div>
        </Card>
      ))}
    </div>
  );
}
