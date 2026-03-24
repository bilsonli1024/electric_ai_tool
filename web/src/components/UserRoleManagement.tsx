import React, { useState, useEffect } from 'react';
import { apiClient } from '../services/api';

interface User {
  id: number;
  username: string;
  email: string;
  user_type: number;
  user_status: number;
}

interface Role {
  id: number;
  role_name: string;
  role_desc: string;
  role_status: number;
}

interface UserRole {
  id: number;
  user_id: number;
  username: string;
  email: string;
  role_id: number;
  role_name: string;
  ctime: number;
}

export const UserRoleManagement: React.FC = () => {
  const [userRoles, setUserRoles] = useState<UserRole[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [loading, setLoading] = useState(true);
  const [showAssignModal, setShowAssignModal] = useState(false);
  const [selectedUserId, setSelectedUserId] = useState<number>(0);
  const [selectedRoleId, setSelectedRoleId] = useState<number>(0);
  const [toast, setToast] = useState<{ message: string; type: 'success' | 'error' } | null>(null);

  // 加载数据
  const loadData = async () => {
    try {
      setLoading(true);
      const [userRolesRes, usersRes, rolesRes] = await Promise.all([
        apiClient.get<{ data: UserRole[]; total: number }>('/api/admin/user-roles'),
        apiClient.get<{ data: User[]; total: number }>('/api/admin/users'),
        apiClient.get<{ data: Role[]; total: number }>('/api/admin/roles'),
      ]);
      
      setUserRoles(userRolesRes.data);
      setUsers(usersRes.data);
      setRoles(rolesRes.data);
    } catch (error: any) {
      setToast({ message: '加载失败: ' + error.message, type: 'error' });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  // 分配角色
  const handleAssignRole = async () => {
    if (!selectedUserId || !selectedRoleId) {
      setToast({ message: '请选择用户和角色', type: 'error' });
      return;
    }

    try {
      await apiClient.post('/api/admin/user-roles/assign', {
        user_id: selectedUserId,
        role_id: selectedRoleId,
      });
      setToast({ message: '角色分配成功', type: 'success' });
      setShowAssignModal(false);
      setSelectedUserId(0);
      setSelectedRoleId(0);
      loadData();
    } catch (error: any) {
      setToast({ message: '分配失败: ' + error.message, type: 'error' });
    }
  };

  // 移除角色
  const handleRemoveRole = async (userId: number, roleId: number) => {
    if (!confirm('确定要移除该角色吗？')) {
      return;
    }

    try {
      await apiClient.post('/api/admin/user-roles/remove', {
        user_id: userId,
        role_id: roleId,
      });
      setToast({ message: '角色移除成功', type: 'success' });
      loadData();
    } catch (error: any) {
      setToast({ message: '移除失败: ' + error.message, type: 'error' });
    }
  };

  // 格式化时间
  const formatTime = (timestamp: number) => {
    return new Date(timestamp * 1000).toLocaleString('zh-CN');
  };

  // 获取用户状态文本
  const getUserStatusText = (status: number) => {
    const statusMap: { [key: number]: string } = {
      0: '待审批',
      1: '正常',
      2: '已删除',
    };
    return statusMap[status] || '未知';
  };

  return (
    <div className="space-y-6">
      {/* Toast */}
      {toast && (
        <div
          className={`fixed top-4 right-4 px-6 py-3 rounded-lg shadow-lg ${
            toast.type === 'success' ? 'bg-green-500' : 'bg-red-500'
          } text-white z-50`}
        >
          {toast.message}
          <button
            onClick={() => setToast(null)}
            className="ml-4 text-white hover:text-gray-200"
          >
            ×
          </button>
        </div>
      )}

      {/* 标题和操作按钮 */}
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-gray-900">用户角色管理</h1>
        <button
          onClick={() => setShowAssignModal(true)}
          className="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors"
        >
          ➕ 分配角色
        </button>
      </div>

      {/* 用户角色列表 */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                用户ID
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                用户名
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                邮箱
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                角色
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                分配时间
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                操作
              </th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {loading ? (
              <tr>
                <td colSpan={6} className="px-6 py-4 text-center text-gray-500">
                  加载中...
                </td>
              </tr>
            ) : userRoles.length === 0 ? (
              <tr>
                <td colSpan={6} className="px-6 py-4 text-center text-gray-500">
                  暂无数据
                </td>
              </tr>
            ) : (
              userRoles.map((ur) => (
                <tr key={ur.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {ur.user_id}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {ur.username}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {ur.email}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    <span className="px-2 py-1 bg-indigo-100 text-indigo-800 rounded-full text-xs">
                      {ur.role_name}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {formatTime(ur.ctime)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    <button
                      onClick={() => handleRemoveRole(ur.user_id, ur.role_id)}
                      className="text-red-600 hover:text-red-900"
                    >
                      移除
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* 分配角色模态框 */}
      {showAssignModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-md">
            <h2 className="text-xl font-bold mb-4">分配角色</h2>
            
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  选择用户
                </label>
                <select
                  value={selectedUserId}
                  onChange={(e) => setSelectedUserId(Number(e.target.value))}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                >
                  <option value={0}>请选择用户</option>
                  {users.map((user) => (
                    <option key={user.id} value={user.id}>
                      {user.username} ({user.email}) - {getUserStatusText(user.user_status)}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  选择角色
                </label>
                <select
                  value={selectedRoleId}
                  onChange={(e) => setSelectedRoleId(Number(e.target.value))}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                >
                  <option value={0}>请选择角色</option>
                  {roles.map((role) => (
                    <option key={role.id} value={role.id}>
                      {role.role_name} - {role.role_desc}
                    </option>
                  ))}
                </select>
              </div>
            </div>

            <div className="flex gap-3 mt-6">
              <button
                onClick={handleAssignRole}
                className="flex-1 px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors"
              >
                确定
              </button>
              <button
                onClick={() => {
                  setShowAssignModal(false);
                  setSelectedUserId(0);
                  setSelectedRoleId(0);
                }}
                className="flex-1 px-4 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 transition-colors"
              >
                取消
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
