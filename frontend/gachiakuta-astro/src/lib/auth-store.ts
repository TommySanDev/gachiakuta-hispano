import { signal, computed } from '@preact/signals';
import type { User, AuthState } from '../types/auth';
import { apiClient } from './api-client';

// Signal-based store for reactive authentication state
const authState = signal<AuthState>({
  isAuthenticated: false,
  user: null,
  token: null,
  isLoading: false,
  error: null
});

// Computed values for convenient access
export const isAuthenticated = computed(() => authState.value.isAuthenticated);
export const currentUser = computed(() => authState.value.user);
export const authToken = computed(() => authState.value.token);
export const isLoading = computed(() => authState.value.isLoading);
export const authError = computed(() => authState.value.error);

// Helper to check user roles
export const hasRole = computed(() => (role: string) => {
  const user = currentUser.value;
  if (!user) return false;

  switch (role) {
    case 'user':
      return true; // All authenticated users have user access
    case 'editor':
      return user.role === 'editor' || user.role === 'admin';
    case 'admin':
      return user.role === 'admin';
    default:
      return false;
  }
});

// Cookie helper functions (client-side only)
function getCookie(name: string): string | null {
  if (typeof document === 'undefined') return null;

  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) {
    return parts.pop()?.split(';').shift() || null;
  }
  return null;
}

function setCookie(name: string, value: string, days: number = 7): void {
  if (typeof document === 'undefined') return;

  const expires = new Date();
  expires.setTime(expires.getTime() + (days * 24 * 60 * 60 * 1000));
  document.cookie = `${name}=${value};expires=${expires.toUTCString()};path=/;SameSite=Lax`;
}

function deleteCookie(name: string): void {
  if (typeof document === 'undefined') return;
  document.cookie = `${name}=;expires=Thu, 01 Jan 1970 00:00:00 UTC;path=/;`;
}

// Initialize auth state from cookies (client-side)
export function initializeAuth(): void {
  if (typeof window === 'undefined') return;

  try {
    const token = getCookie('auth_token');
    const userStr = getCookie('auth_user');

    if (token && userStr) {
      const user = JSON.parse(decodeURIComponent(userStr));
      authState.value = {
        isAuthenticated: true,
        user,
        token,
        isLoading: false,
        error: null
      };
    }
  } catch (error) {
    console.error('Failed to initialize auth from cookies:', error);
    clearAuth();
  }
}

// Set authenticated user and token
export function setAuth(user: User, token: string): void {
  authState.value = {
    isAuthenticated: true,
    user,
    token,
    isLoading: false,
    error: null
  };

  // Persist to cookies (client-side only, for compatibility)
  if (typeof window !== 'undefined') {
    setCookie('auth_token', token);
    setCookie('auth_user', encodeURIComponent(JSON.stringify(user)));
  }
}

// Clear authentication state
export function clearAuth(): void {
  authState.value = {
    isAuthenticated: false,
    user: null,
    token: null,
    isLoading: false,
    error: null
  };

  // Clear cookies
  if (typeof window !== 'undefined') {
    deleteCookie('auth_token');
    deleteCookie('auth_user');
  }
}

// Set loading state
export function setAuthLoading(loading: boolean): void {
  authState.value = {
    ...authState.value,
    isLoading: loading
  };
}

// Set auth error
export function setAuthError(error: string | null): void {
  authState.value = {
    ...authState.value,
    error,
    isLoading: false
  };
}

// Update current user data
export function updateUser(updates: Partial<User>): void {
  const currentState = authState.value;
  if (!currentState.user) return;

  const updatedUser = { ...currentState.user, ...updates };

  authState.value = {
    ...currentState,
    user: updatedUser
  };

  // Update cookies
  if (typeof window !== 'undefined') {
    setCookie('auth_user', encodeURIComponent(JSON.stringify(updatedUser)));
  }
}

// Refresh user data from server
export async function refreshUser(): Promise<void> {
  const token = authToken.value;
  if (!token) return;

  try {
    setAuthLoading(true);
    const user = await apiClient.get<User>('/api/users/me', token);
    updateUser(user);
    setAuthError(null);
  } catch (error) {
    console.error('Failed to refresh user data:', error);
    setAuthError('Failed to refresh user data');
  } finally {
    setAuthLoading(false);
  }
}

// Check if token is still valid
export async function validateToken(): Promise<boolean> {
  const token = authToken.value;
  if (!token) return false;

  try {
    await apiClient.get('/api/users/me', token);
    return true;
  } catch (error) {
    clearAuth();
    return false;
  }
}

// Export the store for debugging
export { authState };
