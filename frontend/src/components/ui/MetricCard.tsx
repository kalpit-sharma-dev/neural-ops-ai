import { ReactNode } from 'react';
import { TrendingDown, TrendingUp, Minus } from 'lucide-react';
import { Line, LineChart, ResponsiveContainer } from 'recharts';
import './ui.css';

interface MetricCardProps {
  label: string;
  value: string | number;
  trend?: number;
  sparkline?: number[];
  accent?: 'default' | 'danger' | 'success';
  footer?: ReactNode;
}

export function MetricCard({ label, value, trend, sparkline, accent = 'default', footer }: MetricCardProps) {
  const chartData = (sparkline ?? [3, 5, 4, 7, 6, 8, 5]).map((v, i) => ({ i, v }));
  const trendIcon =
    trend === undefined || trend === 0 ? (
      <Minus size={14} />
    ) : trend > 0 ? (
      <TrendingUp size={14} />
    ) : (
      <TrendingDown size={14} />
    );

  return (
    <article className={`ui-metric ui-metric--${accent}`}>
      <p className="ui-metric__label">{label}</p>
      <div className="ui-metric__row">
        <span className="ui-metric__value">{value}</span>
        {trend !== undefined && (
          <span className={`ui-metric__trend ${trend > 0 ? 'up' : trend < 0 ? 'down' : ''}`}>
            {trendIcon}
            {Math.abs(trend).toFixed(1)}%
          </span>
        )}
      </div>
      {sparkline && sparkline.length > 0 && (
        <div className="ui-metric__sparkline">
          <ResponsiveContainer width="100%" height={40}>
            <LineChart data={chartData}>
              <Line type="monotone" dataKey="v" stroke="var(--accent-primary)" strokeWidth={2} dot={false} />
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}
      {footer}
    </article>
  );
}
