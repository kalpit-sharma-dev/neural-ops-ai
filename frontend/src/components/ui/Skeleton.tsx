import './ui.css';

interface SkeletonProps {
  width?: string | number;
  height?: string | number;
  className?: string;
}

export function Skeleton({ width = '100%', height = 16, className = '' }: SkeletonProps) {
  return (
    <div
      className={`ui-skeleton ${className}`.trim()}
      style={{ width, height }}
      aria-hidden="true"
    />
  );
}

export function CardSkeleton() {
  return (
    <div className="ui-card">
      <Skeleton height={20} width="40%" />
      <Skeleton height={36} width="60%" className="mt-3" />
      <Skeleton height={48} className="mt-3" />
    </div>
  );
}
