import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';
import type { ApiError, ApiResponse } from './types';
import { useAuthStore } from '../store/authStore';

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '/api/v1';

export const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30_000,
  headers: { 'Content-Type': 'application/json' },
});

let refreshPromise: Promise<string | null> | null = null;

async function refreshAccessToken(): Promise<string | null> {
  const { refreshToken, setTokens, logout } = useAuthStore.getState();
  if (!refreshToken) {
    logout();
    return null;
  }

  try {
    const response = await axios.post<ApiResponse<{ accessToken: string; refreshToken?: string }>>(
      `${API_BASE_URL}/auth/refresh`,
      { refreshToken },
    );
    const { accessToken, refreshToken: nextRefresh } = response.data.data;
    setTokens(accessToken, nextRefresh ?? refreshToken);
    return accessToken;
  } catch {
    logout();
    return null;
  }
}

apiClient.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = useAuthStore.getState().accessToken;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  const tenantId = useAuthStore.getState().tenantId;
  if (tenantId) {
    config.headers['X-Tenant-ID'] = tenantId;
  }
  return config;
});

apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError<ApiError>) => {
    const original = error.config as InternalAxiosRequestConfig & { _retry?: boolean };
    if (error.response?.status === 401 && original && !original._retry) {
      original._retry = true;
      refreshPromise ??= refreshAccessToken().finally(() => {
        refreshPromise = null;
      });
      const token = await refreshPromise;
      if (token) {
        original.headers.Authorization = `Bearer ${token}`;
        return apiClient(original);
      }
    }

    const normalized: ApiError = error.response?.data ?? {
      status: 'error',
      errorCode: 'NET001',
      message: error.message || 'Network request failed',
      timestamp: new Date().toISOString(),
    };
    return Promise.reject(normalized);
  },
);

export async function getData<T>(url: string, params?: Record<string, unknown>): Promise<T> {
  const response = await apiClient.get<ApiResponse<T>>(url, { params });
  return response.data.data;
}

export async function postData<T>(url: string, body?: unknown): Promise<T> {
  const response = await apiClient.post<ApiResponse<T>>(url, body);
  return response.data.data;
}

export async function putData<T>(url: string, body?: unknown): Promise<T> {
  const response = await apiClient.put<ApiResponse<T>>(url, body);
  return response.data.data;
}

export async function deleteData<T>(url: string): Promise<T> {
  const response = await apiClient.delete<ApiResponse<T>>(url);
  return response.data.data;
}

export interface PlatformInfo {
  service: string;
  environment: string;
  version: string;
}

export interface HealthStatus {
  healthy: boolean;
}

export async function fetchPlatformInfo(): Promise<PlatformInfo> {
  const response = await apiClient.get<ApiResponse<PlatformInfo>>('/info');
  return response.data.data;
}

export async function fetchServiceHealth(serviceName: string, port: number): Promise<HealthStatus> {
  const devBaseUrl = import.meta.env.DEV
    ? `http://localhost:${port}`
    : `http://${serviceName}:${port}`;

  const response = await axios.get<ApiResponse<HealthStatus>>(`${devBaseUrl}/health`, {
    timeout: 3000,
  });
  return response.data.data;
}

export function getApiErrorMessage(error: unknown): string {
  if (typeof error === 'object' && error !== null && 'message' in error) {
    return String((error as ApiError).message);
  }
  return 'An unexpected error occurred';
}
