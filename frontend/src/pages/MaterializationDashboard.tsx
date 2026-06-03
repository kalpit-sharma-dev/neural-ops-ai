import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import {
  fetchAlertPolicies,
  fetchAlertPolicyScores,
  fetchDerivedMetricSamples,
  triggerAlertPolicy,
} from '../api/observability';
import { StitchPageShell } from '../components/stitch';
import { Card } from '../components/ui/Card';
import { Select } from '../components/ui/Select';
import { Button } from '../components/ui/Button';
import { LoadingState } from '../components/ui/PageStates';
import { Badge } from '../components/ui/Badge';
import { useI18n } from '../i18n/I18nProvider';

export default function MaterializationDashboard() {
  const { t } = useI18n();
  const policiesQuery = useQuery({ queryKey: ['alert-policies'], queryFn: fetchAlertPolicies });
  const [policyId, setPolicyId] = useState('');
  const [metricId, setMetricId] = useState('dm-1');

  const scoresQuery = useQuery({
    queryKey: ['alert-scores', policyId],
    queryFn: () => fetchAlertPolicyScores(policyId),
    enabled: !!policyId,
  });

  const samplesQuery = useQuery({
    queryKey: ['derived-samples', metricId],
    queryFn: () => fetchDerivedMetricSamples(metricId),
    enabled: !!metricId,
  });

  return (
    <StitchPageShell title={t('mat.title')} subtitle={t('mat.subtitle')}>
      <Card title={t('mat.alertScores')}>
        <Select label={t('mat.policy')} value={policyId} onChange={(e) => setPolicyId(e.target.value)}>
          <option value="">—</option>
          {(policiesQuery.data ?? []).map((p) => (
            <option key={p.id} value={p.id}>
              {p.name}
            </option>
          ))}
        </Select>
        {policyId && (
          <Button
            variant="secondary"
            size="sm"
            style={{ marginTop: 8 }}
            data-testid="mat-trigger-policy"
            onClick={() => {
              void triggerAlertPolicy(policyId, { service: 'payment-service', severity: 'P1' }).then(() =>
                scoresQuery.refetch(),
              );
            }}
          >
            {t('mat.triggerDemo')}
          </Button>
        )}
        {scoresQuery.isLoading && policyId && <LoadingState />}
        {(scoresQuery.data ?? []).map((s) => (
          <div key={`${s.policyId}-${s.bucketTs}`} className="list-row">
            <span>{s.service}</span>
            <Badge variant={s.fatigueScore > 5 ? 'critical' : 'warning'}>
              fatigue {s.fatigueScore.toFixed(2)}
            </Badge>
            <span className="muted">{new Date(s.bucketTs).toLocaleString()}</span>
          </div>
        ))}
      </Card>

      <Card title={t('mat.derivedSamples')}>
        <Select label={t('mat.metric')} value={metricId} onChange={(e) => setMetricId(e.target.value)}>
          <option value="dm-1">checkout_error_budget</option>
        </Select>
        {samplesQuery.isLoading && <LoadingState />}
        {(samplesQuery.data ?? []).map((s) => (
          <div key={s.bucketTs} className="list-row">
            <span>{s.service}</span>
            <strong>{s.value.toFixed(3)}</strong>
            <span className="muted">n={s.sampleCount}</span>
          </div>
        ))}
      </Card>
    </StitchPageShell>
  );
}
