import { ReactNode, useEffect } from 'react';
import { useLocation } from '@tanstack/react-router';
import { Sidebar } from './layout/Sidebar';
import { TopBar } from './layout/TopBar';
import { DemoBanner } from './layout/DemoBanner';
import { TooltipProvider } from './ui/Tooltip';
import { CommandPalette } from './CommandPalette';
import { RUMConsentBanner } from './rum/RUMConsentBanner';
import { SkipLink } from './SkipLink';
import { useRealtimeStore } from '../store/realtimeStore';
import { InvestigationContextBar } from './layout/InvestigationContextBar';
import { PageNavigationBar } from './navigation/PageNavigationBar';
import { useGlobalAutoRefresh } from '../hooks/useGlobalAutoRefresh';
import { useInvestigationUrlSync } from '../hooks/useInvestigationUrlSync';
import { useFilterUrlSync } from '../hooks/useFilterUrlSync';

interface LayoutProps {
  children: ReactNode;
}

export function Layout({ children }: LayoutProps) {
  const location = useLocation();
  const connect = useRealtimeStore((s) => s.connect);
  const disconnect = useRealtimeStore((s) => s.disconnect);
  const focusMode = /^\/incidents\/[^/]+$/.test(location.pathname);
  useFilterUrlSync();
  useInvestigationUrlSync();
  useGlobalAutoRefresh();

  useEffect(() => {
    connect();
    return () => disconnect();
  }, [connect, disconnect]);

  if (focusMode) {
    return (
      <TooltipProvider>
        <SkipLink />
        <CommandPalette />
        <div className="app-shell app-shell--focus">
          <DemoBanner />
          <main id="main-content" className="app-content app-content--full" tabIndex={-1}>
            <PageNavigationBar />
            {children}
          </main>
        </div>
      </TooltipProvider>
    );
  }

  return (
    <TooltipProvider>
      <SkipLink />
      <CommandPalette />
      <RUMConsentBanner />
      <div className="app-shell">
        <Sidebar />
        <div className="app-main-column">
          <DemoBanner />
          <TopBar />
          <InvestigationContextBar />
          <main id="main-content" className="app-content" tabIndex={-1}>
            <PageNavigationBar />
            {children}
          </main>
        </div>
      </div>
    </TooltipProvider>
  );
}
