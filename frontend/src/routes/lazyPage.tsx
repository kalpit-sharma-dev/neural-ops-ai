import { Component, lazy, Suspense, type ComponentType, type ReactNode } from 'react';
import { ErrorState, LoadingState } from '../components/ui/PageStates';

const CHUNK_RELOAD_KEY = 'neuralops:chunk-reload';

function isChunkLoadError(err: unknown): boolean {
  const msg = err instanceof Error ? err.message : String(err);
  return (
    msg.includes('Failed to fetch dynamically imported module') ||
    msg.includes('Importing a module script failed') ||
    msg.includes('error loading dynamically imported module') ||
    msg.includes('Loading chunk') ||
    msg.includes('Loading CSS chunk')
  );
}

function lazyWithChunkRecovery(factory: () => Promise<{ default: ComponentType }>) {
  return lazy(() =>
    factory().catch((err: unknown) => {
      if (isChunkLoadError(err) && !sessionStorage.getItem(CHUNK_RELOAD_KEY)) {
        sessionStorage.setItem(CHUNK_RELOAD_KEY, '1');
        window.location.reload();
        return new Promise<{ default: ComponentType }>(() => {});
      }
      throw err;
    }),
  );
}

interface ChunkErrorBoundaryProps {
  children: ReactNode;
}

interface ChunkErrorBoundaryState {
  error: Error | null;
}

class ChunkErrorBoundary extends Component<ChunkErrorBoundaryProps, ChunkErrorBoundaryState> {
  state: ChunkErrorBoundaryState = { error: null };

  static getDerivedStateFromError(error: Error): ChunkErrorBoundaryState {
    return { error };
  }

  render() {
    if (this.state.error) {
      const chunkError = isChunkLoadError(this.state.error);
      return (
        <ErrorState
          message={
            chunkError
              ? 'A newer version of NeuralOps is available. Reload to load this page.'
              : this.state.error.message
          }
          onRetry={() => {
            sessionStorage.removeItem(CHUNK_RELOAD_KEY);
            window.location.reload();
          }}
        />
      );
    }
    return this.props.children;
  }
}

export function lazyPage(factory: () => Promise<{ default: ComponentType }>) {
  const Lazy = lazyWithChunkRecovery(factory);
  return function LazyRoute() {
    return (
      <ChunkErrorBoundary>
        <Suspense fallback={<LoadingState label="Loading page…" />}>
          <Lazy />
        </Suspense>
      </ChunkErrorBoundary>
    );
  };
}

/** Clears one-shot chunk reload guard after a successful app boot. */
export function clearChunkReloadGuard() {
  sessionStorage.removeItem(CHUNK_RELOAD_KEY);
}
