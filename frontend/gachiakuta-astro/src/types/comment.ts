export interface Comment {
  id: number;
  user_id: number;
  chapter_id: number;
  content: string;
  created_at: string;
  // From CommentWithUser struct
  username: string;
  first_name: string;
  last_name: string;
}

export interface CreateCommentInput {
  chapter_id: number;
  content: string;
}

export interface UpdateCommentInput {
  content: string;
}
