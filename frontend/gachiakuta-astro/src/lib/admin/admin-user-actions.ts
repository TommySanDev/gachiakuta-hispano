// Centralized admin actions and UI logic for user management
import { AdminApiClient } from './admin-api';

// --- Types ---
interface User {
  id: number;
  username: string;
  first_name: string;
  last_name: string;
  email: string;
  role: 'admin' | 'editor' | 'user' | string;
  is_active: boolean;
  email_verified: boolean;
  magic_link_enabled: boolean;
  totp_enabled: boolean;
  created_at?: string;
  last_login_at?: string;
  updated_at?: string;
}

interface UserUpdatePayload {
  first_name?: string;
  last_name?: string;
  email?: string;
  role?: string;
  is_active?: boolean;
}

/**
 * ADMIN USER ACTIONS
 * Functions for server-side operations like fetch, update, delete, etc.
 */
export const adminUserActions = {
  async getUser(userId: number): Promise<User> {
    return await AdminApiClient.get(`/users/${userId}`);
  },

  async updateUser(userId: number, data: UserUpdatePayload): Promise<User> {
    return await AdminApiClient.put(`/users/${userId}`, data);
  },

  async deleteUser(userId: number): Promise<void> {
    await AdminApiClient.delete(`/users/${userId}`);
  },

  async toggleUserStatus(userId: number, isActive: boolean): Promise<User> {
    return await AdminApiClient.put(`/users/${userId}`, { is_active: isActive });
  },

  async resetPassword(userId: number): Promise<{ temporary_password: string }> {
    return await AdminApiClient.post(`/users/${userId}/reset-password`);
  },

  async disable2FA(userId: number): Promise<void> {
    await AdminApiClient.delete(`/users/${userId}/disable-2fa`);
  },
};

/**
 * ADMIN UTILS
 * Functions for formatting and small helpers
 */
export const adminUtils = {
  formatDate(dateStr?: string): string {
    if (!dateStr) return 'Never';
    return new Date(dateStr).toLocaleString('en-US', {
      year: 'numeric', month: 'long', day: 'numeric',
      hour: '2-digit', minute: '2-digit'
    });
  },

  getRoleBadgeClasses(role: string): string {
    const map: Record<string, string> = {
      admin: 'bg-red-100 text-red-800',
      editor: 'bg-blue-100 text-blue-800',
      user: 'bg-gray-100 text-gray-800',
    };
    return map[role] || map.user;
  },

  getStatusBadgeClasses(isActive: boolean): string {
    return isActive ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800';
  },

  confirmDestructiveAction(action: string, target: string): boolean {
    return confirm(`Are you sure you want to ${action} "${target}"? This cannot be undone.`);
  },

  refreshCurrentPage(): void {
    if (typeof window.refreshUserData === 'function') {
      window.refreshUserData();
    } else {
      window.location.reload();
    }
  },
};

/**
 * UI POPULATION HELPERS
 * Populate DOM elements from user object for Astro components
 */
export const adminUserProfile = {
  populate(user: User): void {
    const setText = (id: string, value: string) => {
      const el = document.getElementById(id);
      if (el) el.textContent = value;
    };

    setText('user-initials', user.first_name[0] + user.last_name[0]);
    setText('user-name', `${user.first_name} ${user.last_name}`);
    setText('user-email', user.email);

    const roleBadge = document.getElementById('user-role-badge');
    if (roleBadge) {
      roleBadge.textContent = user.role;
      roleBadge.className = `inline-flex items-center px-3 py-1 rounded-full text-sm font-medium ${adminUtils.getRoleBadgeClasses(user.role)}`;
    }

    const statusBadge = document.getElementById('user-status-badge');
    if (statusBadge) {
      statusBadge.textContent = user.is_active ? 'Active' : 'Inactive';
      statusBadge.className = `inline-flex items-center px-3 py-1 rounded-full text-sm font-medium ${adminUtils.getStatusBadgeClasses(user.is_active)}`;
    }

    const toggleBtn = document.getElementById('toggle-status-btn');
    if (toggleBtn) {
      toggleBtn.textContent = user.is_active ? '❌ Deactivate Account' : '✅ Activate Account';
    }
  }
};

export const adminUserDetailsGrid = {
  populate(user: User): void {
    const safeText = (id: string, val: string) => {
      const el = document.getElementById(id);
      if (el) el.textContent = val;
    };

    safeText('detail-username', `@${user.username}`);
    safeText('detail-fullname', `${user.first_name} ${user.last_name}`);
    safeText('detail-email', user.email);
    safeText('detail-role', user.role);
    safeText('status-active', user.is_active ? '✅ Yes' : '❌ No');
    safeText('status-verified', user.email_verified ? '✅ Yes' : '⚠️ No');
    safeText('status-magic-link', user.magic_link_enabled ? '✅ Yes' : '❌ No');
    safeText('status-2fa', user.totp_enabled ? '✅ Yes' : '❌ No');
    safeText('detail-created', adminUtils.formatDate(user.created_at));
    safeText('detail-last-login', adminUtils.formatDate(user.last_login_at));
    safeText('detail-updated', adminUtils.formatDate(user.updated_at));

    const disable2FABtn = document.getElementById('disable-2fa-btn');
    if (disable2FABtn) {
      disable2FABtn.style.display = user.totp_enabled ? 'block' : 'none';
    }
  }
};

