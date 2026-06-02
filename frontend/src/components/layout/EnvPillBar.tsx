import { useFilterStore, type Environment } from '../../store/filterStore';

const ENVIRONMENTS: Environment[] = ['ALL', 'PROD', 'STAGING', 'DEV'];

/**
 * Environment pill bar (ALL / PROD / STAGING / DEV) — segmented control that
 * mirrors the env tabs in the NeuralOps Command Center Stitch design.
 *
 * Backed by the global filter store so every page respects the selection.
 */
export function EnvPillBar({ className }: { className?: string }) {
  const environment = useFilterStore((s) => s.environment);
  const setEnvironment = useFilterStore((s) => s.setEnvironment);

  return (
    <div
      className={`env-pill-bar ${className ?? ''}`.trim()}
      role="radiogroup"
      aria-label="Environment"
    >
      {ENVIRONMENTS.map((env) => {
        const active = environment === env;
        return (
          <button
            key={env}
            type="button"
            role="radio"
            aria-checked={active}
            className={`env-pill ${active ? 'env-pill--active' : ''}`}
            onClick={() => setEnvironment(env)}
          >
            {env}
          </button>
        );
      })}
    </div>
  );
}
