import { getData, postData } from './client';
import type { Incident, IncidentTimelineResponse, Recommendation } from './types';

export interface IncidentListParams {
  status?: string;
  severity?: string;
  service?: string;
  page?: number;
  size?: number;
}

export function fetchIncidents(params?: IncidentListParams): Promise<Incident[]> {
  return getData<Incident[]>('/incidents', params as Record<string, unknown>);
}

export function fetchIncident(id: string): Promise<Incident> {
  return getData<Incident>(`/incidents/${id}`);
}

export function fetchIncidentTimeline(id: string): Promise<IncidentTimelineResponse> {
  return getData<IncidentTimelineResponse>(`/incidents/${id}/timeline`);
}

export function fetchRecommendations(id: string): Promise<Recommendation[]> {
  return getData<Recommendation[]>(`/incidents/${id}/recommendations`);
}

export function acknowledgeIncident(id: string): Promise<Incident> {
  return postData<Incident>(`/incidents/${id}/acknowledge`);
}

export function resolveIncident(id: string, notes?: string): Promise<Incident> {
  return postData<Incident>(`/incidents/${id}/resolve`, { notes });
}
