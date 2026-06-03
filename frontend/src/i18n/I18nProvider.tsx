import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from 'react';
import { type LocaleCode, translate } from './messages';

const STORAGE_KEY = 'neuralops.locale';

interface I18nContextValue {
  locale: LocaleCode;
  setLocale: (code: LocaleCode) => void;
  t: (key: string) => string;
  /** Translate key with English fallback when missing. */
  tr: (key: string, fallback: string) => string;
}

const I18nContext = createContext<I18nContextValue | null>(null);

function readStoredLocale(): LocaleCode {
  const raw = localStorage.getItem(STORAGE_KEY);
  if (raw === 'es' || raw === 'de' || raw === 'en') return raw;
  return 'en';
}

export function I18nProvider({ children }: { children: ReactNode }) {
  const [locale, setLocaleState] = useState<LocaleCode>(readStoredLocale);

  useEffect(() => {
    document.documentElement.lang = locale;
  }, [locale]);

  const setLocale = useCallback((code: LocaleCode) => {
    localStorage.setItem(STORAGE_KEY, code);
    setLocaleState(code);
    document.documentElement.lang = code;
  }, []);

  const t = useCallback((key: string) => translate(locale, key), [locale]);
  const tr = useCallback(
    (key: string, fallback: string) => {
      const val = translate(locale, key);
      return val === key ? fallback : val;
    },
    [locale],
  );

  const value = useMemo(() => ({ locale, setLocale, t, tr }), [locale, setLocale, t, tr]);

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}

export function useI18n() {
  const ctx = useContext(I18nContext);
  if (!ctx) throw new Error('useI18n must be used within I18nProvider');
  return ctx;
}
