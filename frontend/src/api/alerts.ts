import { deleteData, getData, postData, putData } from './client';

export interface AlertRecord {
  id: string;
  tenantId: string;
  source: string;
  alertName: string;
  service: string;
  title: string;
  description: string;
  severity: string;
  status: string;
  firedAt: string;
  lastSeenAt: string;
  linkedIncidentId?: string;
}

export interface AlertRule {
  id: string;
  name: string;
  source: string;
  servicePattern?: string;
  severity: string;
  enabled: boolean;
  labels?: Record<string, string>;
}

export interface NotificationChannel {
  id: string;
  name: string;
  channelType: string;
  config: Record<string, unknown>;
  enabled: boolean;
}

export interface Silence {
  id: string;
  servicePattern?: string;
  alertNamePattern?: string;
  reason?: string;
  startsAt: string;
  endsAt: string;
}

export async function fetchAlerts(params?: { status?: string; page?: number; size?: number }) {
  return getData<AlertRecord[]>('/alerts', params);
}

export async function acknowledgeAlert(id: string) {
  return postData<AlertRecord>(`/alerts/${id}/acknowledge`);
}

export async function suppressAlert(id: string, duration: string, reason: string) {
  return postData<AlertRecord>(`/alerts/${id}/suppress`, { duration, reason });
}

export async function fetchAlertRules() {
  return getData<AlertRule[]>('/alerts/rules');
}

export async function createAlertRule(body: Omit<AlertRule, 'id'>) {
  return postData<AlertRule>('/alerts/rules', body);
}

export async function updateAlertRule(id: string, body: Omit<AlertRule, 'id'>) {
  return putData<AlertRule>(`/alerts/rules/${id}`, body);
}

export async function deleteAlertRule(id: string) {
  return deleteData<{ deleted: boolean }>(`/alerts/rules/${id}`);
}

export async function fetchChannels() {
  return getData<NotificationChannel[]>('/notifications/channels');
}

export async function createChannel(body: Omit<NotificationChannel, 'id'>) {
  return postData<NotificationChannel>('/notifications/channels', body);
}

export async function updateChannel(id: string, body: Omit<NotificationChannel, 'id'>) {
  return putData<NotificationChannel>(`/notifications/channels/${id}`, body);
}

export async function deleteChannel(id: string) {
  return deleteData<{ deleted: boolean }>(`/notifications/channels/${id}`);
}

export async function testChannel(id: string) {
  return postData<{ sent: boolean }>(`/notifications/channels/${id}/test`);
}

export async function fetchSilences() {
  return getData<Silence[]>('/alerts/silences');
}

export async function createSilence(body: {
  servicePattern?: string;
  alertNamePattern?: string;
  reason?: string;
  duration: string;
}) {
  return postData<Silence>('/alerts/silences', body);
}

export interface EscalationLevel {
  level: number;
  after: string;
  target: string;
  channelType: string;
}

export interface EscalationPolicy {
  id: string;
  name: string;
  servicePattern?: string;
  levels: EscalationLevel[];
  enabled: boolean;
}

export async function fetchEscalationPolicies() {
  return getData<EscalationPolicy[]>('/escalation/policies');
}

export async function createEscalationPolicy(body: Omit<EscalationPolicy, 'id'>) {
  return postData<EscalationPolicy>('/escalation/policies', body);
}

export async function deleteEscalationPolicy(id: string) {
  return deleteData(`/escalation/policies/${id}`);
}
