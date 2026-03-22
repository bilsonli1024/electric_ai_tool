import React, { useState, useEffect } from 'react';
import { Upload, Tag, Hash, Sparkles, Loader2 } from 'lucide-react';
import { useLocation, useNavigate } from 'react-router-dom';
import { CopywritingSelector } from './CopywritingSelector';
import { apiClient } from '../services/api';
import { Toast, ToastType } from './Toast';

export const ImageGenerationPage: React.FC = () => {
  const location = useLocation();
  const navigate = useNavigate();
  
  const [sku, setSku] = useState('');
  const [keywords, setKeywords] = useState('');
  const [sellingPoints, setSellingPoints] = useState('');
  const [competitorLink, setCompetitorLink] = useState('');
  const [selectedCopywritingTaskId, setSelectedCopywritingTaskId] = useState<string | null>(null);
  const [uploadedImages, setUploadedImages] = useState<File[]>([]);
  const [imagePreviews, setImagePreviews] = useState<string[]>([]);
  const [isDragging, setIsDragging] = useState(false);
  const [selectedModel, setSelectedModel] = useState('gemini');
  const [isGenerating, setIsGenerating] = useState(false);
  const [toast, setToast] = useState<{ message: string; type: ToastType } | null>(null);
  const [currentTaskId, setCurrentTaskId] = useState<string | null>(null);
  const [taskStatus, setTaskStatus] = useState<string>('');
  const [generatedImageUrls, setGeneratedImageUrls] = useState<string[]>([]);
  const [isLoadingTask, setIsLoadingTask] = useState(false);

  // 从URL加载任务
  useEffect(() => {
    const params = new URLSearchParams(location.search);
    const taskIdFromUrl = params.get('task_id');
    
    if (taskIdFromUrl) {
      loadTaskDetail(taskIdFromUrl);
    }
  }, [location.search]);

  // 轮询任务状态
  useEffect(() => {
    // 只有在生成中才轮询
    if (!currentTaskId || !isGenerating || taskStatus === 'completed' || taskStatus === 'failed') {
      return;
    }

    const pollInterval = setInterval(() => {
      pollTaskStatus(currentTaskId);
    }, 3000); // 每3秒轮询一次

    return () => clearInterval(pollInterval);
  }, [currentTaskId, isGenerating, taskStatus]);

  const loadTaskDetail = async (taskId: string) => {
    setIsLoadingTask(true);
    try {
      const response = await apiClient.getTaskCenterDetail(taskId);
      const detail = response.data;
      
      if (detail.task_type !== 'image') {
        setToast({ message: '任务类型不匹配', type: 'error' });
        return;
      }

      const imageDetail = detail.detail_data as any;
      
      // 恢复表单数据
      setSku(imageDetail.sku || '');
      setKeywords(imageDetail.keywords || '');
      setSellingPoints(imageDetail.selling_points || '');
      setCompetitorLink(imageDetail.competitor_link || '');
      setSelectedModel(imageDetail.generate_model || 'gemini');
      setSelectedCopywritingTaskId(imageDetail.copywriting_task_id || null);
      
      // 设置任务状态
      setCurrentTaskId(taskId);
      setTaskStatus(detail.task_status);
      
      // 如果已完成，显示结果
      if (detail.task_status === 'completed' && imageDetail.generated_image_urls) {
        const urls = imageDetail.generated_image_urls.split(',').filter((url: string) => url.trim());
        setGeneratedImageUrls(urls);
      }
      
      // 如果正在生成，开始轮询
      if (detail.task_status === 'ongoing') {
        setIsGenerating(true);
      }
      
    } catch (error: any) {
      setToast({ message: '加载任务失败: ' + error.message, type: 'error' });
    } finally {
      setIsLoadingTask(false);
    }
  };

  const pollTaskStatus = async (taskId: string) => {
    try {
      const response = await apiClient.getTaskCenterDetail(taskId);
      const detail = response.data;
      const imageDetail = detail.detail_data as any;
      
      setTaskStatus(detail.task_status);
      
      if (detail.task_status === 'completed') {
        setIsGenerating(false);
        if (imageDetail.generated_image_urls) {
          const urls = imageDetail.generated_image_urls.split(',').filter((url: string) => url.trim());
          setGeneratedImageUrls(urls);
          // 跳转到结果页面
          navigate(`/image-generation/result?task_id=${taskId}`);
        }
        setToast({ message: '图片生成成功！', type: 'success' });
      } else if (detail.task_status === 'failed') {
        setIsGenerating(false);
        setToast({ message: '图片生成失败: ' + (imageDetail.error_message || '未知错误'), type: 'error' });
      }
    } catch (error: any) {
      console.error('Failed to poll task status:', error);
    }
  };

  const handleCopywritingSelect = (task: any) => {
    try {
      const copy = JSON.parse(task.generated_copy);
      
      // 自动填充关键词和卖点
      if (copy.searchTerms) {
        const keywordsFromCopy = copy.searchTerms.split(' ').slice(0, 10).join(', ');
        setKeywords(keywordsFromCopy);
      }
      
      if (copy.bulletPoints && copy.bulletPoints.length > 0) {
        setSellingPoints(copy.bulletPoints.join('\n'));
      }

      setSelectedCopywritingTaskId(task.id);
    } catch (error) {
      console.error('Failed to parse copywriting task:', error);
    }
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files || []) as File[];
    processFiles(files);
  };

  const handleDragOver = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    setIsDragging(true);
  };

  const handleDragLeave = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    setIsDragging(false);
  };

  const handleDrop = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    setIsDragging(false);
    const files = Array.from(e.dataTransfer.files) as File[];
    processFiles(files);
  };

  const processFiles = (files: File[]) => {
    const imageFiles = files.filter(file => file.type.startsWith('image/'));
    
    if (imageFiles.length === 0) {
      alert('请上传图片文件');
      return;
    }

    setUploadedImages(prev => [...prev, ...imageFiles]);

    // 生成预览
    imageFiles.forEach(file => {
      const reader = new FileReader();
      reader.onloadend = () => {
        setImagePreviews(prev => [...prev, reader.result as string]);
      };
      reader.readAsDataURL(file);
    });
  };

  const removeImage = (index: number) => {
    setUploadedImages(prev => prev.filter((_, i) => i !== index));
    setImagePreviews(prev => prev.filter((_, i) => i !== index));
  };

  const handleGenerateImages = async () => {
    if (!sku && !keywords && !sellingPoints) {
      setToast({ message: '请至少填写SKU、关键词或产品卖点', type: 'error' });
      return;
    }

    setIsGenerating(true);
    try {
      const taskName = `图片生成_${sku || keywords}_${Date.now()}`;
      
      const response = await apiClient.generateImages({
        sku,
        keywords,
        sellingPoints,
        competitorLink,
        model: selectedModel,
        taskName,
        copywritingTaskId: selectedCopywritingTaskId || undefined,
      });
      
      // 设置当前任务ID并开始轮询
      setCurrentTaskId(response.task_id);
      setTaskStatus('ongoing');
      setGeneratedImageUrls([]); // 清空之前的结果
      
      setToast({ 
        message: `图片生成任务已创建！任务ID: ${response.task_id}\n正在生成中，请稍候...也可在任务中心查看进度`, 
        type: 'success' 
      });
      
      // 不清空表单，保持当前状态
    } catch (error: any) {
      setToast({ message: '生成失败: ' + error.message, type: 'error' });
    } finally {
      setIsGenerating(false);
    }
  };

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      {isLoadingTask && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-2xl p-8 flex flex-col items-center gap-4">
            <Loader2 className="animate-spin text-purple-600" size={48} />
            <p className="text-lg font-semibold">加载任务中...</p>
          </div>
        </div>
      )}
      
      {toast && (
        <Toast
          message={toast.message}
          type={toast.type}
          duration={5000}
          onClose={() => setToast(null)}
        />
      )}
      <div className="mb-8">
        <div className="flex items-center gap-3 mb-2">
          <div className="w-10 h-10 bg-gradient-to-r from-purple-500 to-indigo-500 rounded-xl flex items-center justify-center text-white">
            <Sparkles size={20} />
          </div>
          <h2 className="text-3xl font-bold text-gray-800">图片生成</h2>
        </div>
        <p className="text-gray-600">基于产品信息生成专业的Amazon产品图</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 space-y-6">{/* CopywritingSelector 只在不是从任务恢复时显示 */}
          {!currentTaskId && <CopywritingSelector onSelectCopywriting={handleCopywritingSelect} />}

          {/* Product Information Form */}
          <div className="bg-white rounded-2xl shadow-lg p-8 space-y-6">
            <h3 className="text-xl font-bold mb-4">产品信息</h3>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="space-y-2">
                <label className="text-sm font-semibold flex items-center gap-2">
                  <Hash size={16} className="text-gray-400" />
                  SKU
                </label>
                <input 
                  type="text" 
                  placeholder="例如: CHAIR-001-BLK"
                  className="w-full px-4 py-3 rounded-xl border border-gray-200 focus:ring-2 focus:ring-purple-500 focus:border-transparent outline-none transition-all"
                  value={sku}
                  onChange={(e) => setSku(e.target.value)}
                />
              </div>

              <div className="space-y-2">
                <label className="text-sm font-semibold flex items-center gap-2">
                  <Tag size={16} className="text-gray-400" />
                  核心关键词
                </label>
                <input 
                  type="text" 
                  placeholder="例如: Ergonomic Office Chair"
                  className="w-full px-4 py-3 rounded-xl border border-gray-200 focus:ring-2 focus:ring-purple-500 focus:border-transparent outline-none transition-all"
                  value={keywords}
                  onChange={(e) => setKeywords(e.target.value)}
                />
              </div>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-semibold">产品卖点 & 描述</label>
              <textarea 
                rows={4}
                placeholder="描述您的产品优势，AI 将基于此提炼核心卖点..."
                className="w-full px-4 py-3 rounded-xl border border-gray-200 focus:ring-2 focus:ring-purple-500 focus:border-transparent outline-none transition-all resize-none"
                value={sellingPoints}
                onChange={(e) => setSellingPoints(e.target.value)}
              />
            </div>

            <div className="space-y-2">
              <label className="text-sm font-semibold">竞品链接 (可选)</label>
              <input 
                type="text" 
                placeholder="https://amazon.com/..."
                className="w-full px-4 py-3 rounded-xl border border-gray-200 focus:ring-2 focus:ring-purple-500 focus:border-transparent outline-none transition-all"
                value={competitorLink}
                onChange={(e) => setCompetitorLink(e.target.value)}
              />
            </div>

            <div className="space-y-2">
              <label className="text-sm font-semibold">选择AI模型</label>
              <div className="grid grid-cols-3 gap-3">
                <button
                  type="button"
                  onClick={() => setSelectedModel('gemini')}
                  className={`px-4 py-3 rounded-xl border-2 transition-all font-medium ${
                    selectedModel === 'gemini'
                      ? 'border-purple-500 bg-purple-50 text-purple-700'
                      : 'border-gray-200 hover:border-purple-300'
                  }`}
                >
                  Google Gemini
                </button>
                <button
                  type="button"
                  onClick={() => setSelectedModel('gpt')}
                  className={`px-4 py-3 rounded-xl border-2 transition-all font-medium ${
                    selectedModel === 'gpt'
                      ? 'border-purple-500 bg-purple-50 text-purple-700'
                      : 'border-gray-200 hover:border-purple-300'
                  }`}
                >
                  OpenAI GPT
                </button>
                <button
                  type="button"
                  onClick={() => setSelectedModel('deepseek')}
                  className={`px-4 py-3 rounded-xl border-2 transition-all font-medium ${
                    selectedModel === 'deepseek'
                      ? 'border-purple-500 bg-purple-50 text-purple-700'
                      : 'border-gray-200 hover:border-purple-300'
                  }`}
                >
                  DeepSeek
                </button>
              </div>
            </div>

            <div className="space-y-4">
              <label className="text-sm font-semibold">产品白底图 (可多选)</label>
              <input
                type="file"
                id="image-upload"
                multiple
                accept="image/*"
                className="hidden"
                onChange={handleFileChange}
              />
              <div 
                className={`border-2 border-dashed rounded-2xl p-8 text-center transition-all cursor-pointer group ${
                  isDragging 
                    ? 'border-purple-500 bg-purple-50' 
                    : 'border-gray-200 hover:border-purple-500 hover:bg-purple-50/50'
                }`}
                onDragOver={handleDragOver}
                onDragLeave={handleDragLeave}
                onDrop={handleDrop}
                onClick={() => document.getElementById('image-upload')?.click()}
              >
                <div className="space-y-4">
                  <div className="w-16 h-16 bg-purple-50 rounded-full flex items-center justify-center mx-auto group-hover:scale-110 transition-transform">
                    <Upload className="text-purple-500" size={24} />
                  </div>
                  <div>
                    <p className="font-bold text-gray-700">点击或拖拽上传产品图</p>
                    <p className="text-sm text-gray-500">支持多张上传，AI 将自动识别产品特征</p>
                  </div>
                </div>
              </div>

              {/* Image Previews */}
              {imagePreviews.length > 0 && (
                <div className="grid grid-cols-3 gap-4 mt-4">
                  {imagePreviews.map((preview, index) => (
                    <div key={index} className="relative group">
                      <img 
                        src={preview} 
                        alt={`Preview ${index + 1}`}
                        className="w-full h-32 object-cover rounded-lg border-2 border-gray-200"
                      />
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          removeImage(index);
                        }}
                        className="absolute top-2 right-2 bg-red-500 text-white rounded-full w-6 h-6 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity"
                      >
                        ×
                      </button>
                      <div className="absolute bottom-2 left-2 bg-black/50 text-white text-xs px-2 py-1 rounded">
                        {uploadedImages[index].name}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>

            <button 
              onClick={handleGenerateImages}
              disabled={isGenerating || taskStatus === 'completed'}
              className="w-full py-4 bg-gradient-to-r from-purple-600 to-indigo-600 text-white rounded-xl font-bold text-lg hover:from-purple-700 hover:to-indigo-700 transition-all flex items-center justify-center gap-3 shadow-lg disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isGenerating ? (
                <>
                  <Loader2 className="animate-spin" />
                  正在生成中...
                </>
              ) : taskStatus === 'completed' ? (
                <>
                  已完成
                </>
              ) : (
                <>
                  开始生成图片
                  <Sparkles size={20} />
                </>
              )}
            </button>
            
            {taskStatus === 'completed' && (
              <div className="mt-4 p-4 bg-green-50 border border-green-200 rounded-xl text-center">
                <p className="text-green-800 font-semibold">此任务已完成</p>
                <button
                  onClick={() => navigate('/image-generation')}
                  className="mt-2 px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700 transition-all"
                >
                  创建新任务
                </button>
              </div>
            )}
          </div>
        </div>

        {/* Info Panel */}
        <div className="space-y-6">
          <div className="bg-gradient-to-br from-purple-50 to-indigo-50 rounded-2xl p-6 border border-purple-100">
            <h3 className="font-bold text-purple-900 mb-4">功能说明</h3>
            <ul className="space-y-3 text-sm text-purple-800">
              <li className="flex items-start gap-2">
                <span className="text-purple-500">•</span>
                <span>支持引用已生成的文案，自动填充关键词</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-purple-500">•</span>
                <span>可独立输入产品信息，不依赖文案</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-purple-500">•</span>
                <span>AI自动分析并生成多张场景图</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-purple-500">•</span>
                <span>支持自定义编辑和重新生成</span>
              </li>
            </ul>
          </div>

          {selectedCopywritingTaskId && (
            <div className="bg-green-50 rounded-2xl p-6 border border-green-200">
              <div className="flex items-center gap-2 text-green-800 font-bold mb-2">
                <Sparkles size={16} />
                已关联文案任务
              </div>
              <p className="text-sm text-green-700">
                图片生成完成后，将自动关联到文案任务 #{selectedCopywritingTaskId}
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
