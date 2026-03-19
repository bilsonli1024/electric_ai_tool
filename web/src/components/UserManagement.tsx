import React, { useState, useEffect } from 'react';
import { User } from '../types';
import { apiClient } from '../services/api';

export const UserManagement: React.FC = () => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadUser();
  }, []);

  const loadUser = async () => {
    setLoading(true);
    try {
      const userData = await apiClient.me();
      setUser(userData);
    } catch (err) {
      console.error('Failed to load user:', err);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-8">
        <div className="text-center py-12">
          <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600"></div>
          <p className="mt-4 text-gray-600">加载中...</p>
        </div>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-8">
        <div className="text-center py-12 bg-white rounded-lg shadow">
          <p className="text-gray-500">无法加载用户信息</p>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <h2 className="text-3xl font-bold text-gray-800 mb-6">用户管理</h2>

      <div className="bg-white rounded-lg shadow-lg overflow-hidden">
        <div className="bg-gradient-to-r from-indigo-500 to-purple-600 px-6 py-8">
          <div className="flex items-center space-x-4">
            <div className="w-20 h-20 rounded-full bg-white flex items-center justify-center text-indigo-600 text-3xl font-bold shadow-lg">
              {user.username.charAt(0).toUpperCase()}
            </div>
            <div className="text-white">
              <h3 className="text-2xl font-bold">{user.username}</h3>
              <p className="text-indigo-100">{user.email}</p>
            </div>
          </div>
        </div>

        <div className="p-6 space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-2">
              <label className="text-sm font-medium text-gray-500">用户ID</label>
              <div className="text-lg text-gray-800">#{user.id}</div>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium text-gray-500">账号状态</label>
              <div>
                <span className={`px-3 py-1 rounded-full text-sm font-semibold ${
                  user.status === 1
                    ? 'bg-green-100 text-green-800'
                    : 'bg-red-100 text-red-800'
                }`}>
                  {user.status === 1 ? '正常' : '已禁用'}
                </span>
              </div>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium text-gray-500">注册时间</label>
              <div className="text-lg text-gray-800">
                {new Date(user.created_at).toLocaleString('zh-CN')}
              </div>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium text-gray-500">最后登录</label>
              <div className="text-lg text-gray-800">
                {user.last_login_at
                  ? new Date(user.last_login_at).toLocaleString('zh-CN')
                  : '从未登录'}
              </div>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium text-gray-500">用户名</label>
              <div className="text-lg text-gray-800">{user.username}</div>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium text-gray-500">邮箱</label>
              <div className="text-lg text-gray-800">{user.email}</div>
            </div>
          </div>

          <div className="border-t border-gray-200 pt-6 mt-6">
            <h4 className="text-lg font-semibold text-gray-800 mb-4">账号信息</h4>
            <div className="bg-blue-50 border-l-4 border-blue-500 p-4 rounded">
              <div className="flex items-start">
                <div className="flex-shrink-0">
                  <svg className="h-5 w-5 text-blue-500" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
                  </svg>
                </div>
                <div className="ml-3">
                  <p className="text-sm text-blue-700">
                    您的账号已激活并可以正常使用所有功能。如需修改密码或其他设置，请联系管理员。
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
