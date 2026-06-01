import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { useState } from 'react';
import toast from 'react-hot-toast';
import {
  createOncallSchedule,
  deleteOncallSchedule,
  fetchOncallSchedules,
  syncOncallPagerDuty,
  updateOncallSchedule,
} from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState, PageHeader } from '../components/ui/PageStates';

export default function SettingsOncall() {
  const queryClient = useQueryClient();
  const { data, isLoading } = useQuery({ queryKey: ['oncall'], queryFn: fetchOncallSchedules });
  const [team, setTeam] = useState('');
  const [pdScheduleId, setPdScheduleId] = useState('');

  const invalidate = () => void queryClient.invalidateQueries({ queryKey: ['oncall'] });

  const createMut = useMutation({
    mutationFn: () =>
      createOncallSchedule({
        team,
        timezone: 'UTC',
        enabled: true,
        rotation: [
          { name: 'Primary', email: 'oncall@neuralops.ai', after: '0m' },
          { name: 'Secondary', email: 'backup@neuralops.ai', after: '30m' },
        ],
      }),
    onSuccess: () => { toast.success('On-call schedule created'); setTeam(''); invalidate(); },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const deleteMut = useMutation({
    mutationFn: deleteOncallSchedule,
    onSuccess: () => { toast.success('Schedule deleted'); invalidate(); },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const syncMut = useMutation({
    mutationFn: ({ id, scheduleId }: { id: string; scheduleId?: string }) => syncOncallPagerDuty(id, scheduleId),
    onSuccess: () => { toast.success('Synced from PagerDuty'); invalidate(); },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const toggleMut = useMutation({
    mutationFn: (s: NonNullable<typeof data>[number]) =>
      updateOncallSchedule(s.id, { team: s.team, timezone: s.timezone, enabled: !s.enabled, rotation: s.rotation }),
    onSuccess: () => { toast.success('Schedule updated'); invalidate(); },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  return (
    <div>
      <PageHeader title="On-call schedules" subtitle="Escalation rotation and PagerDuty sync" actions={<Link to="/settings">← Settings</Link>} />
      <Card title="New schedule" style={{ marginBottom: 24 }}>
        <div className="form-stack">
          <input placeholder="Team name" value={team} onChange={(e) => setTeam(e.target.value)} />
          <Button variant="primary" disabled={!team} onClick={() => createMut.mutate()}>Create schedule</Button>
        </div>
      </Card>
      <Card title="PagerDuty sync" style={{ marginBottom: 24 }}>
        <div className="form-stack">
          <input placeholder="PagerDuty schedule ID (optional override)" value={pdScheduleId} onChange={(e) => setPdScheduleId(e.target.value)} />
        </div>
      </Card>
      {isLoading && <LoadingState />}
      {(data ?? []).map((s) => (
        <Card
          key={s.id}
          title={s.team}
          action={
            <div style={{ display: 'flex', gap: 8 }}>
              <Button variant="secondary" size="sm" onClick={() => toggleMut.mutate(s)}>
                {s.enabled ? 'Disable' : 'Enable'}
              </Button>
              <Button variant="secondary" size="sm" onClick={() => syncMut.mutate({ id: s.id, scheduleId: pdScheduleId || undefined })}>
                Sync PagerDuty
              </Button>
              <Button variant="secondary" size="sm" onClick={() => deleteMut.mutate(s.id)}>Delete</Button>
            </div>
          }
        >
          <Badge variant={s.enabled ? 'healthy' : 'info'}>{s.timezone}</Badge>
          <table className="data-table" style={{ marginTop: 12 }}>
            <thead><tr><th>Name</th><th>Email</th><th>After</th></tr></thead>
            <tbody>
              {s.rotation.map((r) => (
                <tr key={r.email}>
                  <td>{r.name}</td>
                  <td>{r.email}</td>
                  <td>{r.after}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      ))}
    </div>
  );
}
