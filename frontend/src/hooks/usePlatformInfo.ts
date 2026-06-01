import { useQuery } from '@tanstack/react-query';
import { fetchPlatformInfo } from '../api/client';

export function usePlatformInfo() {
  return useQuery({
    queryKey: ['platform-info'],
    queryFn: fetchPlatformInfo,
    retry: 1,
  });
}
