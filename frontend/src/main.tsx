import React from 'react';
import ReactDOM from 'react-dom/client';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Toaster } from 'react-hot-toast';
import App from './App';
import { clearChunkReloadGuard } from './routes/lazyPage';
import { I18nProvider } from './i18n/I18nProvider';
import { initTheme } from './store/themeStore';
import './styles/global.css';

initTheme();
clearChunkReloadGuard();

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 1,
    },
  },
});

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <I18nProvider>
        <App />
      </I18nProvider>
      <Toaster
          position="top-right"
          toastOptions={{
            style: {
              background: 'var(--bg-elevated)',
              color: 'var(--text-primary)',
              border: '1px solid var(--border-subtle)',
            },
          }}
        />
    </QueryClientProvider>
  </React.StrictMode>,
);
