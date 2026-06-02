import { useEffect, useMemo, useRef } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useParams } from '@tanstack/react-router';
import { Clock, ListOrdered, MonitorPlay } from 'lucide-react';
import rrwebPlayer from 'rrweb-player';
import 'rrweb-player/dist/style.css';
import { fetchSessionReplay } from '../api/observability';
import { EmptyState } from '../components/ui/EmptyState';
import { ErrorState, LoadingState } from '../components/ui/PageStates';
import { StitchPageShell } from '../components/stitch';

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

  const events = data ?? [];
  const hasRrweb = events.some((ev) => ev.type === 'rrweb');
  const hasEvents = events.length > 0;

  const range = useMemo(() => {
    if (events.length === 0) return null;
    const times = events.map((ev) => new Date(ev.recordedAt).getTime()).filter((t) => !Number.isNaN(t));
    if (times.length === 0) return null;
    const start = Math.min(...times);
    const end = Math.max(...times);
    return { start, durationSec: Math.round((end - start) / 1000) };
  }, [events]);

  return (
    <StitchPageShell
      title={`Session replay ${sessionId}`}
      subtitle={hasRrweb ? 'rrweb player' : 'Snapshot event timeline'}
      actions={<Link to="/rum">← RUM sessions</Link>}
    >
      {isLoading && <LoadingState />}
      {error && <ErrorState message="Replay not found" />}

      {!isLoading && !error && hasEvents && (
        <div className="replay-meta">
          <span className="replay-meta__item">
            <MonitorPlay size={14} aria-hidden /> {hasRrweb ? 'rrweb capture' : 'Snapshot capture'}
          </span>
          <span className="replay-meta__item">
            <ListOrdered size={14} aria-hidden /> {events.length} events
          </span>
          {range && (
            <span className="replay-meta__item">
              <Clock size={14} aria-hidden /> {range.durationSec}s · started {new Date(range.start).toLocaleTimeString()}
            </span>
          )}
        </div>
      )}

      {hasRrweb && <div ref={containerRef} className="replay-player" />}

      {!hasRrweb && !isLoading && hasEvents && (
        <div className="replay-timeline">
          {events.map((ev) => (
            <div key={ev.seq} className="replay-event">
              <span className="replay-event__type">
                <span className="muted">#{ev.seq}</span>
                <strong>{ev.type}</strong>
              </span>
              <span className="muted">{new Date(ev.recordedAt).toLocaleTimeString()}</span>
              {ev.type === 'snapshot' && typeof ev.payload?.html === 'string' && (
                <iframe title={`snapshot-${ev.seq}`} sandbox="" srcDoc={ev.payload.html} className="replay-snapshot" />
              )}
            </div>
          ))}
        </div>
      )}

      {!isLoading && !error && !hasEvents && (
        <EmptyState
          title="No replay events"
          description="This session has no recorded events. Enable the RUM SDK with rrweb capture to record replays."
          icon={<MonitorPlay size={32} />}
        />
      )}
    </StitchPageShell>
  );
}
