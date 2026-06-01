import { useMemo } from 'react';
import { Bar, BarChart, ResponsiveContainer, Tooltip, XAxis } from 'recharts';
import { chartAxisProps, chartTooltipStyle } from '../../lib/chartTheme';
import type { LogHit } from '../../api/types';

interface LogHistogramProps {
  hits: LogHit[];
  bucketMinutes?: number;
  onBrush?: (start: Date, end: Date) => void;
}

export function LogHistogram({ hits, bucketMinutes = 15, onBrush }: LogHistogramProps) {
  const data = useMemo(() => {
    if (hits.length === 0) return [];
    const bucketMs = bucketMinutes * 60 * 1000;
    const buckets = new Map<number, number>();
    hits.forEach((h) => {
      const t = new Date(h.timestamp).getTime();
      const key = Math.floor(t / bucketMs) * bucketMs;
      buckets.set(key, (buckets.get(key) ?? 0) + 1);
    });
    return Array.from(buckets.entries())
      .sort(([a], [b]) => a - b)
      .map(([ts, count]) => ({
        label: new Date(ts).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
        count,
        start: new Date(ts),
        end: new Date(ts + bucketMs),
      }));
  }, [hits, bucketMinutes]);

  if (data.length === 0) return null;

  return (
    <div className="log-histogram" role="img" aria-label="Log volume over time">
      <ResponsiveContainer width="100%" height={72}>
        <BarChart data={data} margin={{ top: 4, right: 8, left: 0, bottom: 0 }}>
          <XAxis dataKey="label" {...chartAxisProps()} />
          <Tooltip contentStyle={chartTooltipStyle()} />
          <Bar
            dataKey="count"
            fill="var(--accent-primary)"
            radius={[2, 2, 0, 0]}
            onClick={(bar) => {
              const payload = bar?.payload as { start?: Date; end?: Date } | undefined;
              if (payload?.start && payload?.end && onBrush) onBrush(payload.start, payload.end);
            }}
            style={{ cursor: onBrush ? 'pointer' : 'default' }}
          />
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}
