import { defineMiddleware } from 'astro:middleware';
import { getUserFromCookies, userHasRole } from './lib/auth-server';

// Define which routes require authentication and specific roles
const protectedRoutes = {
  // Routes that require any authentication
  auth: [
    '/auth/profile',
    '/auth/2fa-setup',
    '/favorites',
  ],
  // Routes that require user role or higher
  user: [
    '/auth/profile',
    '/auth/2fa-setup',
    '/favorites',
  ],
  // Routes that require editor role or higher
  editor: [
    '/editor',
    '/content-manager',
  ],
  // Routes that require admin role
  admin: [
    '/admin',
    '/admin/users',
    '/admin/content',
  ]
};

// Routes that should redirect to home if user is already authenticated
// Removed '/auth/reset-password' to allow logged-in users to reset their password
const guestOnlyRoutes = [
  '/auth/login',
  '/auth/register',
];

export const onRequest = defineMiddleware(async (context, next) => {
  const { url, cookies, redirect } = context;
  const pathname = url.pathname;

  // Get user from cookies (server-side validation)
  const user = await getUserFromCookies(cookies);

  // Check if user is authenticated
  const isAuthenticated = !!user;

  // Handle guest-only routes (login, register)
  if (guestOnlyRoutes.some(route => pathname.startsWith(route))) {
    if (isAuthenticated) {
      // User is already logged in, redirect to home
      return redirect('/');
    }
    // Continue to the guest-only page
    return next();
  }

  // Check authentication requirements
  const requiresAuth = protectedRoutes.auth.some(route => pathname.startsWith(route));

  if (requiresAuth && !isAuthenticated) {
    // User needs to be authenticated but isn't
    const returnTo = encodeURIComponent(pathname + url.search);
    return redirect(`/auth/login?returnTo=${returnTo}`);
  }

  // Check role-based access
  if (isAuthenticated && user) {
    // Check admin routes
    const requiresAdmin = protectedRoutes.admin.some(route => pathname.startsWith(route));
    if (requiresAdmin && !userHasRole(user, 'admin')) {
      return redirect('/auth/profile?error=insufficient-permissions');
    }

    // Check editor routes
    const requiresEditor = protectedRoutes.editor.some(route => pathname.startsWith(route));
    if (requiresEditor && !userHasRole(user, 'editor')) {
      return redirect('/auth/profile?error=insufficient-permissions');
    }

    // User routes are handled by basic authentication check above
  }

  // Continue to the requested page
  return next();
});

// Helper function to check if a path matches any pattern
function matchesPattern(pathname: string, patterns: string[]): boolean {
  return patterns.some(pattern => {
    if (pattern.endsWith('*')) {
      // Wildcard pattern
      return pathname.startsWith(pattern.slice(0, -1));
    }
    return pathname === pattern || pathname.startsWith(pattern + '/');
  });
}

// Export middleware configuration
export const config = {
  // Only run middleware on specific paths to improve performance
  matcher: [
    // Authentication pages
    '/auth/*',
    // Protected user areas
    '/favorites',
    '/profile',
    // Editor areas
    '/editor/*',
    '/content-manager/*',
    // Admin areas
    '/admin/*',
    // API routes (if any client-side API routes exist)
    '/api/*'
  ]
};
