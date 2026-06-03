import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import { Link, useParams } from '@tanstack/react-router';
import { correlateSecurityFinding, fetchSecurityFinding } from '../api/observability';
import { Button } from '../components/ui/Button';
import { getApiErrorMessage } from '../api/client';
import { StitchPageShell } from '../components/stitch';
import { Card } from '../components/ui/Card';
import { Badge } from '../components/ui/Badge';
import { ErrorState, LoadingState } from '../components/ui/PageStates';

export default function SecurityFindingDetail() {
  const { id = '' } = useParams({ strict: false });
  const qc = useQueryClient();
  const findingQuery = useQuery({
    queryKey: ['security-finding', id],
    queryFn: () => fetchSecurityFinding(id),
    enabled: !!id,
  });

  const correlateMut = useMutation({
    mutationFn: () => correlateSecurityFinding(id),
    onSuccess: () => {
      toast.success('Finding correlated to incident');
      void qc.invalidateQueries({ queryKey: ['security-finding', id] });
    },
  });

  if (findingQuery.isLoading) return <LoadingState />;
  if (findingQuery.error) {
    return <ErrorState message={getApiErrorMessage(findingQuery.error)} onRetry={() => findingQuery.refetch()} />;
  }
  const f = findingQuery.data;
  if (!f) return null;

  return (
    <StitchPageShell
      title={f.title}
      subtitle={`${f.category} · ${f.service ?? f.asset ?? 'unknown asset'}`}
      breadcrumb={
        <p className="muted">
          <Link to="/security">Security</Link> › {f.id}
        </p>
      }
    >
      <Card title="Finding">
        <div className="list-row">
          <Badge variant={f.severity === 'HIGH' ? 'critical' : 'warning'}>{f.severity}</Badge>
          <Badge variant={f.status === 'OPEN' ? 'critical' : 'healthy'}>{f.status}</Badge>
        </div>
        <p>Detected {new Date(f.detectedAt).toLocaleString()}</p>
        {f.exploitability && <p>Exploitability: {f.exploitability}</p>}
      </Card>
      <Card title="Incident correlation">
        {f.incidentId ? (
          <Link to="/incidents/$id" params={{ id: f.incidentId }}>
            Open incident {f.incidentId} →
          </Link>
        ) : (
          <Button variant="primary" onClick={() => correlateMut.mutate()} disabled={correlateMut.isPending}>
            Correlate to incident
          </Button>
        )}
      </Card>
    </StitchPageShell>
  );
}
