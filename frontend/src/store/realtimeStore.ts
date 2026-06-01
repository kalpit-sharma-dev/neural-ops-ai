import { create } from 'zustand';
import toast from 'react-hot-toast';
import type { RealtimeEvent } from '../api/types';

const WS_URL =
  import.meta.env.VITE_WS_URL ??
  (import.meta.env.DEV ? 'ws://localhost:8080/api/v1/ws/realtime' : `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.host}/api/v1/ws/realtime`);

interface RealtimeState {
  connected: boolean;
  events: RealtimeEvent[];
  unreadCount: number;
  connect: () => void;
  disconnect: () => void;
  clearUnread: () => void;
}

let socket: WebSocket | null = null;
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;

function handleEvent(event: RealtimeEvent, addEvent: (e: RealtimeEvent) => void) {
  addEvent(event);
  if (event.type === 'incident.created') {
    toast.error(event.message || 'New incident detected');
  } else if (event.type === 'incident.resolved') {
    toast.success(event.message || 'Incident resolved');
  } else if (event.type === 'anomaly.detected') {
    toast(`Anomaly: ${event.message}`, { icon: '⚡' });
  }
}

export const useRealtimeStore = create<RealtimeState>((set, get) => ({
  connected: false,
  events: [],
  unreadCount: 0,

  connect: () => {
    if (socket?.readyState === WebSocket.OPEN || socket?.readyState === WebSocket.CONNECTING) {
      return;
    }

    socket = new WebSocket(WS_URL);

    socket.onopen = () => {
      set({ connected: true });
      socket?.send(JSON.stringify({ services: [], severities: [] }));
    };

    socket.onmessage = (message) => {
      try {
        const event = JSON.parse(message.data as string) as RealtimeEvent;
        handleEvent(event, (e) => {
          set((state) => ({
            events: [e, ...state.events].slice(0, 100),
            unreadCount: state.unreadCount + (e.type !== 'heartbeat' ? 1 : 0),
          }));
        });
      } catch {
        /* ignore malformed payloads */
      }
    };

    socket.onclose = () => {
      set({ connected: false });
      reconnectTimer = setTimeout(() => get().connect(), 5000);
    };

    socket.onerror = () => {
      socket?.close();
    };
  },

  disconnect: () => {
    if (reconnectTimer) clearTimeout(reconnectTimer);
    socket?.close();
    socket = null;
    set({ connected: false });
  },

  clearUnread: () => set({ unreadCount: 0 }),
}));
