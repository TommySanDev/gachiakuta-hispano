import apiClient from '../lib/api-client';
import type { Character } from '../types/character';

// API methods related to characters
export async function getAllCharacters(): Promise<Character[]> {
  return await apiClient.get<Character[]>('/api/characters');
}

export async function getCharacterById(id: string): Promise<Character> {
  return await apiClient.get<Character>(`/api/characters/${id}`);
}

