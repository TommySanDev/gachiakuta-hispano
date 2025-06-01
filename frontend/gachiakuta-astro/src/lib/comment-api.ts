import apiClient from './api-client';
import type { Comment, CreateCommentInput, UpdateCommentInput } from '../types/comment';

// API methods following character-api.ts pattern
export async function getCommentsByChapter(chapterId: string): Promise<Comment[]> {
  return await apiClient.get<Comment[]>(`/api/comments/chapter/${chapterId}`);
}

export async function createComment(data: CreateCommentInput, token: string): Promise<Comment> {
  return await apiClient.post<Comment>('/api/comments', data, token);
}

export async function updateComment(id: string, data: UpdateCommentInput, token: string): Promise<Comment> {
  return await apiClient.put<Comment>(`/api/comments/${id}`, data, token);
}

export async function deleteComment(id: string, token: string): Promise<void> {
  await apiClient.delete(`/api/comments/${id}`, token);
}
