import { ReactNode } from 'react';
import { Activity, Bell, Flame, GitBranch, Globe, Server, Shield, Terminal } from 'lucide-react';
import { EmptyState } from './EmptyState';
import { useI18n } from '../../i18n/I18nProvider';

export type EmptyDomain =
  | 'logs'
  | 'incidents'
  | 'traces'
  | 'alerts'
  | 'rum'
  | 'kubernetes'
  | 'infrastructure'
  | 'security'
  | 'generic';

const PRESETS: Record<
  EmptyDomain,
  { title: string; description: string; icon: ReactNode; actionLabel?: string }
> = {
  logs: {
    title: 'No logs in this window',
    description: 'Broaden the time range or switch to AI search for natural language queries.',
    icon: <Terminal size={32} />,
    actionLabel: 'Try AI search',
  },
  incidents: {
    title: 'No incidents',
    description: 'All clear — no active incidents match your filters.',
    icon: <Flame size={32} />,
  },
  traces: {
    title: 'No traces found',
    description: 'Adjust service filters or widen the time range.',
    icon: <GitBranch size={32} />,
  },
  alerts: {
    title: 'No alerts',
    description: 'No firing alerts in the selected environment.',
    icon: <Bell size={32} />,
  },
  rum: {
    title: 'No sessions yet',
    description: 'Install the RUM SDK to collect real user sessions.',
    icon: <Globe size={32} />,
  },
  kubernetes: {
    title: 'No workloads',
    description: 'Connect a cluster collector or check agent health.',
    icon: <Server size={32} />,
  },
  infrastructure: {
    title: 'No hosts',
    description: 'Deploy the host agent to populate infrastructure inventory.',
    icon: <Server size={32} />,
  },
  security: {
    title: 'No threats detected',
    description: 'Security events will appear here when rules trigger.',
    icon: <Shield size={32} />,
  },
  generic: {
    title: 'Nothing here yet',
    description: 'Data will appear when collectors and APIs are connected.',
    icon: <Activity size={32} />,
  },
};

interface DomainEmptyStateProps {
  domain: EmptyDomain;
  onAction?: () => void;
  actionLabel?: string;
}

export function DomainEmptyState({ domain, onAction, actionLabel }: DomainEmptyStateProps) {
  const { t } = useI18n();
  const preset = PRESETS[domain];
  // t() falls back to the key when a translation is missing, so we keep the
  // English preset copy as the resilient default for every locale.
  const tr = (key: string, fallback: string) => {
    const value = t(key);
    return value === key ? fallback : value;
  };
  const presetActionKey = domain === 'logs' ? 'empty.logs.action' : undefined;
  const presetActionLabel = presetActionKey
    ? tr(presetActionKey, preset.actionLabel ?? '')
    : preset.actionLabel;
  return (
    <EmptyState
      title={tr(`empty.${domain}.title`, preset.title)}
      description={tr(`empty.${domain}.desc`, preset.description)}
      icon={preset.icon}
      actionLabel={actionLabel ?? presetActionLabel}
      onAction={onAction}
    />
  );
}
