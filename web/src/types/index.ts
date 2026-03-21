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
  task_id?: string;
  user_id?: number;
  task_type?: string;
  type?: string;
  sku?: string;
  keywords?: string;
  selling_points?: string;
  competitor_link?: string;
  status?: string;
  task_status?: string;
  result_data?: string;
  error_message?: string;
  created_at?: string;
  updated_at?: string;
  username?: string;
  operator?: string;
  ctime?: number;
  mtime?: number;
  task_name?: string;
}

export interface TaskCenterBase {
  id: number;
  task_id: string;
  task_type: string;
  task_status: string;
  operator: string;
  ctime: number;
  mtime: number;
}

export interface CopywritingTaskDetail {
  id: number;
  task_id: string;
  competitor_urls: string;
  analysis_result: string;
  analyze_model: string;
  user_selected_data: string;
  product_details: string;
  generated_copy: string;
  generate_model: string;
  error_message: string;
  created_at: string;
  updated_at: string;
}

export interface ImageTaskDetail {
  id: number;
  task_id: string;
  sku: string;
  keywords: string;
  selling_points: string;
  competitor_link: string;
  copywriting_task_id: string;
  generate_model: string;
  aspect_ratio: string;
  result_data: string;
  generated_image_urls: string;
  error_message: string;
  created_at: string;
  updated_at: string;
}

export interface TaskCenterDetail {
  id: number;
  task_id: string;
  task_type: string;
  task_status: string;
  operator: string;
  ctime: number;
  mtime: number;
  detail_data: CopywritingTaskDetail | ImageTaskDetail;
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
  email: string;
  password_hash: string;
  verification_code: string;
}

export interface LoginRequest {
  login_id: string;
  password_hash: string;
}
