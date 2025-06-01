import { signal, computed } from '@preact/signals';
import type { FavoritesResponse, CreateFavoriteInput, RemoveFavoriteInput, FavoriteStatus } from '../types/favorite';
import { getUserFavorites, addFavorite, removeFavorite } from './favorite-api';
import { authToken, isAuthenticated } from './auth-store';

// State interface for favorites
interface FavoritesState {
  data: FavoritesResponse;
  isLoading: boolean;
  error: string | null;
  isInitialized: boolean;
}

// Signal-based store for favorites
const favoritesState = signal<FavoritesState>({
  data: {
    characters: [],
    chapters: [],
    vital_instruments: []
  },
  isLoading: false,
  error: null,
  isInitialized: false
});

// Computed values for convenient access
export const favoritesData = computed(() => favoritesState.value.data);
export const favoriteCharacters = computed(() => favoritesState.value.data.characters);
export const favoriteChapters = computed(() => favoritesState.value.data.chapters);
export const favoriteVitalInstruments = computed(() => favoritesState.value.data.vital_instruments);
export const favoritesLoading = computed(() => favoritesState.value.isLoading);
export const favoritesError = computed(() => favoritesState.value.error);
export const favoritesInitialized = computed(() => favoritesState.value.isInitialized);

// Helper computed to check if specific entity is favorited
export const isFavorited = computed(() => (entityType: string, entityId: number): FavoriteStatus => {
  const favorites = favoritesState.value.data;
  let targetArray: typeof favorites.characters = [];

  switch (entityType) {
    case 'character':
      targetArray = favorites.characters;
      break;
    case 'chapter':
      targetArray = favorites.chapters;
      break;
    case 'vital_instrument':
      targetArray = favorites.vital_instruments;
      break;
    default:
      return { isFavorited: false };
  }

  const favorite = targetArray.find(fav => fav.entity_id === entityId);
  return {
    isFavorited: !!favorite,
    favoriteId: favorite?.id
  };
});

// Set loading state
export function setFavoritesLoading(loading: boolean): void {
  favoritesState.value = {
    ...favoritesState.value,
    isLoading: loading
  };
}

// Set error state
export function setFavoritesError(error: string | null): void {
  favoritesState.value = {
    ...favoritesState.value,
    error,
    isLoading: false
  };
}

// Set favorites data
export function setFavoritesData(data: FavoritesResponse): void {
  favoritesState.value = {
    ...favoritesState.value,
    data,
    isLoading: false,
    error: null,
    isInitialized: true
  };
}

// Clear favorites (on logout)
export function clearFavorites(): void {
  favoritesState.value = {
    data: {
      characters: [],
      chapters: [],
      vital_instruments: []
    },
    isLoading: false,
    error: null,
    isInitialized: false
  };
}

// Load user favorites from server
export async function loadFavorites(): Promise<void> {
  const token = authToken.value;
  if (!token || !isAuthenticated.value) {
    clearFavorites();
    return;
  }

  try {
    setFavoritesLoading(true);
    setFavoritesError(null);

    const data = await getUserFavorites(token);
    setFavoritesData(data);

  } catch (error) {
    console.error('Failed to load favorites:', error);
    setFavoritesError(error instanceof Error ? error.message : 'Failed to load favorites');
  }
}

// Add entity to favorites
export async function addToFavorites(entityType: string, entityId: number): Promise<void> {
  const token = authToken.value;
  if (!token) {
    throw new Error('Authentication required');
  }

  const input: CreateFavoriteInput = { entity_type: entityType, entity_id: entityId };

  try {
    const newFavorite = await addFavorite(input, token);

    // Update local state optimistically
    const currentData = favoritesState.value.data;
    const updatedData = { ...currentData };

    switch (entityType) {
      case 'character':
        updatedData.characters = [...currentData.characters, newFavorite];
        break;
      case 'chapter':
        updatedData.chapters = [...currentData.chapters, newFavorite];
        break;
      case 'vital_instrument':
        updatedData.vital_instruments = [...currentData.vital_instruments, newFavorite];
        break;
    }

    setFavoritesData(updatedData);

  } catch (error) {
    console.error('Failed to add favorite:', error);
    throw error;
  }
}

// Remove entity from favorites
export async function removeFromFavorites(entityType: string, entityId: number): Promise<void> {
  const token = authToken.value;
  if (!token) {
    throw new Error('Authentication required');
  }

  const input: RemoveFavoriteInput = { entity_type: entityType, entity_id: entityId };

  try {
    await removeFavorite(input, token);

    // Update local state optimistically
    const currentData = favoritesState.value.data;
    const updatedData = { ...currentData };

    switch (entityType) {
      case 'character':
        updatedData.characters = currentData.characters.filter(fav => fav.entity_id !== entityId);
        break;
      case 'chapter':
        updatedData.chapters = currentData.chapters.filter(fav => fav.entity_id !== entityId);
        break;
      case 'vital_instrument':
        updatedData.vital_instruments = currentData.vital_instruments.filter(fav => fav.entity_id !== entityId);
        break;
    }

    setFavoritesData(updatedData);

  } catch (error) {
    console.error('Failed to remove favorite:', error);
    throw error;
  }
}

// Toggle favorite status
export async function toggleFavorite(entityType: string, entityId: number): Promise<void> {
  const status = isFavorited.value(entityType, entityId);

  if (status.isFavorited) {
    await removeFromFavorites(entityType, entityId);
  } else {
    await addToFavorites(entityType, entityId);
  }
}

// Initialize favorites store
export function initializeFavorites(): void {
  // Only load if user is authenticated
  if (isAuthenticated.value) {
    loadFavorites();
  }
}

// Export the store for debugging
export { favoritesState };
