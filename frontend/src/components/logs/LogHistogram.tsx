import { useEffect, useMemo, useState } from 'react';
import { Bar, BarChart, Brush, ResponsiveContainer, Tooltip, XAxis } from 'recharts';
import { chartAxisProps, chartTooltipStyle } from '../../lib/chartTheme';
import type { LogHit } from '../../api/types';

interface LogHistogramProps {
  hits: LogHit[];
  bucketMinutes?: number;
  onBrush?: (start: Date, end: Date) => void;
}

export function LogHistogram({ hits, bucketMinutes = 15, onBrush }: LogHistogramProps) {
  const [brushRange, setBrushRange] = useState<{ startIndex: number; endIndex: number } | null>(null);

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
      .map(([ts, count], index) => ({
        index,
        label: new Date(ts).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
        count,
        start: new Date(ts),
        end: new Date(ts + bucketMs),
      }));
  }, [hits, bucketMinutes]);

  useEffect(() => {
    setBrushRange(null);
  }, [hits]);

  if (data.length === 0) return null;

  const applyBrush = (startIndex: number, endIndex: number) => {
    const startRow = data[startIndex];
    const endRow = data[endIndex];
    if (startRow?.start && endRow?.end && onBrush) {
      onBrush(startRow.start, endRow.end);
    }
  };

  return (
    <div className="log-histogram" role="img" aria-label="Log volume over time — click or drag to narrow range">
      {onBrush && (
        <p className="log-histogram__hint muted">Click a bar or drag the brush below to filter by time</p>
      )}
      <ResponsiveContainer width="100%" height={onBrush ? 96 : 72}>
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
          {onBrush && data.length > 2 && (
            <Brush
              dataKey="label"
              height={22}
              stroke="var(--accent-primary)"
              fill="var(--accent-primary-subtle)"
              onChange={(range) => {
                if (range && typeof range.startIndex === 'number' && typeof range.endIndex === 'number') {
                  setBrushRange({ startIndex: range.startIndex, endIndex: range.endIndex });
                }
              }}
            />
          )}
        </BarChart>
      </ResponsiveContainer>
      {brushRange && onBrush && (
        <button
          type="button"
          className="ui-button ui-button--ghost ui-button--sm log-histogram__apply"
          onClick={() => applyBrush(brushRange.startIndex, brushRange.endIndex)}
        >
          Apply brush selection
        </button>
      )}
    </div>
  );
}
