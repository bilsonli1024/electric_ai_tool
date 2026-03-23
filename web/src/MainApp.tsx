import React, { useState, useEffect } from 'react';
import { BrowserRouter, Routes, Route, useNavigate, useLocation } from 'react-router-dom';
import { Auth } from './components/Auth';
import { Navbar } from './components/Navbar';
import { TaskCenter } from './components/TaskCenter';
import { UserManagement } from './components/UserManagement';
import { UserList } from './components/UserList';
import { RoleList } from './components/RoleList';
import { PermissionList } from './components/PermissionList';
import { RolePermissionList } from './components/RolePermissionList';
import { ModelTest } from './components/ModelTest';
import { CopywritingGenerator } from './components/CopywritingGenerator';
import { ImageGenerationPage } from './components/ImageGenerationPage';
import { ImageGenerationResult } from './components/ImageGenerationResult';
import ErrorToastContainer from './components/ErrorToastContainer';
import { apiClient } from './services/api';
import { User } from './types';

type Page = 'copywriting' | 'generator' | 'tasks' | 'user' | 'modeltest' | 'admin-users' | 'admin-roles' | 'admin-permissions' | 'admin-role-permissions';

const AppContent: React.FC<{ isAuthenticated: boolean; setIsAuthenticated: (val: boolean) => void }> = ({ 
  isAuthenticated, 
  setIsAuthenticated 
}) => {
  const navigate = useNavigate();
  const location = useLocation();
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  
  useEffect(() => {
    loadCurrentUser();
  }, []);

  const loadCurrentUser = async () => {
    try {
      const user = await apiClient.me();
      setCurrentUser(user);
    } catch (err) {
      console.error('Failed to load current user:', err);
    }
  };

  const isAdmin = currentUser?.user_type === 99;
  
  const currentPage = (): Page => {
    const path = location.pathname;
    if (path.startsWith('/copywriting')) return 'copywriting';
    if (path.startsWith('/image-generation')) return 'generator';
    if (path.startsWith('/tasks')) return 'tasks';
    if (path === '/user') return 'user';
    if (path === '/admin/users') return 'admin-users';
    if (path === '/admin/roles') return 'admin-roles';
    if (path === '/admin/permissions') return 'admin-permissions';
    if (path === '/admin/role-permissions') return 'admin-role-permissions';
    if (path.startsWith('/modeltest')) return 'modeltest';
    return 'copywriting';
  };

  const setCurrentPage = (page: Page) => {
    const paths = {
      copywriting: '/copywriting',
      generator: '/image-generation',
      tasks: '/tasks',
      user: '/user',
      'admin-users': '/admin/users',
      'admin-roles': '/admin/roles',
      'admin-permissions': '/admin/permissions',
      'admin-role-permissions': '/admin/role-permissions',
      modeltest: '/modeltest'
    };
    navigate(paths[page]);
  };

  return (
    <div className="min-h-screen bg-gray-50 flex">
      {/* 左侧边栏 */}
      <div className="w-64 bg-white border-r border-gray-200 flex flex-col fixed h-full">
        <div className="p-6 border-b border-gray-200">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-indigo-600 rounded-lg flex items-center justify-center text-white font-bold text-lg">
              E
            </div>
            <div>
              <h1 className="font-bold text-lg">Electric AI</h1>
              <p className="text-xs text-gray-500">智能营销工具</p>
            </div>
          </div>
        </div>

        <nav className="flex-1 p-4 space-y-2 overflow-y-auto">
          <button
            onClick={() => setCurrentPage('copywriting')}
            className={`w-full text-left px-4 py-3 rounded-lg transition-colors ${
              currentPage() === 'copywriting'
                ? 'bg-indigo-50 text-indigo-600 font-medium'
                : 'text-gray-700 hover:bg-gray-50'
            }`}
          >
            📝 文案生成
          </button>
          <button
            onClick={() => setCurrentPage('generator')}
            className={`w-full text-left px-4 py-3 rounded-lg transition-colors ${
              currentPage() === 'generator'
                ? 'bg-indigo-50 text-indigo-600 font-medium'
                : 'text-gray-700 hover:bg-gray-50'
            }`}
          >
            🎨 图片生成
          </button>
          <button
            onClick={() => setCurrentPage('tasks')}
            className={`w-full text-left px-4 py-3 rounded-lg transition-colors ${
              currentPage() === 'tasks'
                ? 'bg-indigo-50 text-indigo-600 font-medium'
                : 'text-gray-700 hover:bg-gray-50'
            }`}
          >
            📊 任务中心
          </button>
          <button
            onClick={() => setCurrentPage('modeltest')}
            className={`w-full text-left px-4 py-3 rounded-lg transition-colors ${
              currentPage() === 'modeltest'
                ? 'bg-indigo-50 text-indigo-600 font-medium'
                : 'text-gray-700 hover:bg-gray-50'
            }`}
          >
            🧪 联通性测试
          </button>

          {isAdmin && (
            <>
              <div className="pt-4 mt-4 border-t border-gray-200">
                <p className="text-xs font-semibold text-gray-400 uppercase tracking-wider px-4 mb-2">
                  管理员功能
                </p>
              </div>
              <button
                onClick={() => setCurrentPage('admin-users')}
                className={`w-full text-left px-4 py-3 rounded-lg transition-colors ${
                  currentPage() === 'admin-users'
                    ? 'bg-indigo-50 text-indigo-600 font-medium'
                    : 'text-gray-700 hover:bg-gray-50'
                }`}
              >
                👥 用户列表
              </button>
              <button
                onClick={() => setCurrentPage('admin-roles')}
                className={`w-full text-left px-4 py-3 rounded-lg transition-colors ${
                  currentPage() === 'admin-roles'
                    ? 'bg-indigo-50 text-indigo-600 font-medium'
                    : 'text-gray-700 hover:bg-gray-50'
                }`}
              >
                🎭 角色列表
              </button>
              <button
                onClick={() => setCurrentPage('admin-permissions')}
                className={`w-full text-left px-4 py-3 rounded-lg transition-colors ${
                  currentPage() === 'admin-permissions'
                    ? 'bg-indigo-50 text-indigo-600 font-medium'
                    : 'text-gray-700 hover:bg-gray-50'
                }`}
              >
                🔑 权限列表
              </button>
              <button
                onClick={() => setCurrentPage('admin-role-permissions')}
                className={`w-full text-left px-4 py-3 rounded-lg transition-colors ${
                  currentPage() === 'admin-role-permissions'
                    ? 'bg-indigo-50 text-indigo-600 font-medium'
                    : 'text-gray-700 hover:bg-gray-50'
                }`}
              >
                🔗 角色权限
              </button>
            </>
          )}

          <div className="pt-4 mt-4 border-t border-gray-200"></div>
          <button
            onClick={() => setCurrentPage('user')}
            className={`w-full text-left px-4 py-3 rounded-lg transition-colors ${
              currentPage() === 'user'
                ? 'bg-indigo-50 text-indigo-600 font-medium'
                : 'text-gray-700 hover:bg-gray-50'
            }`}
          >
            👤 用户管理
          </button>
        </nav>

        <div className="p-4 border-t border-gray-200">
          <button
            onClick={async () => {
              await apiClient.logout();
              setIsAuthenticated(false);
            }}
            className="w-full px-4 py-3 text-left text-red-600 hover:bg-red-50 rounded-lg transition-colors"
          >
            🚪 退出登录
          </button>
        </div>
      </div>

      {/* 主内容区域 */}
      <div className="flex-1 ml-64">
        <div className="py-6 px-8">
          <Routes>
            <Route path="/copywriting" element={<CopywritingGenerator />} />
            <Route path="/image-generation" element={<ImageGenerationPage />} />
            <Route path="/image-generation/result" element={<ImageGenerationResult />} />
            <Route path="/tasks" element={<TaskCenter />} />
            <Route path="/user" element={<UserManagement />} />
            <Route path="/admin/users" element={<UserList />} />
            <Route path="/admin/roles" element={<RoleList />} />
            <Route path="/admin/permissions" element={<PermissionList />} />
            <Route path="/admin/role-permissions" element={<RolePermissionList />} />
            <Route path="/modeltest" element={<ModelTest />} />
            <Route path="/" element={<CopywritingGenerator />} />
          </Routes>
        </div>
      </div>
    </div>
  );
};

export const MainApp: React.FC = () => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    checkAuth();
  }, []);

  const checkAuth = async () => {
    try {
      if (apiClient.isAuthenticated()) {
        await apiClient.me();
        setIsAuthenticated(true);
      }
    } catch (err) {
      console.error('Auth check failed:', err);
      setIsAuthenticated(false);
    } finally {
      setLoading(false);
    }
  };

  const handleAuthSuccess = () => {
    setIsAuthenticated(true);
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600"></div>
          <p className="mt-4 text-gray-600">加载中...</p>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Auth onAuthSuccess={handleAuthSuccess} />;
  }

  return (
    <BrowserRouter>
      <ErrorToastContainer />
      <AppContent isAuthenticated={isAuthenticated} setIsAuthenticated={setIsAuthenticated} />
    </BrowserRouter>
  );
};
