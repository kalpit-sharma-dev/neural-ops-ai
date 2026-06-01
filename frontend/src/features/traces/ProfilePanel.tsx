import { useQuery } from '@tanstack/react-query';
import { fetchProfiles } from '../../api/observability';
import { Card } from '../../components/ui/Card';
import { LoadingState } from '../../components/ui/PageStates';

interface ProfilePanelProps {
  service: string;
}

export function ProfilePanel({ service }: ProfilePanelProps) {
  const { data, isLoading } = useQuery({
    queryKey: ['profiles', service],
    queryFn: () => fetchProfiles(service),
    enabled: service.length > 0,
  });

  if (isLoading) return <LoadingState />;
  if (!data?.length) return null;

  return (
    <Card title="Code hotspots (continuous profiling)">
      <table className="data-table">
        <thead>
          <tr>
            <th>Function</th>
            <th>Location</th>
            <th>Self time</th>
            <th>Samples</th>
          </tr>
        </thead>
        <tbody>
          {data.map((p: { functionName: string; filePath?: string; lineNo: number; selfTimeMs: number; sampleCount: number }) => (
            <tr key={`${p.functionName}-${p.lineNo}`}>
              <td>{p.functionName}</td>
              <td className="muted">{p.filePath}:{p.lineNo}</td>
              <td>{p.selfTimeMs.toFixed(1)}ms</td>
              <td>{p.sampleCount.toLocaleString()}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </Card>
  );
}
