export interface User {
  id: number;
  username: string;
  email: string;
  created_at: string;
  updated_at: string;
  last_login_at?: string;
  status: number;
}

export interface AuthResponse {
  user: User;
  session_id: string;
}

export interface Task {
  id: number;
  user_id: number;
  task_type: string;
  sku?: string;
  keywords?: string;
  selling_points?: string;
  competitor_link?: string;
  status: string;
  result_data?: string;
  error_message?: string;
  created_at: string;
  updated_at: string;
  username?: string;
}

export interface TaskHistory {
  id: number;
  task_id: number;
  user_id: number;
  version: number;
  prompt?: string;
  aspect_ratio?: string;
  product_images_urls?: string;
  style_ref_image_url?: string;
  generated_image_url?: string;
  edit_instruction?: string;
  status: string;
  error_message?: string;
  created_at: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}
