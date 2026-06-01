/**
 * @neuralops/mobile-sdk — native RUM + push for iOS/Android (Expo/React Native).
 * Publish to npm / embed via Expo config plugin for App Store / Play Store apps.
 */
export { trackScreen, trackAction, trackError, recordRUMEvent, getRUMSessionId } from '../src/rum';
export { useAlertPushPolling } from '../src/push';

export type MobileSDKConfig = {
  apiBase: string;
  tenantId: string;
  easProjectId?: string;
};

export function configureMobileSDK(config: MobileSDKConfig) {
  return config;
}
