import { StitchPageShell, HubCardGrid, HubCard } from '../components/stitch';

const sections = [
  { title: 'Users & Teams', desc: 'Manage tenant users and RBAC roles', to: '/settings/users' },
  { title: 'API Keys', desc: 'Create and revoke API tokens', to: '/settings/api-keys' },
  { title: 'Audit Log', desc: 'Security and configuration audit trail', to: '/settings/audit' },
  { title: 'Usage', desc: 'Ingest volume and AI token consumption', to: '/settings/usage' },
  { title: 'Log Settings', desc: 'Parsing rules and log-based metrics', to: '/logs/settings' },
  { title: 'Integrations', desc: 'Jira, Slack, PagerDuty connections', to: '/integrations' },
  { title: 'SSO / IdP', desc: 'OpenID Connect provider configuration', to: '/settings/sso' },
  { title: 'Tenant policies', desc: 'Log retention and ingestion limits', to: '/settings/policies' },
  { title: 'On-call schedules', desc: 'Escalation rotation and paging', to: '/settings/oncall' },
  { title: 'Design System', desc: 'Component gallery', to: '/design-system' },
  { title: 'System Health', desc: 'Microservice health checks', to: '/health' },
];

export default function Settings() {
  return (
    <StitchPageShell title="Settings" subtitle="Platform configuration and administration">
      <HubCardGrid>
        {sections.map((s) => (
          <HubCard key={s.to} title={s.title} description={s.desc} to={s.to} />
        ))}
      </HubCardGrid>
    </StitchPageShell>
  );
}
