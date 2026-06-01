import { Platform } from 'react-native';
import Constants from 'expo-constants';

const API_BASE = Constants.expoConfig?.extra?.apiBase ?? 'http://localhost:8080/api/v1';
const TENANT_ID = Constants.expoConfig?.extra?.tenantId ?? 'default';

export type RUMEvent = {
  sessionId: string;
  eventType: 'view' | 'action' | 'error' | 'timing';
  name: string;
  durationMs?: number;
  metadata?: Record<string, string>;
};

let sessionId = `mobile-${Date.now()}`;

export function getRUMSessionId() {
  return sessionId;
}

export async function recordRUMEvent(event: Omit<RUMEvent, 'sessionId'>) {
  const body = {
    sessionId,
    page: event.name,
    device: Platform.OS,
    durationMs: event.durationMs ?? 0,
    errors: event.eventType === 'error' ? 1 : 0,
    metadata: event.metadata,
    eventType: event.eventType,
  };
  try {
    await fetch(`${API_BASE}/rum/beacon`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'X-Tenant-ID': TENANT_ID },
      body: JSON.stringify(body),
    });
  } catch {
    /* offline */
  }
}

export function trackScreen(name: string) {
  void recordRUMEvent({ eventType: 'view', name });
}

export function trackAction(name: string, metadata?: Record<string, string>) {
  void recordRUMEvent({ eventType: 'action', name, metadata });
}

export function trackError(name: string, metadata?: Record<string, string>) {
  void recordRUMEvent({ eventType: 'error', name, metadata });
}
