import { authToken, updateUser } from '../auth-store';
import type { UpdateProfileInput, ChangePasswordInput, User } from '../../types/auth';

const API_BASE_URL = import.meta.env.PUBLIC_API_URL || 'http://localhost:8080';

// Profile update action
export async function updateProfile(data: UpdateProfileInput): Promise<User> {
  const token = authToken.value;
  if (!token) {
    throw new Error('Authentication required');
  }

  console.log('🔄 Updating profile:', data);

  const response = await fetch(`${API_BASE_URL}/api/users/me`, {
    method: 'PUT',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(data)
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || `HTTP ${response.status}`);
  }

  const updatedUser = await response.json();
  console.log('✅ Profile updated successfully:', updatedUser);

  // Update global auth state
  updateUser(updatedUser);

  return updatedUser;
}

// Password change action
export async function changePassword(data: ChangePasswordInput): Promise<void> {
  const token = authToken.value;
  if (!token) {
    throw new Error('Authentication required');
  }

  console.log('🔄 Changing password...');

  const response = await fetch(`${API_BASE_URL}/api/users/me/change-password`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(data)
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || `HTTP ${response.status}`);
  }

  console.log('✅ Password changed successfully');
}

// Magic link toggle action
export async function toggleMagicLink(enabled: boolean): Promise<User> {
  const token = authToken.value;
  if (!token) {
    throw new Error('Authentication required');
  }

  console.log('🔄 Toggling magic link:', enabled);

  const response = await fetch(`${API_BASE_URL}/api/users/me`, {
    method: 'PUT',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ magic_link_enabled: enabled })
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || `HTTP ${response.status}`);
  }

  const updatedUser = await response.json();
  console.log('✅ Magic link toggled successfully');

  // Update global auth state
  updateUser(updatedUser);

  return updatedUser;
}

// 2FA Actions
export async function generateRecoveryCodes(): Promise<string[]> {
  const token = authToken.value;
  if (!token) {
    throw new Error('Authentication required');
  }

  console.log('🔄 Generating recovery codes...');

  const response = await fetch(`${API_BASE_URL}/api/2fa/recovery-codes`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || `HTTP ${response.status}`);
  }

  const data = await response.json();
  console.log('✅ Recovery codes generated successfully');

  return data.recovery_codes || [];
}

export async function disable2FA(): Promise<void> {
  const token = authToken.value;
  if (!token) {
    throw new Error('Authentication required');
  }

  console.log('🔄 Disabling 2FA...');

  const response = await fetch(`${API_BASE_URL}/api/2fa/disable`, {
    method: 'DELETE',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || `HTTP ${response.status}`);
  }

  console.log('✅ 2FA disabled successfully');

  // Update user state
  updateUser({ totp_enabled: false });
}

// Utility function to download recovery codes
export function downloadRecoveryCodes(codes: string[]): void {
  const content = `Gachiakuta Hispano - Nuevos Códigos de Recuperación 2FA

⚠️  IMPORTANTE: Guarde estos códigos en un lugar seguro
🔑 Cada código se puede usar solo una vez
📱 Los necesitará si pierde acceso a su teléfono

Códigos de Recuperación:
${codes.map((code, index) => `${index + 1}. ${code}`).join('\n')}

Fecha de generación: ${new Date().toLocaleString()}

NOTA: Estos códigos reemplazan a los anteriores.`;

  const blob = new Blob([content], { type: 'text/plain' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = 'gachiakuta-recovery-codes-new.txt';
  a.style.display = 'none';
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}
