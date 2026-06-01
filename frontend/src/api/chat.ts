import { API_BASE_URL } from './client';
import { useAuthStore } from '../store/authStore';

export interface ChatSource {
  service: string;
  message: string;
  severity: string;
  timestamp: string;
}

export interface ChatStreamOptions {
  question: string;
  /** Optional page context for route-aware answers */
  context?: string;
  signal?: AbortSignal;
  onChunk: (chunk: string) => void;
  onSources?: (sources: ChatSource[]) => void;
  onDone?: () => void;
  onError?: (message: string) => void;
}

export async function streamChatQuery(options: ChatStreamOptions): Promise<void> {
  const { question, context, signal, onChunk, onSources, onDone, onError } = options;
  const token = useAuthStore.getState().accessToken;
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (token) headers.Authorization = `Bearer ${token}`;

  const response = await fetch(`${API_BASE_URL}/chat/query`, {
    method: 'POST',
    headers,
    body: JSON.stringify({ question, ...(context ? { context } : {}) }),
    signal,
  });

  if (!response.ok) {
    onError?.(`Chat request failed (${response.status})`);
    return;
  }

  const reader = response.body?.getReader();
  if (!reader) {
    onError?.('Streaming not supported');
    return;
  }

  const decoder = new TextDecoder();
  let buffer = '';

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;

    buffer += decoder.decode(value, { stream: true });
    const parts = buffer.split('\n\n');
    buffer = parts.pop() ?? '';

    for (const part of parts) {
      const lines = part.split('\n');
      let eventType = 'message';
      let data = '';

      for (const line of lines) {
        if (line.startsWith('event:')) eventType = line.slice(6).trim();
        if (line.startsWith('data:')) data += line.slice(5).trim();
      }

      if (eventType === 'error') {
        onError?.(data);
        return;
      }
      if (eventType === 'sources' && data) {
        try {
          onSources?.(JSON.parse(data) as ChatSource[]);
        } catch {
          /* ignore malformed sources */
        }
        continue;
      }
      if (eventType === 'done' || data === '[DONE]') {
        onDone?.();
        return;
      }
      if (data) onChunk(data);
    }
  }

  onDone?.();
}
