import { useEffect, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import { ArrowRight, Search } from 'lucide-react';
import { fetchTransaction } from '../api/search';
import { getApiErrorMessage } from '../api/client';
import type { Transaction } from '../api/types';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { EmptyState } from '../components/ui/EmptyState';
import { ErrorState, LoadingState } from '../components/ui/PageStates';
import { StitchPageShell } from '../components/stitch';

const RECENT_KEY = 'neuralops-recent-txn';

function hopClass(latencyMs: number, status: string) {
  if (status === 'FAILED') return 'txn-hop__box--failed';
  if (latencyMs > 500) return 'txn-hop__box--slow';
  return '';
}

function statusBadge(status: string) {
  if (status === 'SUCCESS') return 'healthy' as const;
  if (status === 'FAILED') return 'critical' as const;
  return 'warning' as const;
}

export default function TransactionJourney() {
  const [txnId, setTxnId] = useState('');
  const [searchId, setSearchId] = useState('');
  const [recent, setRecent] = useState<string[]>(() => {
    try {
      return JSON.parse(localStorage.getItem(RECENT_KEY) ?? '[]') as string[];
    } catch {
      return [];
    }
  });

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['transaction', searchId],
    queryFn: () => fetchTransaction(searchId),
    enabled: searchId.length > 0,
  });

  useEffect(() => {
    if (data?.txnId) {
      setRecent((prev) => {
        const next = [data.txnId, ...prev.filter((id) => id !== data.txnId)].slice(0, 8);
        localStorage.setItem(RECENT_KEY, JSON.stringify(next));
        return next;
      });
    }
  }, [data?.txnId]);

  const search = () => {
    if (txnId.trim()) setSearchId(txnId.trim());
  };

  return (
    <StitchPageShell title="Transaction Journey" subtitle="Trace every hop of a financial transaction">
      <Card className="ui-card" hover={false}>
        <div style={{ display: 'flex', gap: 12, maxWidth: 640, margin: '0 auto' }}>
          <input
            type="text"
            value={txnId}
            onChange={(e) => setTxnId(e.target.value)}
            placeholder="Enter Transaction ID, UPI Ref, NEFT UTR…"
            style={{
              flex: 1,
              padding: '12px 16px',
              background: 'var(--bg-elevated)',
              border: '1px solid var(--border-subtle)',
              borderRadius: 'var(--radius-md)',
              color: 'var(--text-primary)',
              fontSize: 16,
            }}
            onKeyDown={(e) => e.key === 'Enter' && search()}
          />
          <Button variant="primary" onClick={search}>
            <Search size={16} /> Trace
          </Button>
        </div>
        {recent.length > 0 && (
          <div style={{ marginTop: 16, textAlign: 'center' }}>
            <span className="muted">Recent: </span>
            {recent.map((id) => (
              <button
                key={id}
                type="button"
                className="pill"
                style={{ margin: '0 4px' }}
                onClick={() => {
                  setTxnId(id);
                  setSearchId(id);
                }}
              >
                {id.slice(0, 12)}…
              </button>
            ))}
          </div>
        )}
      </Card>

      {isLoading && <LoadingState label="Tracing transaction…" />}
      {error && searchId && (
        <ErrorState message={getApiErrorMessage(error)} onRetry={() => refetch()} />
      )}

      {data && <TransactionFlow txn={data} />}
      {!searchId && !isLoading && (
        <EmptyState
          title="Enter a transaction ID"
          description="Visualize hops across API Gateway, auth, UPI, ledger, and notification services."
        />
      )}
    </StitchPageShell>
  );
}

function TransactionFlow({ txn }: { txn: Transaction }) {
  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} style={{ marginTop: 24 }}>
      <div className="dashboard-kpis" style={{ gridTemplateColumns: 'repeat(4, 1fr)' }}>
        <Card title="Txn ID">
          <code style={{ fontFamily: 'var(--font-mono)', fontSize: 13 }}>{txn.txnId}</code>
        </Card>
        <Card title="Type">
          <Badge variant="info">{txn.txnType}</Badge>
        </Card>
        <Card title="Status">
          <Badge variant={statusBadge(txn.status)}>{txn.status}</Badge>
        </Card>
        <Card title="Total Latency">
          <strong>{txn.totalLatencyMs}ms</strong>
        </Card>
      </div>

      <Card title="Transaction Flow" hover={false}>
        <div className="txn-flow">
          {txn.hops.map((hop, idx) => (
            <div key={`${hop.serviceName}-${idx}`} style={{ display: 'flex', alignItems: 'center' }}>
              <div className="txn-hop">
                <div className={`txn-hop__box ${hopClass(hop.latencyMs, hop.status)}`}>
                  <strong>{hop.serviceName}</strong>
                  <div style={{ fontSize: 12, marginTop: 4 }}>{hop.latencyMs}ms</div>
                  <Badge variant={statusBadge(hop.status)}>{hop.status}</Badge>
                  {hop.errorMessage && (
                    <p style={{ fontSize: 11, color: 'var(--error)', marginTop: 8 }}>{hop.errorMessage}</p>
                  )}
                </div>
              </div>
              {idx < txn.hops.length - 1 && <ArrowRight className="txn-arrow" size={20} />}
            </div>
          ))}
        </div>
        {txn.failedAt && (
          <p className="error-text" style={{ marginTop: 16 }}>
            Failure point: {txn.failedAt}
          </p>
        )}
      </Card>
    </motion.div>
  );
}
