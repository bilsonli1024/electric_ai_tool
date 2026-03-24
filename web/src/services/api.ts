import { User, AuthResponse, LoginRequest, RegisterRequest, Task, TaskHistory } from '../types';

const API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:3002';

// 统一响应格式
interface StandardResponse<T = any> {
  code: number;
  message: string;
  data: T;
}

class ApiClient {
  private sessionId: string | null = null;

  constructor() {
    this.sessionId = localStorage.getItem('session_id');
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...options.headers,
    };

    if (this.sessionId) {
      headers['Authorization'] = `Bearer ${this.sessionId}`;
    }

    const response = await fetch(`${API_BASE}${endpoint}`, {
      ...options,
      headers,
    });

    // 解析响应
    const result: StandardResponse<T> = await response.json().catch(() => ({
      code: 500,
      message: 'Failed to parse response',
      data: null as T
    }));

    // 检查业务状态码
    if (result.code !== 0) {
      // 显示错误提示
      this.showError(result.message);
      throw new Error(result.message);
    }

    // 返回data部分
    return result.data;
  }

  // 显示错误提示（全局Toast）
  private showError(message: string) {
    // 触发自定义事件，由App组件监听并显示Toast
    window.dispatchEvent(new CustomEvent('api-error', {
      detail: { message }
    }));
  }

  // 便捷方法
  async get<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: 'GET' });
  }

  async post<T>(endpoint: string, data?: any): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  async put<T>(endpoint: string, data?: any): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  async delete<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: 'DELETE' });
  }

  async register(data: RegisterRequest): Promise<AuthResponse> {
    const response = await this.request<AuthResponse>('/api/auth/register', {
      method: 'POST',
      body: JSON.stringify(data),
    });
    this.sessionId = response.session_id;
    localStorage.setItem('session_id', response.session_id);
    return response;
  }

  async sendVerificationCode(email: string, purpose: string): Promise<{ message: string }> {
    return this.request('/api/auth/send-verification-code', {
      method: 'POST',
      body: JSON.stringify({ email, purpose }),
    });
  }

  async login(data: LoginRequest): Promise<AuthResponse> {
    // 直接使用传入的login_id和password_hash
    const payload = {
      login_id: data.login_id,
      password_hash: data.password_hash,
    };
    
    const response = await this.request<AuthResponse>('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
    this.sessionId = response.session_id;
    localStorage.setItem('session_id', response.session_id);
    return response;
  }

  async forgotPassword(email: string): Promise<{ message: string }> {
    return this.request('/api/auth/forgot-password', {
      method: 'POST',
      body: JSON.stringify({ email }),
    });
  }

  async resetPassword(token: string, newPasswordHash: string): Promise<{ message: string }> {
    return this.request('/api/auth/reset-password', {
      method: 'POST',
      body: JSON.stringify({ token, new_password_hash: newPasswordHash }),
    });
  }

  async logout(): Promise<void> {
    await this.request('/api/auth/logout', { method: 'POST' });
    this.sessionId = null;
    localStorage.removeItem('session_id');
  }

  async me(): Promise<User> {
    return this.request<User>('/api/auth/me');
  }

  async getTasks(params?: { limit?: number; offset?: number; type?: string }): Promise<{ data: Task[]; total: number }> {
    const queryParams = new URLSearchParams();
    if (params?.limit) queryParams.append('limit', params.limit.toString());
    if (params?.offset) queryParams.append('offset', params.offset.toString());
    if (params?.type) queryParams.append('type', params.type);

    return this.request<{ data: Task[]; total: number }>(
      `/api/tasks?${queryParams.toString()}`
    );
  }

  async getAllTasks(params?: { limit?: number; offset?: number }): Promise<{ data: Task[]; total: number }> {
    const queryParams = new URLSearchParams();
    if (params?.limit) queryParams.append('limit', params.limit.toString());
    if (params?.offset) queryParams.append('offset', params.offset.toString());

    return this.request<{ data: Task[]; total: number }>(
      `/api/tasks/all?${queryParams.toString()}`
    );
  }

  // 统一任务接口
  // 新任务中心API
  async getTaskCenterTasks(params?: {
    page_size?: number;
    page_no?: number;
    task_type?: string;
    task_status?: string;
    operator?: string;
    start_time?: number;
    end_time?: number;
    view_all?: boolean;
  }): Promise<{ data: any[]; total: number }> {
    const queryParams = new URLSearchParams();
    if (params?.page_size) queryParams.append('page_size', params.page_size.toString());
    if (params?.page_no) queryParams.append('page_no', params.page_no.toString());
    if (params?.task_type) queryParams.append('task_type', params.task_type);
    if (params?.task_status) queryParams.append('task_status', params.task_status);
    if (params?.operator) queryParams.append('operator', params.operator);
    if (params?.start_time) queryParams.append('start_time', params.start_time.toString());
    if (params?.end_time) queryParams.append('end_time', params.end_time.toString());
    if (params?.view_all) queryParams.append('view_all', 'true');

    return this.request<{ data: any[]; total: number }>(
      `/api/task-center/list?${queryParams.toString()}`
    );
  }

  async getTaskCenterDetail(taskId: string): Promise<{ data: any }> {
    return this.request(`/api/task-center/detail?task_id=${taskId}`);
  }

  async getTaskCenterStatistics(): Promise<{ data: any }> {
    return this.request('/api/task-center/statistics');
  }

  async copyTask(taskId: string): Promise<{ task_id: string; message: string }> {
    return this.request(`/api/task-center/copy?task_id=${taskId}`, {
      method: 'POST',
    });
  }

  // 旧的统一任务API（兼容）
  async getUnifiedTasks(params?: {
    limit?: number;
    offset?: number;
    task_type?: string;
    status?: number;
    start_time?: string;
    end_time?: string;
    view_all?: boolean;
  }): Promise<{ data: Task[]; total: number }> {
    // 转发到新API
    const newParams = {
      limit: params?.limit,
      offset: params?.offset,
      task_type: params?.task_type,
      task_status: params?.status !== undefined ? this.mapOldStatusToNew(params.status) : undefined,
      start_time: params?.start_time ? parseInt(params.start_time) : undefined,
      end_time: params?.end_time ? parseInt(params.end_time) : undefined,
      view_all: params?.view_all,
    };
    return this.getTaskCenterTasks(newParams);
  }

  private mapOldStatusToNew(oldStatus: number): string {
    // 旧状态映射到新状态
    switch (oldStatus) {
      case 0: return 'ongoing'; // 分析中
      case 1: return 'ongoing'; // 待生成
      case 2: return 'ongoing'; // 生成中
      case 3: return 'completed'; // 已完成
      case 10: return 'failed'; // 分析失败
      case 11: return 'failed'; // 生成失败
      default: return 'pending';
    }
  }

  async getUnifiedTaskStatistics(viewAll?: boolean): Promise<{ data: any }> {
    const queryParams = new URLSearchParams();
    if (viewAll) queryParams.append('view_all', 'true');
    
    return this.request<{ data: any }>(
      `/api/unified-tasks/statistics?${queryParams.toString()}`
    );
  }

  async getTaskHistory(taskId: number, params?: { limit?: number; offset?: number }): Promise<{ data: TaskHistory[]; total: number }> {
    const queryParams = new URLSearchParams({ task_id: taskId.toString() });
    if (params?.limit) queryParams.append('limit', params.limit.toString());
    if (params?.offset) queryParams.append('offset', params.offset.toString());

    return this.request<{ data: TaskHistory[]; total: number }>(
      `/api/tasks/history?${queryParams.toString()}`
    );
  }

  async analyzeWithTask(data: any): Promise<{ data: any; task_id: number }> {
    return this.request('/api/tasks/analyze', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async generateImageWithTask(data: any): Promise<{ data: string; task_id: number }> {
    return this.request('/api/tasks/generate-image', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async testModel(model: string, prompt: string): Promise<any> {
    return this.request('/api/models/test', {
      method: 'POST',
      body: JSON.stringify({ model, prompt }),
    });
  }

  async testAllModels(prompt: string): Promise<{ results: any[] }> {
    return this.request('/api/models/test-all', {
      method: 'POST',
      body: JSON.stringify({ prompt }),
    });
  }

  async analyzeCompetitors(urls: string[], model?: string, taskName?: string): Promise<{ data: any; task_id: number }> {
    // 将model名称转换为枚举值
    const modelMap: { [key: string]: number } = {
      'gemini': 1,
      'gpt': 2,
      'deepseek': 3
    };
    
    const modelValue = model ? (modelMap[model.toLowerCase()] || 1) : 1;
    
    return this.request('/api/copywriting/analyze', {
      method: 'POST',
      body: JSON.stringify({ urls, model: modelValue, task_name: taskName }),
    });
  }

  async generateCopy(data: any): Promise<{ data: any; task_id: number }> {
    // 将model名称转换为枚举值
    const modelMap: { [key: string]: number } = {
      'gemini': 1,
      'gpt': 2,
      'deepseek': 3
    };
    
    const requestData = { ...data };
    if (data.model && typeof data.model === 'string') {
      requestData.model = modelMap[data.model.toLowerCase()] || 1;
    }
    
    return this.request('/api/copywriting/generate', {
      method: 'POST',
      body: JSON.stringify(requestData),
    });
  }

  async getCopywritingTasks(params?: { limit?: number; offset?: number }): Promise<{ data: any[]; total: number }> {
    const queryParams = new URLSearchParams();
    if (params?.limit) queryParams.append('limit', params.limit.toString());
    if (params?.offset) queryParams.append('offset', params.offset.toString());

    return this.request<{ data: any[]; total: number }>(
      `/api/copywriting/tasks?${queryParams.toString()}`
    );
  }

  async generateImages(data: {
    sku: string;
    keywords: string;
    sellingPoints: string;
    competitorLink?: string;
    model: string;
    taskName?: string;
    copywritingTaskId?: string;
    productImages?: string[]; // 产品图片URL数组
  }): Promise<{ data: any; task_id: string }> {
    // 将model名称转换为枚举值
    const modelMap: { [key: string]: number } = {
      'gemini': 1,
      'gpt': 2,
      'deepseek': 3
    };
    
    const modelValue = modelMap[data.model.toLowerCase()] || 1;
    
    return this.request('/api/tasks/generate-image', {
      method: 'POST',
      body: JSON.stringify({
        ...data,
        model: modelValue
      }),
    });
  }

  async uploadImageBase64(image: string, filename?: string): Promise<{ url: string; path: string }> {
    return this.request('/api/upload/image-base64', {
      method: 'POST',
      body: JSON.stringify({ image, filename }),
    });
  }

  async getCopywritingTask(taskId: number): Promise<{ data: any }> {
    return this.request(`/api/copywriting/task?task_id=${taskId}`);
  }

  async searchCopywritingTasks(keyword: string, limit?: number): Promise<{ data: any[] }> {
    const queryParams = new URLSearchParams();
    queryParams.append('keyword', keyword);
    if (limit) queryParams.append('limit', limit.toString());

    return this.request<{ data: any[] }>(
      `/api/copywriting/search?${queryParams.toString()}`
    );
  }

  isAuthenticated(): boolean {
    return this.sessionId !== null;
  }
}

export const apiClient = new ApiClient();
