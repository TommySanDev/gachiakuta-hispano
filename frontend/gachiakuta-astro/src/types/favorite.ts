// Types for favorites functionality matching backend Favorite struct
export interface Favorite {
  id?: number;
  user_id?: number;
  entity_type: string; // 'character', 'chapter', 'vital_instrument'
  entity_id: number;
  created_at?: string;
}

// Input type for creating favorites
export interface CreateFavoriteInput {
  entity_type: string;
  entity_id: number;
}

// Input type for removing favorites
export interface RemoveFavoriteInput {
  entity_type: string;
  entity_id: number;
}

// Response type from backend - grouped by entity type
export interface FavoritesResponse {
  characters: Favorite[];
  chapters: Favorite[];
  vital_instruments: Favorite[];
}

// Helper type for favorite status
export interface FavoriteStatus {
  isFavorited: boolean;
  favoriteId?: number;
}
