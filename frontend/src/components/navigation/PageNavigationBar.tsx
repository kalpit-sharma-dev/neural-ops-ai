import { useMemo } from 'react';
import { Link, useNavigate, useRouter, useRouterState } from '@tanstack/react-router';
import { ArrowLeft, ChevronRight } from 'lucide-react';
import { buildRouteCrumbs, isSubScreen } from '../../lib/routeBreadcrumbs';
import { resolvePageId } from '../../i18n/routeKeys';
import { useI18n } from '../../i18n/I18nProvider';
import { pageTitle } from '../../i18n/messages';
import { Button } from '../ui/Button';

export function PageNavigationBar() {
  const router = useRouter();
  const navigate = useNavigate();
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const params = useRouterState({
    select: (s) => {
      const merged: Record<string, string> = {};
      for (const match of s.matches) {
        Object.assign(merged, match.params);
      }
      return merged;
    },
  });
  const { locale } = useI18n();

  const crumbs = useMemo(() => {
    const built = buildRouteCrumbs(pathname, params);
    const localized = built.map((crumb) => {
      const pageId = resolvePageId(crumb.path);
      const label = pageId ? pageTitle(locale, pageId, crumb.label) : crumb.label;
      return { ...crumb, label };
    });
    // Sub-screens: drop home; trail starts at the parent section (e.g. Log Explorer › Settings).
    return localized.length > 2 ? localized.slice(1) : localized;
  }, [pathname, params, locale]);

  if (!isSubScreen(pathname)) return null;

  const handleBack = () => {
    if (window.history.length > 1) {
      router.history.back();
      return;
    }
    const parent = crumbs.length > 1 ? crumbs[crumbs.length - 2] : crumbs[0];
    void navigate({ to: parent.path as '/' });
  };

  return (
    <nav className="page-navigation" aria-label="Page navigation">
      <Button
        type="button"
        variant="ghost"
        size="sm"
        className="page-navigation__back"
        onClick={handleBack}
        aria-label="Go back to previous screen"
      >
        <ArrowLeft size={16} aria-hidden />
        Back
      </Button>

      <ol className="page-navigation__breadcrumbs">
        {crumbs.map((crumb, index) => (
          <li key={crumb.path} className="page-navigation__crumb">
            {index > 0 && <ChevronRight size={14} className="page-navigation__sep" aria-hidden />}
            {crumb.isCurrent ? (
              <span className="page-navigation__current" aria-current="page">
                {crumb.label}
              </span>
            ) : (
              <Link to={crumb.path as '/'} className="page-navigation__link">
                {crumb.label}
              </Link>
            )}
          </li>
        ))}
      </ol>
    </nav>
  );
}
