import { ReactNode, useEffect } from 'react';
import { useLocation } from '@tanstack/react-router';
import { Sidebar } from './layout/Sidebar';
import { TopBar } from './layout/TopBar';
import { DemoBanner } from './layout/DemoBanner';
import { TooltipProvider } from './ui/Tooltip';
import { CommandPalette } from './CommandPalette';
import { RUMConsentBanner } from './rum/RUMConsentBanner';
import { useRealtimeStore } from '../store/realtimeStore';

interface LayoutProps {
  children: ReactNode;
}

export function Layout({ children }: LayoutProps) {
  const location = useLocation();
  const connect = useRealtimeStore((s) => s.connect);
  const disconnect = useRealtimeStore((s) => s.disconnect);
  const focusMode = /^\/incidents\/[^/]+$/.test(location.pathname);

  useEffect(() => {
    connect();
    return () => disconnect();
  }, [connect, disconnect]);

  if (focusMode) {
    return (
      <TooltipProvider>
        <CommandPalette />
        <div className="app-shell app-shell--focus">
          <DemoBanner />
          <main className="app-content app-content--full">{children}</main>
        </div>
      </TooltipProvider>
    );
  }

  return (
    <TooltipProvider>
      <CommandPalette />
      <div className="app-shell">
        <DemoBanner />
        <RUMConsentBanner />
        <Sidebar />
        <div className="app-main-column">
          <TopBar />
          <main className="app-content">{children}</main>
        </div>
      </div>
    </TooltipProvider>
  );
}
