import { useCallback, useEffect, useRef, useState } from 'react';
import { useAuthStore } from '../store/authStore';
import type { LogHit } from '../api/types';

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? '/api/v1';

function wsBase(): string {
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const host = window.location.host;
  return `${proto}//${host}${API_BASE}/logs/tail/ws`;
}

export function useLogTail(enabled: boolean, filters: { services: string[]; severities: string[]; query: string }) {
  const token = useAuthStore((s) => s.accessToken);
  const [lines, setLines] = useState<LogHit[]>([]);
  const [connected, setConnected] = useState(false);
  const socketRef = useRef<WebSocket | null>(null);
  const retryRef = useRef(0);
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const clear = useCallback(() => setLines([]), []);

  useEffect(() => {
    if (!enabled) {
      if (reconnectTimer.current) clearTimeout(reconnectTimer.current);
      socketRef.current?.close();
      socketRef.current = null;
      setConnected(false);
      retryRef.current = 0;
      return;
    }

    let cancelled = false;

    const connect = () => {
      if (cancelled) return;
      const url = new URL(wsBase());
      if (token) {
        url.searchParams.set('token', token);
      }
      const socket = new WebSocket(url.toString());
      socketRef.current = socket;

      socket.onopen = () => {
        retryRef.current = 0;
        setConnected(true);
        socket.send(
          JSON.stringify({
            services: filters.services,
            severities: filters.severities,
            query: filters.query,
          }),
        );
      };

      socket.onmessage = (ev) => {
      try {
        const row = JSON.parse(ev.data as string) as {
          timestamp: string;
          service: string;
          severity: string;
          message: string;
          traceId?: string;
          host?: string;
          pod?: string;
        };
        const hit: LogHit = {
          id: `${row.timestamp}-${row.service}-${Math.random().toString(36).slice(2, 8)}`,
          score: 1,
          timestamp: row.timestamp,
          service: row.service,
          severity: row.severity,
          message: row.message,
          traceId: row.traceId,
          host: row.host,
          pod: row.pod,
        };
        setLines((prev) => [hit, ...prev].slice(0, 500));
      } catch {
        /* ignore */
      }
    };

      socket.onclose = () => {
        setConnected(false);
        if (cancelled) return;
        const attempt = retryRef.current + 1;
        retryRef.current = attempt;
        const delay = Math.min(30_000, 500 * 2 ** attempt);
        reconnectTimer.current = setTimeout(connect, delay);
      };
      socket.onerror = () => socket.close();
    };

    connect();

    return () => {
      cancelled = true;
      if (reconnectTimer.current) clearTimeout(reconnectTimer.current);
      socketRef.current?.close();
      socketRef.current = null;
      setConnected(false);
    };
  }, [enabled, token, filters.services.join(','), filters.severities.join(','), filters.query]);

  useEffect(() => {
    const socket = socketRef.current;
    if (!socket || socket.readyState !== WebSocket.OPEN) return;
    socket.send(
      JSON.stringify({
        services: filters.services,
        severities: filters.severities,
        query: filters.query,
      }),
    );
  }, [filters.services, filters.severities, filters.query]);

  return { lines, connected, clear };
}
