import React, { useState, useEffect } from 'react';
import { Task } from '../types';
import { apiClient } from '../services/api';
import { useNavigate } from 'react-router-dom';
import { Calendar, User } from 'lucide-react';

export const TaskCenter: React.FC = () => {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [pageNo, setPageNo] = useState(1);
  const [viewMode, setViewMode] = useState<'my' | 'all'>('my');
  const pageSize = 20;
  const navigate = useNavigate();

  // 筛选器状态
  const [filterOperator, setFilterOperator] = useState('');
  const [filterStartTime, setFilterStartTime] = useState('');
  const [filterEndTime, setFilterEndTime] = useState('');

  // 格式化时间戳（秒级转为可读格式）
  const formatTimestamp = (timestamp: number | string | undefined) => {
    if (!timestamp) return '-';
    
    // 如果是字符串且为ISO格式（旧格式）
    if (typeof timestamp === 'string') {
      const date = new Date(timestamp);
      if (!isNaN(date.getTime())) {
        return date.toLocaleString('zh-CN');
      }
      return timestamp;
    }
    
    // 如果是秒级时间戳（新格式）
    const date = new Date(timestamp * 1000);
    return date.toLocaleString('zh-CN');
  };

  // 查看任务详情
  const handleViewDetail = (task: any) => {
    const taskType = task.task_type || task.type;
    const taskId = task.task_id || task.id;
    
    if (taskType === 'copywriting') {
      navigate(`/copywriting?task_id=${taskId}`);
    } else if (taskType === 'image') {
      navigate(`/image-generation?task_id=${taskId}`);
    }
  };

  // 获取任务名称/SKU
  const getTaskName = (task: any) => {
    const taskType = task.task_type || task.type;
    if (taskType === 'copywriting') {
      // 文案生成任务显示任务名称
      return task.task_name || '-';
    } else if (taskType === 'image') {
      // 图片生成任务显示SKU
      return task.sku || '-';
    }
    return task.task_name || task.sku || task.keywords || '-';
  };

  useEffect(() => {
    let cancelled = false;
    
    const loadTasks = async () => {
      setLoading(true);
      try {
        // 转换时间筛选为秒级时间戳
        let startTimestamp: number | undefined;
        let endTimestamp: number | undefined;
        
        if (filterStartTime) {
          startTimestamp = Math.floor(new Date(filterStartTime).getTime() / 1000);
        }
        if (filterEndTime) {
          // 结束时间设为当天的23:59:59
          const endDate = new Date(filterEndTime);
          endDate.setHours(23, 59, 59, 999);
          endTimestamp = Math.floor(endDate.getTime() / 1000);
        }
        
        // 使用新的任务中心接口
        const response = await apiClient.getTaskCenterTasks({
          page_size: pageSize,
          page_no: pageNo,
          view_all: viewMode === 'all',
          operator: filterOperator || undefined,
          start_time: startTimestamp,
          end_time: endTimestamp,
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
  }, [pageNo, viewMode, filterOperator, filterStartTime, filterEndTime]);

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
    
    // 旧架构使用数字状态（保持兼容）
    switch (status) {
      case 0:
      case 2:
        return 'bg-blue-100 text-blue-800';
      case 1:
        return 'bg-yellow-100 text-yellow-800';
      case 3:
        return 'bg-green-100 text-green-800';
      case 10:
      case 11:
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
    
    // 旧架构使用数字状态（保持兼容）
    switch (status) {
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
    const taskType = task.task_type || task.type;
    
    if (taskType === 'copywriting') {
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

  const totalPages = Math.ceil(total / pageSize);

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-3xl font-bold text-gray-800">任务中心</h2>
        
        <div className="flex space-x-2">
          <button
            onClick={() => {
              setViewMode('my');
              setPageNo(1);
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
              setPageNo(1);
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

      {/* 筛选器 */}
      <div className="bg-white rounded-lg shadow p-4 mb-6">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {/* 创建者筛选 */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2 flex items-center gap-2">
              <User size={16} />
              创建者
            </label>
            <input
              type="text"
              placeholder="输入邮箱或用户名"
              value={filterOperator}
              onChange={(e) => {
                setFilterOperator(e.target.value);
                setPageNo(1);
              }}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none"
            />
          </div>

          {/* 开始时间 */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2 flex items-center gap-2">
              <Calendar size={16} />
              开始时间
            </label>
            <input
              type="date"
              value={filterStartTime}
              onChange={(e) => {
                setFilterStartTime(e.target.value);
                setPageNo(1);
              }}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none"
            />
          </div>

          {/* 结束时间 */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2 flex items-center gap-2">
              <Calendar size={16} />
              结束时间
            </label>
            <input
              type="date"
              value={filterEndTime}
              onChange={(e) => {
                setFilterEndTime(e.target.value);
                setPageNo(1);
              }}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none"
            />
          </div>
        </div>

        {/* 清空筛选器按钮 */}
        {(filterOperator || filterStartTime || filterEndTime) && (
          <div className="mt-3 flex justify-end">
            <button
              onClick={() => {
                setFilterOperator('');
                setFilterStartTime('');
                setFilterEndTime('');
                setPageNo(1);
              }}
              className="px-4 py-2 text-sm text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-lg transition-colors"
            >
              清空筛选
            </button>
          </div>
        )}
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
                    任务名称/SKU
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    创建者
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    状态
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    创建时间
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                    操作
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {tasks.map((task) => (
                  <tr key={task.id} className="hover:bg-gray-50 cursor-pointer">
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                      {task.task_id || `#${task.id}`}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {getTaskTypeText(task)}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-500">
                      <div className="max-w-xs truncate">
                        {getTaskName(task)}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {task.operator || task.username || '-'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${getStatusColor(task.task_status || task.status)}`}>
                        {getStatusText(task.task_status || task.status)}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {formatTimestamp(task.ctime || task.created_at)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <button
                        onClick={() => handleViewDetail(task)}
                        className="text-indigo-600 hover:text-indigo-900"
                      >
                        查看详情
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {totalPages > 1 && (
            <div className="mt-6 flex justify-center space-x-2">
              <button
                onClick={() => setPageNo(Math.max(1, pageNo - 1))}
                disabled={pageNo === 1}
                className="px-4 py-2 rounded-lg bg-white border border-gray-300 text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                上一页
              </button>
              <span className="px-4 py-2 text-gray-700">
                第 {pageNo} / {totalPages} 页 (共 {total} 条)
              </span>
              <button
                onClick={() => setPageNo(Math.min(totalPages, pageNo + 1))}
                disabled={pageNo === totalPages}
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
