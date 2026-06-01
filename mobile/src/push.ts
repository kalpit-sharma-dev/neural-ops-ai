import { useEffect, useRef, useState } from 'react';
import * as Notifications from 'expo-notifications';
import Constants from 'expo-constants';
import { Platform } from 'react-native';
import { fetchAlerts } from './src/api';

const API_BASE = Constants.expoConfig?.extra?.apiBase ?? 'http://localhost:8080/api/v1';
const TENANT_ID = Constants.expoConfig?.extra?.tenantId ?? 'default';

async function registerPushToken(token: string) {
  try {
    await fetch(`${API_BASE}/mobile/push/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'X-Tenant-ID': TENANT_ID },
      body: JSON.stringify({ expoPushToken: token, platform: Platform.OS }),
    });
  } catch {
    /* offline */
  }
}

Notifications.setNotificationHandler({
  handleNotification: async () => ({
    shouldShowAlert: true,
    shouldPlaySound: true,
    shouldSetBadge: true,
  }),
});

export function useAlertPushPolling(intervalMs = 60_000) {
  const [token, setToken] = useState<string | null>(null);
  const lastCount = useRef(0);

  useEffect(() => {
    void (async () => {
      if (Platform.OS === 'android') {
        await Notifications.setNotificationChannelAsync('alerts', {
          name: 'Alerts',
          importance: Notifications.AndroidImportance.MAX,
        });
      }
      const { status } = await Notifications.requestPermissionsAsync();
      if (status !== 'granted') return;
      const projectId = Constants.expoConfig?.extra?.eas?.projectId;
      const pushToken = await Notifications.getExpoPushTokenAsync(
        projectId ? { projectId } : undefined,
      );
      setToken(pushToken.data);
      await registerPushToken(pushToken.data);
    })();
  }, []);

  // Server-sent push via alerting → Expo API; polling kept as offline fallback only.
  useEffect(() => {
    if (!token) return;
    const poll = async () => {
      try {
        const alerts = await fetchAlerts();
        const firing = alerts.filter((a) => a.status === 'FIRING').length;
        if (firing > lastCount.current) {
          await Notifications.scheduleNotificationAsync({
            content: {
              title: 'NeuralOps alert',
              body: `${firing} firing alert(s) require attention`,
            },
            trigger: null,
          });
        }
        lastCount.current = firing;
      } catch {
        /* offline */
      }
    };
    void poll();
    const id = setInterval(poll, intervalMs * 3);
    return () => clearInterval(id);
  }, [intervalMs, token]);

  return token;
}
