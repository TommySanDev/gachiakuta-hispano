import type { Character } from "../types/character";

interface CharactersResponse {
  data: Character[];
  meta: {
    total: number;
    page: number;
    page_size: number;
    total_pages: number;
  };
}

const API_BASE = "/api/v1/characters";

// Fetch characters with filters and pagination
export async function getCharacters(params: Record<string, string | number> = {}): Promise<CharactersResponse> {
  const query = new URLSearchParams(params as Record<string, string>).toString();
  const res = await fetch(`${API_BASE}?${query}`);
  if (!res.ok) throw new Error("Failed to fetch characters");
  return res.json();
}

// Fetch character by ID
export async function getCharacterById(id: number): Promise<Character> {
  const res = await fetch(`${API_BASE}/${id}`);
  if (!res.ok) throw new Error("Failed to fetch character");
  return res.json();
}

// Fetch characters by status
export async function getCharactersByStatus(status: string, limit = 10): Promise<Character[]> {
  const res = await fetch(`${API_BASE}/status/${status}?limit=${limit}`);
  if (!res.ok) throw new Error("Failed to fetch characters by status");
  return res.json();
}

// Fetch characters by affiliation
export async function getCharactersByAffiliation(affiliation: string, limit = 10): Promise<Character[]> {
  const res = await fetch(`${API_BASE}/affiliation/${affiliation}?limit=${limit}`);
  if (!res.ok) throw new Error("Failed to fetch characters by affiliation");
  return res.json();
}

// Fetch characters by species
export async function getCharactersBySpecies(species: string, limit = 10): Promise<Character[]> {
  const res = await fetch(`${API_BASE}/species/${species}?limit=${limit}`);
  if (!res.ok) throw new Error("Failed to fetch characters by species");
  return res.json();
}

