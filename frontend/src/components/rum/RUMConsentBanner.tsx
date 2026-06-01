import { useState } from 'react';
import { recordRUMConsent } from '../../api/observability';

const CONSENT_KEY = 'neuralops_rum_consent';

export function RUMConsentBanner() {
  const [visible, setVisible] = useState(() => !localStorage.getItem(CONSENT_KEY));

  if (!visible) return null;

  const decide = async (given: boolean) => {
    localStorage.setItem(CONSENT_KEY, given ? 'granted' : 'denied');
    setVisible(false);
    const sessionId = `rum-${Date.now()}`;
    try {
      await recordRUMConsent({ sessionId, consentGiven: given, consentVersion: '1.0' });
    } catch {
      /* best effort */
    }
    if (given && typeof window !== 'undefined') {
      (window as Window & { __neuralopsRumConsent?: boolean }).__neuralopsRumConsent = true;
      window.dispatchEvent(new CustomEvent('neuralops-rum-consent', { detail: { granted: true } }));
    }
  };

  return (
    <div className="rum-consent-banner" role="dialog" aria-label="Privacy consent">
      <p>
        We use session replay and RUM to improve reliability. Personal data is masked per GDPR.{' '}
        <a href="/settings">Privacy policy</a>
      </p>
      <div style={{ display: 'flex', gap: 8 }}>
        <button type="button" className="ui-button ui-button--primary" onClick={() => void decide(true)}>
          Accept
        </button>
        <button type="button" className="ui-button ui-button--ghost" onClick={() => void decide(false)}>
          Decline
        </button>
      </div>
    </div>
  );
}

export function hasRUMConsent(): boolean {
  return localStorage.getItem(CONSENT_KEY) === 'granted';
}
