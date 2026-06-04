import { create } from 'zustand';

interface InvestigationState {
  service: string | null;
  traceId: string | null;
  txnId: string | null;
  syncFromSearch: (search: Record<string, unknown>) => void;
  setService: (service: string | null) => void;
  setTraceId: (traceId: string | null) => void;
  setTxnId: (txnId: string | null) => void;
  clearAll: () => void;
}

export const useInvestigationStore = create<InvestigationState>((set) => ({
  service: null,
  traceId: null,
  txnId: null,
  syncFromSearch: (search) =>
    set({
      service: typeof search.service === 'string' && search.service ? search.service : null,
      traceId: typeof search.traceId === 'string' && search.traceId ? search.traceId : null,
      txnId: typeof search.txnId === 'string' && search.txnId ? search.txnId : null,
    }),
  setService: (service) => set({ service }),
  setTraceId: (traceId) => set({ traceId }),
  setTxnId: (txnId) => set({ txnId }),
  clearAll: () => set({ service: null, traceId: null, txnId: null }),
}));
