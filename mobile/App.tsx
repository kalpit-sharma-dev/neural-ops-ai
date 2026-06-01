import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { NavigationContainer, DefaultTheme } from '@react-navigation/native';
import { StatusBar } from 'expo-status-bar';
import { StyleSheet, Text, View, FlatList, ActivityIndicator } from 'react-native';
import { useEffect, useState } from 'react';
import { fetchAlerts, fetchIncidents, fetchUsage, type AlertRecord, type IncidentSummary } from './src/api';
import { useAlertPushPolling } from './src/push';

const Tab = createBottomTabNavigator();

const theme = {
  ...DefaultTheme,
  colors: {
    ...DefaultTheme.colors,
    background: '#0f172a',
    card: '#1e293b',
    text: '#f8fafc',
    border: '#334155',
    primary: '#38bdf8',
  },
};

function DashboardScreen() {
  const [usage, setUsage] = useState<{ logsIngestedGb: number; tracesIngested: number; activeUsers: number } | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchUsage()
      .then(setUsage)
      .catch(() => setUsage(null))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <ActivityIndicator style={styles.loader} color="#38bdf8" />;

  return (
    <View style={styles.screen}>
      <Text style={styles.title}>NeuralOps Mobile</Text>
      <View style={styles.card}>
        <Text style={styles.metricLabel}>Logs ingested</Text>
        <Text style={styles.metricValue}>{usage?.logsIngestedGb?.toFixed(1) ?? '—'} GB</Text>
      </View>
      <View style={styles.card}>
        <Text style={styles.metricLabel}>Traces</Text>
        <Text style={styles.metricValue}>{usage?.tracesIngested ?? '—'}</Text>
      </View>
      <View style={styles.card}>
        <Text style={styles.metricLabel}>Active users</Text>
        <Text style={styles.metricValue}>{usage?.activeUsers ?? '—'}</Text>
      </View>
    </View>
  );
}

function AlertsScreen() {
  const [alerts, setAlerts] = useState<AlertRecord[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchAlerts()
      .then(setAlerts)
      .catch(() => setAlerts([]))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <ActivityIndicator style={styles.loader} color="#38bdf8" />;

  return (
    <FlatList
      style={styles.screen}
      data={alerts}
      keyExtractor={(item) => item.id}
      ListHeaderComponent={<Text style={styles.title}>Alerts</Text>}
      renderItem={({ item }) => (
        <View style={styles.card}>
          <Text style={styles.itemTitle}>{item.title}</Text>
          <Text style={styles.muted}>{item.severity} · {item.service} · {item.status}</Text>
        </View>
      )}
      ListEmptyComponent={<Text style={styles.muted}>No alerts</Text>}
    />
  );
}

function IncidentsScreen() {
  const [incidents, setIncidents] = useState<IncidentSummary[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchIncidents()
      .then(setIncidents)
      .catch(() => setIncidents([]))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <ActivityIndicator style={styles.loader} color="#38bdf8" />;

  return (
    <FlatList
      style={styles.screen}
      data={incidents}
      keyExtractor={(item) => item.id}
      ListHeaderComponent={<Text style={styles.title}>Incidents</Text>}
      renderItem={({ item }) => (
        <View style={styles.card}>
          <Text style={styles.itemTitle}>{item.title}</Text>
          <Text style={styles.muted}>{item.severity} · {item.service}</Text>
        </View>
      )}
      ListEmptyComponent={<Text style={styles.muted}>No incidents</Text>}
    />
  );
}

export default function App() {
  useAlertPushPolling();
  return (
    <NavigationContainer theme={theme}>
      <StatusBar style="light" />
      <Tab.Navigator screenOptions={{ headerShown: false, tabBarStyle: { backgroundColor: '#1e293b' } }}>
        <Tab.Screen name="Dashboard" component={DashboardScreen} />
        <Tab.Screen name="Alerts" component={AlertsScreen} />
        <Tab.Screen name="Incidents" component={IncidentsScreen} />
      </Tab.Navigator>
    </NavigationContainer>
  );
}

const styles = StyleSheet.create({
  screen: { flex: 1, backgroundColor: '#0f172a', padding: 16 },
  loader: { flex: 1, marginTop: 40 },
  title: { color: '#f8fafc', fontSize: 22, fontWeight: '700', marginBottom: 16 },
  card: {
    backgroundColor: '#1e293b',
    borderRadius: 12,
    padding: 16,
    marginBottom: 12,
    borderWidth: 1,
    borderColor: '#334155',
  },
  metricLabel: { color: '#94a3b8', fontSize: 13 },
  metricValue: { color: '#38bdf8', fontSize: 28, fontWeight: '700', marginTop: 4 },
  itemTitle: { color: '#f8fafc', fontSize: 16, fontWeight: '600' },
  muted: { color: '#94a3b8', fontSize: 13, marginTop: 4 },
});
