import './ui.css';

type Status = 'healthy' | 'degraded' | 'down' | 'unknown';

interface StatusDotProps {
  status?: Status;
  pulse?: boolean;
  label?: string;
}

export function StatusDot({ status = 'unknown', pulse = false, label }: StatusDotProps) {
  return (
    <span className="ui-status-dot" title={label ?? status}>
      <span className={`ui-status-dot__inner ui-status-dot__inner--${status} ${pulse ? 'pulse' : ''}`} />
      {label && <span className="ui-status-dot__label">{label}</span>}
    </span>
  );
}
