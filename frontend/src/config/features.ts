/** Feature flags for phased Dynatrace parity rollout */
export const features = {
  alertingUI: true,
  apmTraces: true,
  metricsExplorer: true,
  customDashboards: true,
  liveTopology: true,
  adminUI: true,
  logSettings: true,
  slos: true,
  infra: true,
  databases: true,
  rum: true,
  synthetic: true,
  workflows: true,
  security: true,
  integrations: true,
} as const;
