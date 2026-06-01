import { getData, postData } from './client';

export interface AuthConfig {
  authEnabled: boolean;
  oidcEnabled: boolean;
  devLoginEnabled: boolean;
  defaultTenant: string;
}

export interface AuthUserProfile {
  user: {
    id: string;
    email: string;
    name: string;
    role: string;
  };
  tenantId: string;
  tenantName: string;
  plan?: string;
}

export interface TokenResponse {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
  user?: AuthUserProfile['user'];
  tenantId?: string;
  tenantName?: string;
}

export function fetchAuthConfig(): Promise<AuthConfig> {
  return getData<AuthConfig>('/auth/config');
}

export function fetchMe(): Promise<AuthUserProfile> {
  return getData<AuthUserProfile>('/auth/me');
}

export function startOIDCLogin(): Promise<{ authorizationUrl: string; state: string }> {
  return postData<{ authorizationUrl: string; state: string }>('/auth/oidc/start');
}

export function devLogin(email: string, tenantId?: string): Promise<TokenResponse> {
  return postData<TokenResponse>('/auth/dev/login', { email, tenantId });
}

export function refreshSession(refreshToken: string): Promise<TokenResponse> {
  return postData<TokenResponse>('/auth/refresh', { refreshToken });
}

export function exchangeOIDCCode(code: string): Promise<TokenResponse> {
  return postData<TokenResponse>('/auth/oidc/exchange', { code });
}

export function logoutSession(refreshToken?: string | null): Promise<{ loggedOut: boolean }> {
  return postData<{ loggedOut: boolean }>('/auth/logout', {
    refreshToken: refreshToken ?? undefined,
  });
}
