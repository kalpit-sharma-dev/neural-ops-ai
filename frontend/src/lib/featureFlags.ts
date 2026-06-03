/** Feature flags via Vite env — hide unfinished nav when false. */
const flags: Record<string, boolean> = {
  KUBERNETES: import.meta.env.VITE_FEATURE_KUBERNETES !== 'false',
  RUM: import.meta.env.VITE_FEATURE_RUM !== 'false',
  SYNTHETIC: import.meta.env.VITE_FEATURE_SYNTHETIC !== 'false',
  SERVICE_MAP: import.meta.env.VITE_FEATURE_SERVICE_MAP !== 'false',
  NOTEBOOKS: import.meta.env.VITE_FEATURE_NOTEBOOKS !== 'false',
  QUERY_WORKBENCH: import.meta.env.VITE_FEATURE_QUERY_WORKBENCH !== 'false',
  AI_OPS: import.meta.env.VITE_FEATURE_AI_OPS !== 'false',
  COLLECTORS_FLEET: import.meta.env.VITE_FEATURE_COLLECTORS_FLEET !== 'false',
  ENTERPRISE_GOVERNANCE: import.meta.env.VITE_FEATURE_ENTERPRISE_GOVERNANCE !== 'false',
  NFR_CERTIFICATION: import.meta.env.VITE_FEATURE_NFR_CERTIFICATION !== 'false',
  ALERT_POLICIES: import.meta.env.VITE_FEATURE_ALERT_POLICIES !== 'false',
  MATERIALIZATION: import.meta.env.VITE_FEATURE_MATERIALIZATION !== 'false',
  BUSINESS_OBSERVABILITY: import.meta.env.VITE_FEATURE_BUSINESS_OBSERVABILITY !== 'false',
  CLOUD_FINOPS: import.meta.env.VITE_FEATURE_CLOUD_FINOPS !== 'false',
};

export function isFeatureEnabled(key: string): boolean {
  return flags[key] ?? true;
}

export function filterNavByFeature<T extends { featureFlag?: string }>(items: T[]): T[] {
  return items.filter((item) => !item.featureFlag || isFeatureEnabled(item.featureFlag));
}
