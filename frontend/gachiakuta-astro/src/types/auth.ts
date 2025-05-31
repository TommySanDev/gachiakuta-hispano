// User roles matching backend constants
export type UserRole = 'admin' | 'editor' | 'user';

// User interface matching Go User struct
export interface User {
  id: number;
  email: string;
  username: string;
  first_name: string;
  last_name: string;
  role: UserRole;
  is_active: boolean;
  email_verified: boolean;
  magic_link_enabled: boolean;
  totp_enabled: boolean;
  last_login_at?: string;
  created_at: string;
}

// Authentication input types matching Go input structs
export interface LoginInput {
  email: string;
  password: string;
}

export interface RegisterInput {
  email: string;
  username: string;
  password: string;
  first_name: string;
  last_name: string;
  magic_link_enabled: boolean;
}

export interface UpdateProfileInput {
  username?: string;
  first_name?: string;
  last_name?: string;
  magic_link_enabled?: boolean;
}

export interface ChangePasswordInput {
  current_password: string;
  new_password: string;
}

// Magic link and password reset types
export interface MagicLinkLoginInput {
  email: string;
}

export interface ResetPasswordRequestInput {
  email: string;
}

export interface ResetPasswordInput {
  token: string;
  new_password: string;
}

// TOTP 2FA types
export interface VerifyTOTPInput {
  code: string;
}

export interface UseRecoveryCodeInput {
  code: string;
}

export interface TOTPSetupResponse {
  secret: string;
  qr_code: string; // base64 encoded
  backup_url: string;
}

// Response types matching Go response structs
export interface AuthResponse {
  user: User;
  token: string;
  expires_in: number;
}

// Authentication state for frontend store
export interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  token: string | null;
  isLoading: boolean;
  error: string | null;
}

// API error response type
export interface ApiError {
  error: string;
  message?: string;
}
