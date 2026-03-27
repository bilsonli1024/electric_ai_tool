import React, { useState, useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { Sparkles, Download, ArrowLeft, Loader2, CheckCircle, Copy } from 'lucide-react';
import { apiClient } from '../services/api';
import { Toast, ToastType } from './Toast';

const isImageTaskType = (taskType: unknown): boolean => {
  return taskType === 2 || taskType === '2' || taskType === 'image';
};

export const ImageGenerationResult: React.FC = () => {
  const location = useLocation();
  const navigate = useNavigate();
  
  const [isLoading, setIsLoading] = useState(true);
  const [taskId, setTaskId] = useState<string>('');
  const [generatedImageUrls, setGeneratedImageUrls] = useState<string[]>([]);
  const [taskInfo, setTaskInfo] = useState<any>(null);
  const [toast, setToast] = useState<{ message: string; type: ToastType } | null>(null);

  useEffect(() => {
    const params = new URLSearchParams(location.search);
    const taskIdFromUrl = params.get('task_id');
    
    if (taskIdFromUrl) {
      setTaskId(taskIdFromUrl);
      loadTaskResult(taskIdFromUrl);
    } else {
      navigate('/image-generation');
    }
  }, [location.search]);

  const loadTaskResult = async (taskId: string) => {
    setIsLoading(true);
    try {
      const response = await apiClient.getTaskCenterDetail(taskId);
      const detail = response.data;
      
      if (!isImageTaskType(detail.task_type)) {
        setToast({ message: '任务类型不匹配', type: 'error' });
        navigate('/image-generation');
        return;
      }

      const imageDetail = detail.detail_data as any;
      setTaskInfo({
        sku: imageDetail.sku,
        keywords: imageDetail.keywords,
        sellingPoints: imageDetail.selling_points,
        model: imageDetail.generate_model,
        status: detail.task_status,
      });
      
      if (imageDetail.generated_image_urls) {
        const urls = imageDetail.generated_image_urls.split(',').filter((url: string) => url.trim());
        setGeneratedImageUrls(urls);
      }
    } catch (error: any) {
      setToast({ message: '加载失败: ' + error.message, type: 'error' });
    } finally {
      setIsLoading(false);
    }
  };

  const handleDownloadAll = () => {
    generatedImageUrls.forEach((url, index) => {
      const link = document.createElement('a');
      link.href = url;
      link.download = `generated_image_${index + 1}.png`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    });
    setToast({ message: '开始下载所有图片', type: 'success' });
  };

  const handleGenerateAgain = () => {
    navigate('/image-generation');
  };

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <Loader2 className="animate-spin text-purple-600 mx-auto mb-4" size={48} />
          <p className="text-lg text-gray-600">加载中...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      {toast && (
        <Toast
          message={toast.message}
          type={toast.type}
          duration={3000}
          onClose={() => setToast(null)}
        />
      )}
      
      {/* Header */}
      <div className="mb-8">
        <button
          onClick={() => navigate('/image-generation')}
          className="flex items-center gap-2 text-purple-600 hover:text-purple-700 mb-4 transition-colors"
        >
          <ArrowLeft size={20} />
          返回图片生成
        </button>
        
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-12 h-12 bg-gradient-to-r from-purple-500 to-indigo-500 rounded-xl flex items-center justify-center text-white">
              <Sparkles size={24} />
            </div>
            <div>
              <h2 className="text-3xl font-bold text-gray-800">生成结果</h2>
              <p className="text-gray-600">任务ID: {taskId}</p>
            </div>
          </div>
          
          {taskInfo?.status === 'completed' && (
            <div className="flex items-center gap-2 text-green-600 bg-green-50 px-4 py-2 rounded-lg">
              <CheckCircle size={20} />
              <span className="font-semibold">生成成功</span>
            </div>
          )}
        </div>
      </div>

      {/* Task Info */}
      {taskInfo && (
        <div className="bg-white rounded-2xl shadow-lg p-6 mb-6">
          <h3 className="text-lg font-bold mb-4">任务信息</h3>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
            <div>
              <span className="text-gray-500">SKU:</span>
              <p className="font-semibold mt-1">{taskInfo.sku || '-'}</p>
            </div>
            <div>
              <span className="text-gray-500">关键词:</span>
              <p className="font-semibold mt-1">{taskInfo.keywords || '-'}</p>
            </div>
            <div>
              <span className="text-gray-500">AI模型:</span>
              <p className="font-semibold mt-1">{taskInfo.model}</p>
            </div>
            <div>
              <span className="text-gray-500">图片数量:</span>
              <p className="font-semibold mt-1">{generatedImageUrls.length} 张</p>
            </div>
          </div>
        </div>
      )}

      {/* Generated Images */}
      {generatedImageUrls.length > 0 ? (
        <div className="bg-white rounded-2xl shadow-lg p-8">
          <div className="flex items-center justify-between mb-6">
            <h3 className="text-xl font-bold">生成的图片</h3>
            <button
              onClick={handleDownloadAll}
              className="flex items-center gap-2 px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700 transition-all"
            >
              <Download size={18} />
              下载全部
            </button>
          </div>
          
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {generatedImageUrls.map((url, index) => (
              <div key={index} className="relative group">
                <div className="aspect-square rounded-xl overflow-hidden border-2 border-gray-200 hover:border-purple-500 transition-all">
                  <img
                    src={url}
                    alt={`Generated ${index + 1}`}
                    className="w-full h-full object-cover"
                  />
                </div>
                <div className="absolute inset-0 bg-black/0 group-hover:bg-black/40 transition-all rounded-xl flex items-center justify-center opacity-0 group-hover:opacity-100">
                  <div className="flex gap-2">
                    <a
                      href={url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="bg-white text-purple-600 px-4 py-2 rounded-lg font-semibold hover:bg-purple-50 transition-all"
                    >
                      查看大图
                    </a>
                    <a
                      href={url}
                      download={`generated_${index + 1}.png`}
                      className="bg-purple-600 text-white px-4 py-2 rounded-lg font-semibold hover:bg-purple-700 transition-all"
                    >
                      <Download size={18} />
                    </a>
                  </div>
                </div>
                <div className="mt-2 text-center text-sm text-gray-600">
                  图片 {index + 1}
                </div>
              </div>
            ))}
          </div>

          <div className="mt-8 pt-6 border-t border-gray-200 flex justify-center gap-4">
            <button
              onClick={handleGenerateAgain}
              className="px-6 py-3 bg-gradient-to-r from-purple-600 to-indigo-600 text-white rounded-xl font-semibold hover:from-purple-700 hover:to-indigo-700 transition-all flex items-center gap-2"
            >
              <Sparkles size={20} />
              再次生成
            </button>
            <button
              onClick={() => navigate('/tasks')}
              className="px-6 py-3 bg-gray-100 text-gray-700 rounded-xl font-semibold hover:bg-gray-200 transition-all"
            >
              返回任务中心
            </button>
          </div>
        </div>
      ) : (
        <div className="bg-white rounded-2xl shadow-lg p-12 text-center">
          <p className="text-gray-500 text-lg">暂无生成结果</p>
          <button
            onClick={handleGenerateAgain}
            className="mt-4 px-6 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700 transition-all"
          >
            重新生成
          </button>
        </div>
      )}
    </div>
  );
};
