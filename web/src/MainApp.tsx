import React, { useState, useEffect } from 'react';
import { Auth } from './components/Auth';
import { Navbar } from './components/Navbar';
import { TaskCenter } from './components/TaskCenter';
import { UserManagement } from './components/UserManagement';
import { ModelTest } from './components/ModelTest';
import { CopywritingGenerator } from './components/CopywritingGenerator';
import { ImageGenerationPage } from './components/ImageGenerationPage';
import { apiClient } from './services/api';

type Page = 'copywriting' | 'generator' | 'tasks' | 'user' | 'modeltest';

export const MainApp: React.FC = () => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [loading, setLoading] = useState(true);
  const [currentPage, setCurrentPage] = useState<Page>('copywriting');

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

        <nav className="flex-1 p-4 space-y-2">
          <button
            onClick={() => setCurrentPage('copywriting')}
            className={`w-full text-left px-4 py-3 rounded-lg transition-colors ${
              currentPage === 'copywriting'
                ? 'bg-indigo-50 text-indigo-600 font-medium'
                : 'text-gray-700 hover:bg-gray-50'
            }`}
          >
            📝 文案生成
          </button>
          <button
            onClick={() => setCurrentPage('generator')}
            className={`w-full text-left px-4 py-3 rounded-lg transition-colors ${
              currentPage === 'generator'
                ? 'bg-indigo-50 text-indigo-600 font-medium'
                : 'text-gray-700 hover:bg-gray-50'
            }`}
          >
            🎨 图片生成
          </button>
          <button
            onClick={() => setCurrentPage('tasks')}
            className={`w-full text-left px-4 py-3 rounded-lg transition-colors ${
              currentPage === 'tasks'
                ? 'bg-indigo-50 text-indigo-600 font-medium'
                : 'text-gray-700 hover:bg-gray-50'
            }`}
          >
            📊 任务中心
          </button>
          <button
            onClick={() => setCurrentPage('modeltest')}
            className={`w-full text-left px-4 py-3 rounded-lg transition-colors ${
              currentPage === 'modeltest'
                ? 'bg-indigo-50 text-indigo-600 font-medium'
                : 'text-gray-700 hover:bg-gray-50'
            }`}
          >
            🧪 联通性测试
          </button>
          <button
            onClick={() => setCurrentPage('user')}
            className={`w-full text-left px-4 py-3 rounded-lg transition-colors ${
              currentPage === 'user'
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
          {currentPage === 'copywriting' && <CopywritingGenerator />}
          {currentPage === 'generator' && <ImageGenerationPage />}
          {currentPage === 'tasks' && <TaskCenter />}
          {currentPage === 'user' && <UserManagement />}
          {currentPage === 'modeltest' && <ModelTest />}
        </div>
      </div>
    </div>
  );
};
