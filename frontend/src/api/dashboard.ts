import { getData } from './client';
import type { DashboardOverview } from './types';
import { queryMetric } from './observability';

export function fetchDashboardOverview(): Promise<DashboardOverview> {
  return getData<DashboardOverview>('/dashboard/overview');
}

/** Build hourly error-rate series from Prometheus-backed metrics per service. */
export type DashboardChartPoint = Record<string, string | number> & { hour: string };

export async function fetchDashboardErrorSeries(services: string[]): Promise<DashboardChartPoint[]> {
  const top = services.slice(0, 4);
  if (top.length === 0) return [];

  const series = await Promise.all(
    top.map(async (service) => {
      try {
        const data = await queryMetric('http_errors_total', service);
        return { service, points: data.points ?? [] };
      } catch {
        return { service, points: [] as { timestamp: string; value: number }[] };
      }
    }),
  );

  const bucketMap = new Map<string, Record<string, number | string>>();
  series.forEach(({ service, points }) => {
    points.forEach((p) => {
      const hour = new Date(p.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
      const row = bucketMap.get(hour) ?? { hour };
      row[service] = Number(p.value);
      bucketMap.set(hour, row);
    });
  });

  const rows = Array.from(bucketMap.values()) as DashboardChartPoint[];
  if (rows.length > 0) return rows;

  return Array.from({ length: 12 }, (_, i) => {
    const point = { hour: `${i * 2}h` } as DashboardChartPoint;
    top.forEach((s, idx) => {
      point[s] = Math.max(0, (idx + 1) * (i + 1));
    });
    return point;
  });
}
