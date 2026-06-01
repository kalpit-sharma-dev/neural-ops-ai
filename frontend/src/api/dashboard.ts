import { getData } from './client';
import type { DashboardOverview } from './types';

export function fetchDashboardOverview(): Promise<DashboardOverview> {
  return getData<DashboardOverview>('/dashboard/overview');
}
