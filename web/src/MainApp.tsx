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
    <div className="min-h-screen bg-gray-50">
      <Navbar currentPage={currentPage} onNavigate={setCurrentPage} />
      
      <div className="py-6">
        {currentPage === 'copywriting' && <CopywritingGenerator />}
        {currentPage === 'generator' && <ImageGenerationPage />}
        {currentPage === 'tasks' && <TaskCenter />}
        {currentPage === 'user' && <UserManagement />}
        {currentPage === 'modeltest' && <ModelTest />}
      </div>
    </div>
  );
};
