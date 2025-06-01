import type { AstroCookies } from 'astro';
import type { User } from '../types/auth';

const API_BASE_URL = import.meta.env.API_URL || 'http://localhost:8080';

// Server-side function to get user from cookies
export async function getUserFromCookies(cookies: AstroCookies): Promise<User | null> {
  try {
    const token = cookies.get('auth_token')?.value;

    if (!token) {
      return null;
    }

    // Validate token with backend
    const response = await fetch(`${API_BASE_URL}/api/users/me`, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });

    if (!response.ok) {
      // Token is invalid, clear cookies
      cookies.delete('auth_token', { path: '/' });
      cookies.delete('auth_user', { path: '/' });
      return null;
    }

    const user: User = await response.json();
    return user;

  } catch (error) {
    console.error('Failed to validate user token:', error);
    return null;
  }
}

// Server-side function to check if user has required role
export function userHasRole(user: User | null, requiredRole: string): boolean {
  if (!user) return false;

  switch (requiredRole) {
    case 'user':
      return true; // All authenticated users have user access
    case 'editor':
      return user.role === 'editor' || user.role === 'admin';
    case 'admin':
      return user.role === 'admin';
    default:
      return false;
  }
}

// Server-side function to set auth cookies
export function setAuthCookies(cookies: AstroCookies, user: User, token: string): void {
  const cookieOptions = {
    httpOnly: false, // We need client access for the store
    secure: import.meta.env.PROD, // HTTPS only in production
    sameSite: 'lax' as const,
    maxAge: 60 * 60 * 24 * 7, // 7 days
    path: '/'
  };

  cookies.set('auth_token', token, cookieOptions);
  cookies.set('auth_user', encodeURIComponent(JSON.stringify(user)), cookieOptions);
  //cookies.set('auth_user', JSON.stringify(user), cookieOptions);
}

// Server-side function to clear auth cookies
export function clearAuthCookies(cookies: AstroCookies): void {
  cookies.delete('auth_token', { path: '/' });
  cookies.delete('auth_user', { path: '/' });
}
