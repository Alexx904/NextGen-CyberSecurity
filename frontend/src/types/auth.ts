export type UserRole = 'SuperAdmin' | 'TenantAdmin' | 'SecOps' | 'Auditor';

export interface User {
  id: string;
  tenant_id: string;
  email: string;
  full_name: string;
  role: UserRole;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_at: string;
}

export interface LoginResponseData {
  user: User;
  tokens: TokenPair;
}

export interface ApiError {
  code: string;
  message: string;
}

export interface ApiResponse<T> {
  success: boolean;
  data: T;
  error: ApiError | null;
}
