import { useQuery } from '@tanstack/react-query';
import { fetchServiceHealth } from '../api/client';

const services = [
  { name: 'gateway', port: 8080 },
  { name: 'ingestion', port: 8081 },
  { name: 'analysis', port: 8082 },
  { name: 'correlation', port: 8083 },
  { name: 'incident', port: 8084 },
  { name: 'search', port: 8085 },
  { name: 'alerting', port: 8086 },
];

function ServiceHealthRow({ name, port }: { name: string; port: number }) {
  const { data, isLoading, isError } = useQuery({
    queryKey: ['health', name],
    queryFn: () => fetchServiceHealth(name, port),
    retry: false,
    refetchInterval: 30_000,
  });

  return (
    <tr>
      <td>{name}</td>
      <td>{port}</td>
      <td>
        {isLoading && 'Checking…'}
        {!isLoading && !isError && data?.healthy && <span className="badge ok">Healthy</span>}
        {!isLoading && (isError || !data?.healthy) && <span className="badge error">Unavailable</span>}
      </td>
    </tr>
  );
}

export function HealthPage() {
  return (
    <section className="panel">
      <h1>System Health</h1>
      <p className="muted">Live health checks for backend microservices.</p>
      <table className="health-table">
        <thead>
          <tr>
            <th>Service</th>
            <th>Port</th>
            <th>Status</th>
          </tr>
        </thead>
        <tbody>
          {services.map((service) => (
            <ServiceHealthRow key={service.name} name={service.name} port={service.port} />
          ))}
        </tbody>
      </table>
    </section>
  );
}
