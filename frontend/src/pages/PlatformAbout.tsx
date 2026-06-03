import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { fetchPlatformInfo, type PlatformCapabilities } from '../api/client';
import { StitchPageShell, SettingsBreadcrumb } from '../components/stitch';
import { Card } from '../components/ui/Card';
import { Badge } from '../components/ui/Badge';
import { LoadingState } from '../components/ui/PageStates';
import { useI18n } from '../i18n/I18nProvider';

function CapabilityBadges({ caps }: { caps: PlatformCapabilities }) {
  const flags = [
    ['OpenTelemetry', caps.openTelemetry],
    ['Prometheus', caps.prometheus],
    ['eBPF', caps.ebpf],
    ['OpenAPI', caps.openApi],
    ['Streaming', caps.streamingPipeline],
    ['Unified query', caps.unifiedQuery],
  ] as const;
  return (
    <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
      {flags.map(([label, on]) => (
        <Badge key={label} variant={on ? 'healthy' : 'info'}>
          {label}
        </Badge>
      ))}
    </div>
  );
}

export default function PlatformAbout() {
  const { tr } = useI18n();
  const { data, isLoading } = useQuery({ queryKey: ['platform-info'], queryFn: fetchPlatformInfo });

  return (
    <StitchPageShell
      pageId="platform-about"
      title="Platform capabilities"
      subtitle="Open standards, AI-native workflows, and unified observability vs tool sprawl"
      breadcrumb={<SettingsBreadcrumb page="Platform" />}
    >
      {isLoading && <LoadingState />}
      {data?.capabilities && (
        <>
          <Card title={tr('platform.whyTitle', 'Why NeuralOps')}>
            <p className="muted" style={{ marginBottom: 12 }}>
              Version {data.version} · {data.environment}
              {data.demoMode ? ` · ${tr('platform.demoMode', 'demo mode')}` : ''}
            </p>
            <CapabilityBadges caps={data.capabilities} />
            <ul className="insight-list" style={{ marginTop: 16 }}>
              {data.capabilities.differentiators.map((d) => (
                <li key={d}>{d.replace(/-/g, ' ')}</li>
              ))}
            </ul>
          </Card>

          <Card title={tr('platform.aiTitle', 'AI & automation')} style={{ marginTop: 24 }}>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
              {data.capabilities.aiFeatures.map((f) => (
                <Badge key={f} variant="info">
                  {f}
                </Badge>
              ))}
            </div>
            <p style={{ marginTop: 12 }}>
              <Link to="/ai-ops">AI Ops Center</Link> · <Link to="/query-workbench">Query Workbench</Link> ·{' '}
              <Link to="/workflows">Workflows</Link>
            </p>
          </Card>

          <Card title={tr('platform.govTitle', 'Governance & deployment')} style={{ marginTop: 24 }}>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8, marginBottom: 12 }}>
              {data.capabilities.governanceFeatures.map((f) => (
                <Badge key={f} variant="success">
                  {f}
                </Badge>
              ))}
            </div>
            <p className="muted">
              {tr('platform.deployment', 'Deployment')}: {data.capabilities.deploymentModes.join(', ')}
            </p>
            <p className="muted" style={{ marginTop: 8 }}>
              {tr('platform.standards', 'Standards')}: {data.capabilities.standards.join(' · ')}
            </p>
            <p style={{ marginTop: 12 }}>
              <Link to="/enterprise-governance">Enterprise governance</Link> ·{' '}
              <Link to="/nfr-certification">NFR certification</Link>
            </p>
          </Card>
        </>
      )}
    </StitchPageShell>
  );
}
