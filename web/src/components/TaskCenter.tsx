import React, { useState, useEffect } from 'react';
import { Task } from '../types';
import { apiClient } from '../services/api';

export const TaskCenter: React.FC = () => {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [viewMode, setViewMode] = useState<'my' | 'all'>('my');
  const limit = 20;

  useEffect(() => {
    let cancelled = false;
    
    const loadTasks = async () => {
      setLoading(true);
      try {
        const offset = (page - 1) * limit;
        
        // 使用统一任务接口
        const response = await apiClient.getUnifiedTasks({
          limit,
          offset,
          view_all: viewMode === 'all'
        });
        
        if (!cancelled) {
          setTasks(response?.data || []);
          setTotal(response?.total || 0);
        }
      } catch (err) {
        console.error('Failed to load tasks:', err);
        if (!cancelled) {
          setTasks([]);
          setTotal(0);
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    };
    
    loadTasks();
    
    return () => {
      cancelled = true;
    };
  }, [page, viewMode]);

  const getStatusColor = (status: string | number) => {
    // 新架构使用字符串状态
    if (typeof status === 'string') {
      switch (status) {
        case 'completed':
          return 'bg-green-100 text-green-800';
        case 'ongoing':
          return 'bg-blue-100 text-blue-800';
        case 'pending':
          return 'bg-yellow-100 text-yellow-800';
        case 'failed':
          return 'bg-red-100 text-red-800';
        default:
          return 'bg-gray-100 text-gray-800';
      }
    }
    
    // 兼容旧架构的数字状态
    const statusNum = typeof status === 'string' ? parseInt(status) : status;
    switch (statusNum) {
      case 3: // 已完成
        return 'bg-green-100 text-green-800';
      case 0: // 分析中
      case 2: // 生成中
        return 'bg-blue-100 text-blue-800';
      case 1: // 分析完成，待生成
        return 'bg-yellow-100 text-yellow-800';
      case 10: // 分析失败
      case 11: // 生成失败
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const getStatusText = (status: string | number) => {
    // 新架构使用字符串状态
    if (typeof status === 'string') {
      switch (status) {
        case 'pending':
          return '待处理';
        case 'ongoing':
          return '进行中';
        case 'completed':
          return '已完成';
        case 'failed':
          return '失败';
        default:
          return status;
      }
    }
    
    // 兼容旧架构的数字状态
    const statusNum = typeof status === 'string' ? parseInt(status) : status;
    switch (statusNum) {
      case 0:
        return '分析中';
      case 1:
        return '待生成';
      case 2:
        return '生成中';
      case 3:
        return '已完成';
      case 10:
        return '分析失败';
      case 11:
        return '生成失败';
      default:
        return '未知';
    }
  };

  const getTaskTypeText = (task: any) => {
    // 优先使用task_type
    const taskType = task.task_type || task.type;
    
    if (taskType === 'copywriting') {
      // 根据状态显示文案任务的具体阶段
      if (task.status === 0) return '文案分析中';
      if (task.status === 1) return '文案待生成';
      if (task.status === 2) return '文案生成中';
      if (task.status === 3) return '文案已完成';
      if (task.status === 10) return '文案分析失败';
      if (task.status === 11) return '文案生成失败';
      return '文案生成';
    }
    
    switch (taskType) {
      case 'analyze':
        return '产品分析';
      case 'generate_image':
      case 'image':
        return '图片生成';
      case 'edit_image':
        return '图片编辑';
      case 'aplus_content':
        return 'A+内容';
      default:
        return taskType || '未知任务';
    }
  };

  const totalPages = Math.ceil(total / limit);

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-3xl font-bold text-gray-800">任务中心</h2>
        
        <div className="flex space-x-2">
          <button
            onClick={() => {
              setViewMode('my');
              setPage(1);
            }}
            className={`px-4 py-2 rounded-lg font-medium transition-colors ${
              viewMode === 'my'
                ? 'bg-indigo-600 text-white'
                : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
            }`}
          >
            我的任务
          </button>
          <button
            onClick={() => {
              setViewMode('all');
              setPage(1);
            }}
            className={`px-4 py-2 rounded-lg font-medium transition-colors ${
              viewMode === 'all'
                ? 'bg-indigo-600 text-white'
                : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
            }`}
          >
            全部任务
          </button>
        </div>
      </div>

      {loading ? (
        <div className="text-center py-12">
          <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600"></div>
          <p className="mt-4 text-gray-600">加载中...</p>
        </div>
      ) : !tasks || tasks.length === 0 ? (
        <div className="text-center py-12 bg-white rounded-lg shadow">
          <p className="text-gray-500">暂无任务</p>
        </div>
      ) : (
        <>
          <div className="bg-white rounded-lg shadow overflow-hidden">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    任务ID
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    类型
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    SKU/关键词
                  </th>
                  {viewMode === 'all' && (
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      创建者
                    </th>
                  )}
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    状态
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    创建时间
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {tasks.map((task) => (
                  <tr key={task.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                      #{task.id}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {getTaskTypeText(task)}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-500">
                      <div className="max-w-xs truncate">
                        {task.task_name || task.sku || task.keywords || '-'}
                      </div>
                    </td>
                    {viewMode === 'all' && (
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {task.username || '-'}
                      </td>
                    )}
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full ${getStatusColor(task.status)}`}>
                        {getStatusText(task.status)}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {new Date(task.created_at).toLocaleString('zh-CN')}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {totalPages > 1 && (
            <div className="mt-6 flex justify-center space-x-2">
              <button
                onClick={() => setPage(Math.max(1, page - 1))}
                disabled={page === 1}
                className="px-4 py-2 rounded-lg bg-white border border-gray-300 text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                上一页
              </button>
              <span className="px-4 py-2 text-gray-700">
                第 {page} / {totalPages} 页 (共 {total} 条)
              </span>
              <button
                onClick={() => setPage(Math.min(totalPages, page + 1))}
                disabled={page === totalPages}
                className="px-4 py-2 rounded-lg bg-white border border-gray-300 text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                下一页
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
};
