import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import toast from 'react-hot-toast';
import {
  fetchFinOpsReports,
  fetchNFRCertification,
  fetchAuditLog,
  type FinOpsReport,
} from '../api/observability';
import { fetchAlerts, fetchAlertRules } from '../api/alerts';
import { fetchIncidents } from '../api/incidents';
import { getApiErrorMessage } from '../api/client';
import { DataExportMenu } from '../components/ui/DataExportMenu';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { ErrorState, LoadingState } from '../components/ui/PageStates';
import { StitchPageShell } from '../components/stitch';
import { exportJson, downloadFromUrl } from '../utils/dataExport';

const DOMAIN_EXPORTS: { title: string; desc: string; to: string }[] = [
  { title: 'Logs', desc: 'Search, filter, and export log entries for any time range.', to: '/logs' },
  { title: 'Traces', desc: 'Export trace search results and full span waterfalls.', to: '/traces' },
  { title: 'Metrics', desc: 'Download metric time series and PromQL query results.', to: '/metrics' },
  { title: 'Incidents', desc: 'Export incident list and per-incident timelines.', to: '/incidents' },
  { title: 'Alerts', desc: 'Export firing alerts, rules, and silences.', to: '/alerts' },
  { title: 'Security', desc: 'Vulnerabilities, attacks, and findings.', to: '/security' },
  { title: 'Query workbench', desc: 'Cross-signal NexQL / unified query results.', to: '/query-workbench' },
  { title: 'Service map', desc: 'Topology and dependency graph JSON.', to: '/service-map' },
  { title: 'Trace compare', desc: 'Side-by-side trace comparison export.', to: '/traces/compare' },
  { title: 'Entities', desc: 'Per-service metrics, traces, and overview (open from service map).', to: '/service-map' },
  { title: 'Notebooks', desc: 'Notebook definitions and cell outputs.', to: '/notebooks' },
  { title: 'Workflows', desc: 'Automation workflow definitions.', to: '/workflows' },
  { title: 'SLOs', desc: 'SLO targets, budgets, and burn rates.', to: '/slos' },
  { title: 'Collectors', desc: 'Fleet agents and pipeline configs.', to: '/collectors/fleet' },
  { title: 'Custom dashboards', desc: 'Dashboard definitions and tile layouts.', to: '/dashboards' },
  { title: 'Infrastructure', desc: 'Host inventory and utilization.', to: '/infrastructure' },
  { title: 'Kubernetes', desc: 'Clusters, deployments, and pods.', to: '/kubernetes' },
  { title: 'AI Ops', desc: 'RCA, forecasts, and AutoFix plans.', to: '/ai-ops' },
  { title: 'Business observability', desc: 'RUM funnels, synthetic tests, KPI packs.', to: '/business-observability' },
  { title: 'Governance', desc: 'ABAC, residency, MSP, and exports.', to: '/enterprise-governance' },
];

async function downloadFinOpsReport(report: FinOpsReport) {
  try {
    if (report.url.startsWith('http')) {
      await downloadFromUrl(report.url, `${report.name}.${report.format}`);
    } else {
      exportJson(report, `finops-report-${report.id}`, { name: report.name, scope: report.scope });
    }
    toast.success(`Downloaded ${report.name}`);
  } catch (e) {
    toast.error(getApiErrorMessage(e));
  }
}

export default function ReportsHub() {
  const finopsQuery = useQuery({ queryKey: ['reports-finops'], queryFn: () => fetchFinOpsReports() });
  const certQuery = useQuery({ queryKey: ['reports-nfr-cert'], queryFn: fetchNFRCertification });
  const auditQuery = useQuery({ queryKey: ['reports-audit'], queryFn: fetchAuditLog });
  const incidentsQuery = useQuery({ queryKey: ['reports-incidents'], queryFn: () => fetchIncidents({ size: 200 }) });
  const alertsQuery = useQuery({ queryKey: ['reports-alerts'], queryFn: () => fetchAlerts({ size: 200 }) });
  const rulesQuery = useQuery({ queryKey: ['reports-alert-rules'], queryFn: fetchAlertRules });

  return (
    <StitchPageShell
      title="Reports & exports"
      subtitle="Central hub for generated reports and data downloads across observability domains"
      actions={
        <DataExportMenu
          getData={() => [
            ...(incidentsQuery.data ?? []).map((i) => ({ type: 'incident', ...i })),
            ...(alertsQuery.data ?? []).map((a) => ({ type: 'alert', ...a })),
          ]}
          filenamePrefix="operational-snapshot"
          meta={{ domains: ['incidents', 'alerts'] }}
          label="Export ops snapshot"
          disabled={!incidentsQuery.data?.length && !alertsQuery.data?.length}
        />
      }
    >
      <div className="reports-hub__grid">
        {DOMAIN_EXPORTS.map((d) => (
          <Card key={d.to} title={d.title}>
            <p className="muted" style={{ marginTop: 0 }}>{d.desc}</p>
            <a href={d.to} className="ui-button ui-button--ghost ui-button--sm" style={{ marginTop: 8 }}>
              Open & export →
            </a>
          </Card>
        ))}
      </div>

      <Card title="FinOps generated reports" style={{ marginTop: 24 }}>
        {finopsQuery.isLoading && <LoadingState />}
        {finopsQuery.error && <ErrorState message={getApiErrorMessage(finopsQuery.error)} onRetry={() => finopsQuery.refetch()} />}
        {!finopsQuery.isLoading && (
          <table className="data-table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Format</th>
                <th>Scope</th>
                <th>Generated</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {(finopsQuery.data ?? []).map((r) => (
                <tr key={r.id}>
                  <td>{r.name}</td>
                  <td>{r.format}</td>
                  <td>{r.scope}</td>
                  <td>{new Date(r.generatedAt).toLocaleString()}</td>
                  <td>
                    <Button variant="ghost" size="sm" onClick={() => void downloadFinOpsReport(r)}>
                      Download
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
        {(finopsQuery.data?.length ?? 0) === 0 && !finopsQuery.isLoading && (
          <p className="muted">No FinOps reports yet. Schedule from Cloud & Network → Reports.</p>
        )}
      </Card>

      <div className="dashboard-row-2" style={{ marginTop: 24 }}>
        <Card title="Compliance & certification">
          {certQuery.isLoading && <LoadingState />}
          {certQuery.data && (
            <>
              <p>
                NFR certification v{certQuery.data.version} —{' '}
                {certQuery.data.overallPass ? 'PASS' : 'FAIL'}
              </p>
              <DataExportMenu
                getData={() => [certQuery.data!]}
                filenamePrefix="nfr-certification"
                formats={['json']}
                label="Download certification"
              />
            </>
          )}
          <p className="muted" style={{ marginTop: 12 }}>
            <Link to="/nfr-certification">Open NFR certification →</Link>
          </p>
        </Card>

        <Card title="Audit log export">
          {auditQuery.isLoading && <LoadingState />}
          <DataExportMenu
            getData={() => auditQuery.data ?? []}
            filenamePrefix="audit-log"
            label="Download audit log"
            disabled={!auditQuery.data?.length}
          />
          <p className="muted" style={{ marginTop: 12 }}>
            <Link to="/settings/audit">Settings audit log →</Link>
          </p>
        </Card>
      </div>

      <Card title="Quick operational exports" style={{ marginTop: 24 }}>
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
          <DataExportMenu
            getData={() => incidentsQuery.data ?? []}
            filenamePrefix="incidents"
            label="Incidents"
            disabled={!incidentsQuery.data?.length}
          />
          <DataExportMenu
            getData={() => alertsQuery.data ?? []}
            filenamePrefix="alerts"
            label="Alerts"
            disabled={!alertsQuery.data?.length}
          />
          <DataExportMenu
            getData={() => rulesQuery.data ?? []}
            filenamePrefix="alert-rules"
            label="Alert rules"
            disabled={!rulesQuery.data?.length}
          />
        </div>
      </Card>
    </StitchPageShell>
  );
}
