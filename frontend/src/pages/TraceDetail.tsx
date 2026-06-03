import { Link, useParams } from '@tanstack/react-router';
import { TraceDetailContent } from '../features/traces/TraceDetailContent';
import { StitchPageShell } from '../components/stitch';

export default function TraceDetail() {
  const { traceId = '' } = useParams({ strict: false });

  return (
    <StitchPageShell
      title={`Trace ${traceId}`}
      subtitle="PurePath-style trace analysis"
      actions={
        <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
          <Link to="/traces/settings">Sampling</Link>
          <Link to="/traces/compare" search={{ a: traceId, b: undefined }}>Compare</Link>
          <Link to="/traces">← Back</Link>
        </div>
      }
    >
      <TraceDetailContent traceId={traceId} />
    </StitchPageShell>
  );
}
