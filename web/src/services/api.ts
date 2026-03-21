import { User, AuthResponse, LoginRequest, RegisterRequest, Task, TaskHistory } from '../types';

const API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:3002';

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

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Request failed' }));
      throw new Error(error.error || 'Request failed');
    }

    return response.json();
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
    const response = await this.request<AuthResponse>('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify(data),
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
  async getUnifiedTasks(params?: { 
    limit?: number; 
    offset?: number; 
    task_type?: string;
    status?: number;
    start_time?: string;
    end_time?: string;
    view_all?: boolean;
  }): Promise<{ data: Task[]; total: number }> {
    const queryParams = new URLSearchParams();
    if (params?.limit) queryParams.append('limit', params.limit.toString());
    if (params?.offset) queryParams.append('offset', params.offset.toString());
    if (params?.task_type) queryParams.append('task_type', params.task_type);
    if (params?.status !== undefined) queryParams.append('status', params.status.toString());
    if (params?.start_time) queryParams.append('start_time', params.start_time);
    if (params?.end_time) queryParams.append('end_time', params.end_time);
    if (params?.view_all) queryParams.append('view_all', 'true');

    return this.request<{ data: Task[]; total: number }>(
      `/api/unified-tasks?${queryParams.toString()}`
    );
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
    return this.request('/api/copywriting/analyze', {
      method: 'POST',
      body: JSON.stringify({ urls, model: model || 'gemini', task_name: taskName }),
    });
  }

  async generateCopy(data: any): Promise<{ data: any; task_id: number }> {
    return this.request('/api/copywriting/generate', {
      method: 'POST',
      body: JSON.stringify(data),
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
    copywritingTaskId?: number;
  }): Promise<{ data: any; task_id: number }> {
    return this.request('/api/tasks/generate-image', {
      method: 'POST',
      body: JSON.stringify(data),
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
