const API_BASE = process.env.EXPO_PUBLIC_API_BASE ?? 'http://localhost:8080/api/v1';
const API_KEY = process.env.EXPO_PUBLIC_API_KEY ?? 'demo-api-key';
const TENANT_ID = process.env.EXPO_PUBLIC_TENANT_ID ?? '00000000-0000-0000-0000-000000000002';

async function getData<T>(path: string): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': API_KEY,
      'X-Tenant-ID': TENANT_ID,
    },
  });
  const json = await res.json();
  if (json.status !== 'success') {
    throw new Error(json.message ?? 'API error');
  }
  return json.data as T;
}

export interface AlertRecord {
  id: string;
  title: string;
  severity: string;
  service: string;
  status: string;
}

export interface IncidentSummary {
  id: string;
  title: string;
  severity: string;
  status: string;
  service: string;
}

export function fetchAlerts() {
  return getData<AlertRecord[]>('/alerts?size=20');
}

export function fetchIncidents() {
  return getData<IncidentSummary[]>('/incidents?size=20');
}

export function fetchUsage() {
  return getData<{ logsIngestedGb: number; tracesIngested: number; activeUsers: number }>('/admin/usage');
}
