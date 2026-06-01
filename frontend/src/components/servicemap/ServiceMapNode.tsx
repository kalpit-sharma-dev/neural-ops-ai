import type { Node } from '@xyflow/react';
import { Line, LineChart, ResponsiveContainer } from 'recharts';
import { Handle, Position, type NodeProps } from '@xyflow/react';
import { StatusDot } from '../ui/StatusDot';

export interface ServiceNodeData extends Record<string, unknown> {
  label: string;
  errorRate: number;
  p99: number;
  rps: number;
  health: 'healthy' | 'degraded' | 'down';
  sparkline: { t: number; v: number }[];
  blastRadius?: boolean;
  firstFailing?: boolean;
  criticalPath?: boolean;
}

export function ServiceMapNode({ data }: NodeProps<Node<ServiceNodeData>>) {
  const borderColor =
    data.health === 'down'
      ? 'var(--error)'
      : data.health === 'degraded'
        ? 'var(--warning)'
        : 'var(--status-healthy)';

  return (
    <div
      className={`service-map-node ${data.blastRadius ? 'service-map-node--blast' : ''} ${data.firstFailing ? 'service-map-node--failing' : ''} ${data.criticalPath ? 'service-map-node--critical' : ''}`}
      style={{ borderColor, boxShadow: data.health === 'down' ? 'var(--glow-red)' : undefined }}
    >
      <Handle type="target" position={Position.Left} />
      <div className="service-map-node__header">
        <StatusDot status={data.health} pulse={data.health !== 'healthy'} />
        <strong>{data.label}</strong>
      </div>
      <div className="service-map-node__metrics">
        <span>Err {data.errorRate.toFixed(1)}%</span>
        <span>p99 {Math.round(data.p99)}ms</span>
        <span>{Math.round(data.rps)} req/s</span>
      </div>
      <div className="service-map-node__sparkline">
        <ResponsiveContainer width="100%" height={28}>
          <LineChart data={data.sparkline}>
            <Line type="monotone" dataKey="v" stroke="var(--error)" dot={false} strokeWidth={1.5} />
          </LineChart>
        </ResponsiveContainer>
      </div>
      <Handle type="source" position={Position.Right} />
    </div>
  );
}

export const serviceMapNodeTypes = { service: ServiceMapNode };
