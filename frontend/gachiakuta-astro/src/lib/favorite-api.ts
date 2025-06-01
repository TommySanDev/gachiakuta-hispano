import apiClient from './api-client';
import type {
  FavoritesResponse,
  CreateFavoriteInput,
  RemoveFavoriteInput,
  Favorite
} from '../types/favorite';

// API methods for favorites following the character-api.ts pattern
export async function getUserFavorites(token: string): Promise<FavoritesResponse> {
  return await apiClient.get<FavoritesResponse>('/api/favorites', token);
}

export async function addFavorite(data: CreateFavoriteInput, token: string): Promise<Favorite> {
  return await apiClient.post<Favorite>('/api/favorites', data, token);
}

export async function removeFavorite(data: RemoveFavoriteInput, token: string): Promise<void> {
  // Custom DELETE request with body since apiClient.delete doesn't support body
  const API_BASE_URL = import.meta.env.PUBLIC_API_URL || 'http://localhost:8080';

  const response = await fetch(`${API_BASE_URL}/api/favorites`, {
    method: 'DELETE',
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
}
