export function DemoBanner() {
  if (import.meta.env.VITE_DEMO_MODE !== 'true') return null;

  return (
    <div className="demo-banner" role="status">
      <strong>DEMO MODE</strong>
      <span>Sample data and relaxed auth — not for production use.</span>
    </div>
  );
}
