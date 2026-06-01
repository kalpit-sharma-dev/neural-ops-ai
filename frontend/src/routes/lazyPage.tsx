import { lazy, Suspense, type ComponentType } from 'react';
import { LoadingState } from '../components/ui/PageStates';

export function lazyPage(factory: () => Promise<{ default: ComponentType }>) {
  const Lazy = lazy(factory);
  return function LazyRoute() {
    return (
      <Suspense fallback={<LoadingState label="Loading page…" />}>
        <Lazy />
      </Suspense>
    );
  };
}
