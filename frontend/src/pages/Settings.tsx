import { StitchPageShell, HubCardGrid, HubCard } from '../components/stitch';
import { useI18n } from '../i18n/I18nProvider';

const SECTIONS = [
  { titleKey: 'settings.section.users', title: 'Users & Teams', descKey: 'settings.section.usersDesc', desc: 'Manage tenant users and RBAC roles', to: '/settings/users' },
  { titleKey: 'settings.section.apiKeys', title: 'API Keys', descKey: 'settings.section.apiKeysDesc', desc: 'Create and revoke API tokens', to: '/settings/api-keys' },
  { titleKey: 'settings.section.audit', title: 'Audit Log', descKey: 'settings.section.auditDesc', desc: 'Security and configuration audit trail', to: '/settings/audit' },
  { titleKey: 'settings.section.usage', title: 'Usage', descKey: 'settings.section.usageDesc', desc: 'Ingest volume and AI token consumption', to: '/settings/usage' },
  { titleKey: 'settings.section.logSettings', title: 'Log Settings', descKey: 'settings.section.logSettingsDesc', desc: 'Parsing rules and log-based metrics', to: '/logs/settings' },
  { titleKey: 'settings.section.integrations', title: 'Integrations', descKey: 'settings.section.integrationsDesc', desc: 'Jira, Slack, PagerDuty connections', to: '/integrations' },
  { titleKey: 'settings.section.sso', title: 'SSO / IdP', descKey: 'settings.section.ssoDesc', desc: 'OpenID Connect provider configuration', to: '/settings/sso' },
  { titleKey: 'settings.section.policies', title: 'Tenant policies', descKey: 'settings.section.policiesDesc', desc: 'Log retention and ingestion limits', to: '/settings/policies' },
  { titleKey: 'settings.section.oncall', title: 'On-call schedules', descKey: 'settings.section.oncallDesc', desc: 'Escalation rotation and paging', to: '/settings/oncall' },
  { titleKey: 'settings.section.platform', title: 'Platform capabilities', descKey: 'settings.section.platformDesc', desc: 'Open standards, AI features, and deployment modes', to: '/settings/platform' },
  { titleKey: 'settings.section.governance', title: 'Enterprise governance', descKey: 'settings.section.governanceDesc', desc: 'ABAC, residency, MSP branding, exports', to: '/enterprise-governance' },
  { titleKey: 'settings.section.designSystem', title: 'Design System', descKey: 'settings.section.designSystemDesc', desc: 'Component gallery', to: '/design-system' },
  { titleKey: 'settings.section.health', title: 'System Health', descKey: 'settings.section.healthDesc', desc: 'Microservice health checks', to: '/health' },
  { titleKey: 'settings.section.nfr', title: 'NFR certification', descKey: 'settings.section.nfrDesc', desc: 'Performance, DR, WCAG, and i18n readiness', to: '/nfr-certification' },
] as const;

export default function Settings() {
  const { tr } = useI18n();

  return (
    <StitchPageShell pageId="settings" title="Settings" subtitle="Platform configuration and administration">
      <HubCardGrid>
        {SECTIONS.map((s) => (
          <HubCard key={s.to} title={tr(s.titleKey, s.title)} description={tr(s.descKey, s.desc)} to={s.to} />
        ))}
      </HubCardGrid>
    </StitchPageShell>
  );
}
