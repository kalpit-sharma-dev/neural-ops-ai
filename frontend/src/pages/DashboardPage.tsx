import { usePlatformInfo } from '../hooks/usePlatformInfo';

export function DashboardPage() {
  const { data, isLoading, error } = usePlatformInfo();

  return (
    <section className="panel">
      <h1>Observability Command Center</h1>
      <p className="muted">
        Phase 0 scaffold is ready. Backend microservices, data stores, and the React UI are wired for local development.
      </p>

      <div className="card-grid">
        <article className="card">
          <h2>Gateway Status</h2>
          {isLoading && <p>Checking API gateway…</p>}
          {error && <p className="error">Gateway unreachable. Start docker-compose or run the gateway service locally.</p>}
          {data && (
            <dl>
              <dt>Service</dt>
              <dd>{data.service}</dd>
              <dt>Environment</dt>
              <dd>{data.environment}</dd>
              <dt>Version</dt>
              <dd>{data.version}</dd>
            </dl>
          )}
        </article>

        <article className="card">
          <h2>Next Phases</h2>
          <ul>
            <li>Phase 1 — Domain models</li>
            <li>Phase 2 — Log ingestion pipeline</li>
            <li>Phase 3 — AI analysis engine</li>
            <li>Phase 9 — Full React dashboard</li>
          </ul>
        </article>
      </div>
    </section>
  );
}
