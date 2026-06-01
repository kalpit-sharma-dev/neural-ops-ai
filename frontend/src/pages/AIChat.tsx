import { useRef, useState } from 'react';
import { Cpu, Send, Square } from 'lucide-react';
import { streamChatQuery, type ChatSource } from '../api/chat';
import { MarkdownMessage } from '../components/chat/MarkdownMessage';
import { Button } from '../components/ui/Button';
import { PageHeader } from '../components/ui/PageStates';

interface Message {
  role: 'user' | 'assistant';
  content: string;
  streaming?: boolean;
  sources?: ChatSource[];
}

const SUGGESTED = [
  { category: 'Incidents', question: 'What caused the payment failures last night?' },
  { category: 'Performance', question: 'Which service has the worst latency this week?' },
  { category: 'Trends', question: 'Is our error rate improving over the past month?' },
  { category: 'Banking', question: 'Show me all failed UPI transactions today' },
  { category: 'Deployments', question: 'Did any deployments cause incidents this week?' },
];

export default function AIChat() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [streaming, setStreaming] = useState(false);
  const [status, setStatus] = useState('');
  const abortRef = useRef<AbortController | null>(null);

  const send = async (question: string) => {
    if (!question.trim() || streaming) return;

    setInput('');
    setStreaming(true);
    setStatus('Searching logs…');

    let assistantIdx = 0;
    setMessages((m) => {
      assistantIdx = m.length + 1;
      return [
        ...m,
        { role: 'user', content: question },
        { role: 'assistant', content: '', streaming: true },
      ];
    });

    abortRef.current = new AbortController();
    let buffer = '';

    setTimeout(() => setStatus('Analyzing patterns…'), 800);
    setTimeout(() => setStatus('Generating response…'), 1600);

    await streamChatQuery({
      question,
      signal: abortRef.current.signal,
      onSources: (sources) => {
        setMessages((m) => m.map((msg, i) => (i === assistantIdx ? { ...msg, sources } : msg)));
      },
      onChunk: (chunk) => {
        buffer += chunk;
        setMessages((m) =>
          m.map((msg, i) => (i === assistantIdx ? { ...msg, content: buffer } : msg)),
        );
      },
      onError: (errorMsg) => {
        setMessages((m) =>
          m.map((item, i) =>
            i === assistantIdx ? { ...item, content: `Error: ${errorMsg}`, streaming: false } : item,
          ),
        );
        setStreaming(false);
        setStatus('');
      },
      onDone: () => {
        setMessages((m) => m.map((msg, i) => (i === assistantIdx ? { ...msg, streaming: false } : msg)));
        setStreaming(false);
        setStatus('');
      },
    });
  };

  const stop = () => {
    abortRef.current?.abort();
    setStreaming(false);
    setStatus('');
  };

  return (
    <div>
      <PageHeader title="AI Assistant" subtitle="Senior SRE colleague powered by NeuralOps" />

      <div className="chat-layout">
        <aside className="chat-sidebar">
          <h3 style={{ marginTop: 0, fontFamily: 'var(--font-display)' }}>Suggested Questions</h3>
          {SUGGESTED.map(({ category, question }) => (
            <button
              key={question}
              type="button"
              className="pill"
              style={{ display: 'block', width: '100%', textAlign: 'left', marginBottom: 8 }}
              onClick={() => send(question)}
            >
              <strong>{category}</strong>
              <br />
              <span style={{ fontSize: 12 }}>{question}</span>
            </button>
          ))}
        </aside>

        <section className="chat-main">
          <div className="chat-messages">
            {messages.length === 0 && (
              <p className="muted" style={{ textAlign: 'center', marginTop: 48 }}>
                Ask anything about your logs, incidents, or deployments
              </p>
            )}
            {messages.map((msg, i) => (
              <div
                key={i}
                className={`chat-bubble chat-bubble--${msg.role === 'user' ? 'user' : 'ai'}`}
              >
                {msg.role === 'assistant' && (
                  <Cpu size={16} style={{ marginBottom: 8, color: 'var(--accent-primary)' }} />
                )}
                {msg.role === 'assistant' ? (
                  <>
                    <MarkdownMessage content={msg.content || (msg.streaming ? '…' : '')} />
                    {(msg.sources?.length ?? 0) > 0 && (
                      <details className="chat-sources">
                        <summary>Sources ({msg.sources!.length})</summary>
                        <ul>
                          {msg.sources!.map((source, idx) => (
                            <li key={`${source.service}-${idx}`}>
                              <strong>{source.service}</strong> · {source.severity}
                              <p className="muted">{source.message}</p>
                            </li>
                          ))}
                        </ul>
                      </details>
                    )}
                  </>
                ) : (
                  msg.content
                )}
              </div>
            ))}
            {status && <p className="muted">{status}</p>}
          </div>

          <div className="chat-input-area">
            <textarea
              value={input}
              onChange={(e) => setInput(e.target.value)}
              placeholder="Ask about incidents, logs, performance… (Ctrl+Enter to send)"
              onKeyDown={(e) => {
                if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
                  e.preventDefault();
                  void send(input);
                }
              }}
            />
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
              {streaming ? (
                <Button variant="danger" onClick={stop}>
                  <Square size={14} /> Stop
                </Button>
              ) : (
                <Button variant="primary" onClick={() => send(input)}>
                  <Send size={14} /> Send
                </Button>
              )}
            </div>
          </div>
        </section>
      </div>
    </div>
  );
}
