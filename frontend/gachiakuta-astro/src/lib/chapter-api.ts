import apiClient from '../lib/api-client';
import type { Chapter } from '../types/chapter';

// API methods related to chapters
export async function getAllChapters(): Promise<Chapter[]> {
  return await apiClient.get<Chapter[]>('/api/chapters');
}

export async function getChapterById(id: string): Promise<Chapter> {
  return await apiClient.get<Chapter>(`/api/chapters/${id}`);
}

