/** Feature flags via Vite env — hide unfinished nav when false. */
const flags: Record<string, boolean> = {
  KUBERNETES: import.meta.env.VITE_FEATURE_KUBERNETES !== 'false',
  RUM: import.meta.env.VITE_FEATURE_RUM !== 'false',
  SYNTHETIC: import.meta.env.VITE_FEATURE_SYNTHETIC !== 'false',
  SERVICE_MAP: import.meta.env.VITE_FEATURE_SERVICE_MAP !== 'false',
  NOTEBOOKS: import.meta.env.VITE_FEATURE_NOTEBOOKS !== 'false',
};

export function isFeatureEnabled(key: string): boolean {
  return flags[key] ?? true;
}

export function filterNavByFeature<T extends { featureFlag?: string }>(items: T[]): T[] {
  return items.filter((item) => !item.featureFlag || isFeatureEnabled(item.featureFlag));
}
