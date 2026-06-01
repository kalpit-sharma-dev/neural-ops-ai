import { useEffect, useRef } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useParams } from '@tanstack/react-router';
import rrwebPlayer from 'rrweb-player';
import 'rrweb-player/dist/style.css';
import { fetchSessionReplay } from '../api/observability';
import { ErrorState, LoadingState, PageHeader } from '../components/ui/PageStates';

export default function SessionReplay() {
  const { sessionId = '' } = useParams({ strict: false });
  const containerRef = useRef<HTMLDivElement>(null);
  const playerRef = useRef<rrwebPlayer | null>(null);

  const { data, isLoading, error } = useQuery({
    queryKey: ['rum-replay', sessionId],
    queryFn: () => fetchSessionReplay(sessionId),
    enabled: sessionId.length > 0,
  });

  useEffect(() => {
    if (!data?.length || !containerRef.current) return;

    const rrwebEvents = data
      .filter((ev) => ev.type === 'rrweb' && ev.payload?.event)
      .map((ev) => ev.payload!.event as Record<string, unknown>);

    if (rrwebEvents.length === 0) return;

    playerRef.current?.destroy();
    containerRef.current.innerHTML = '';
    playerRef.current = new rrwebPlayer({
      target: containerRef.current,
      props: {
        events: rrwebEvents,
        width: Math.min(1024, window.innerWidth - 80),
        height: 560,
        autoPlay: false,
        showController: true,
      },
    });

    return () => {
      playerRef.current?.destroy();
      playerRef.current = null;
    };
  }, [data]);

  const hasRrweb = (data ?? []).some((ev) => ev.type === 'rrweb');

  return (
    <div>
      <PageHeader
        title={`Session replay ${sessionId}`}
        subtitle={hasRrweb ? 'rrweb player' : 'Legacy snapshot events'}
        actions={<Link to="/rum">← RUM sessions</Link>}
      />
      {isLoading && <LoadingState />}
      {error && <ErrorState message="Replay not found" />}
      {hasRrweb && <div ref={containerRef} className="replay-player" />}
      {!hasRrweb && !isLoading && (
        <div className="replay-timeline">
          {(data ?? []).map((ev) => (
            <div key={ev.seq} className="replay-event">
              <span className="muted">#{ev.seq}</span>
              <strong>{ev.type}</strong>
              <span className="muted">{new Date(ev.recordedAt).toLocaleTimeString()}</span>
              {ev.type === 'snapshot' && typeof ev.payload?.html === 'string' && (
                <iframe title={`snapshot-${ev.seq}`} sandbox="" srcDoc={ev.payload.html} className="replay-snapshot" />
              )}
            </div>
          ))}
          {(data ?? []).length === 0 && <p className="muted">No replay events — enable RUM SDK with rrweb.</p>}
        </div>
      )}
    </div>
  );
}
