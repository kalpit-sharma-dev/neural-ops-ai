import { Globe } from 'lucide-react';
import { useI18n } from '../../i18n/I18nProvider';
import type { LocaleCode } from '../../i18n/messages';

const LOCALES: { code: LocaleCode; label: string }[] = [
  { code: 'en', label: 'EN' },
  { code: 'es', label: 'ES' },
  { code: 'de', label: 'DE' },
];

export function LocaleSwitcher() {
  const { locale, setLocale, t } = useI18n();

  return (
    <label className="topbar__locale" data-testid="locale-switcher">
      <Globe size={14} aria-hidden />
      <span className="sr-only">{t('common.locale')}</span>
      <select
        value={locale}
        onChange={(e) => setLocale(e.target.value as LocaleCode)}
        aria-label={t('common.locale')}
      >
        {LOCALES.map((l) => (
          <option key={l.code} value={l.code}>
            {l.label}
          </option>
        ))}
      </select>
    </label>
  );
}
