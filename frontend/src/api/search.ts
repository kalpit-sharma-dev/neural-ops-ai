import { getData, postData } from './client';
import type {
  AISearchResponse,
  DependencyMap,
  LogSearchRequest,
  SearchResponse,
  Transaction,
} from './types';

export function searchLogs(body: LogSearchRequest): Promise<SearchResponse> {
  return postData<SearchResponse>('/search/logs', body);
}

export function aiSearch(question: string, size = 50): Promise<AISearchResponse> {
  return postData<AISearchResponse>('/search/ai', { question, size });
}

export function fetchTransaction(txnId: string): Promise<Transaction> {
  return getData<Transaction>(`/search/txn/${encodeURIComponent(txnId)}`);
}

export function fetchDependencyMap(): Promise<DependencyMap> {
  return getData<DependencyMap>('/services/dependency-map');
}

export function searchTrace(traceId: string): Promise<SearchResponse> {
  return getData<SearchResponse>(`/search/trace/${encodeURIComponent(traceId)}`);
}
