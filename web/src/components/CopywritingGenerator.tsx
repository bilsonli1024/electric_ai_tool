import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'motion/react';
import { Plus, Trash2, ArrowRight, ArrowLeft, Sparkles, CheckCircle2, Loader2, Copy, Check } from 'lucide-react';
import { apiClient } from '../services/api';
import { useLocation, useNavigate } from 'react-router-dom';

type Step = 'competitors' | 'configuration' | 'result';

interface Keyword {
  original: string;
  translation: string;
  category: 'core' | 'attribute' | 'extension';
}

interface BilingualText {
  original: string;
  translation: string;
}

interface CompetitorAnalysis {
  keywords: Keyword[];
  sellingPoints: BilingualText[];
  reviewInsights: BilingualText[];
  imageInsights: BilingualText[];
}

interface ProductDetails {
  size: string;
  color: string;
  quantity: string;
  function: string;
  scenario: string;
  audience: string;
  material: string;
  sellingPoints: string;
  keywords: string;
}

interface GeneratedCopy {
  title: string;
  bulletPoints: string[];
  description: string;
  searchTerms: string;
}

export const CopywritingGenerator: React.FC = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const [step, setStep] = useState<Step>('competitors');
  const [competitorUrls, setCompetitorUrls] = useState<string[]>(['']);
  const [taskName, setTaskName] = useState<string>('');
  const [selectedModel, setSelectedModel] = useState('gemini');
  const [isAnalyzing, setIsAnalyzing] = useState(false);
  const [analysis, setAnalysis] = useState<CompetitorAnalysis | null>(null);
  const [taskId, setTaskId] = useState<string | null>(null);
  const [selectedKeywords, setSelectedKeywords] = useState<string[]>([]);
  const [selectedSellingPoints, setSelectedSellingPoints] = useState<string[]>([]);
  const [selectedReviewInsights, setSelectedReviewInsights] = useState<string[]>([]);
  const [selectedImageInsights, setSelectedImageInsights] = useState<string[]>([]);
  const [productDetails, setProductDetails] = useState<ProductDetails>({
    size: '',
    color: '',
    quantity: '',
    function: '',
    scenario: '',
    audience: '',
    material: '',
    sellingPoints: '',
    keywords: ''
  });
  const [isGenerating, setIsGenerating] = useState(false);
  const [result, setResult] = useState<GeneratedCopy | null>(null);
  const [copiedField, setCopiedField] = useState<string | null>(null);
  const [isLoadingTask, setIsLoadingTask] = useState(false);

  // 从URL参数加载任务
  useEffect(() => {
    const searchParams = new URLSearchParams(location.search);
    const taskIdParam = searchParams.get('task_id');
    
    if (taskIdParam) {
      loadTaskDetail(taskIdParam);
    }
  }, [location.search]);

  const loadTaskDetail = async (taskIdToLoad: string) => {
    setIsLoadingTask(true);
    try {
      const response = await apiClient.getTaskCenterDetail(taskIdToLoad);
      const detail = response.data; // 提取data字段
      
      if (detail.task_type !== 'copywriting') {
        alert('任务类型不匹配');
        return;
      }

      const copyDetail = detail.detail_data as any;
      
      // 设置基本信息
      setTaskId(detail.task_id);
      setSelectedModel(copyDetail.analyze_model || 'gemini');
      
      // 设置任务名称
      if (copyDetail.task_name) {
        setTaskName(copyDetail.task_name);
      }
      
      // 解析竞品链接
      if (copyDetail.competitor_urls) {
        try {
          const urls = JSON.parse(copyDetail.competitor_urls);
          setCompetitorUrls(urls.length > 0 ? urls : ['']);
        } catch (e) {
          setCompetitorUrls(['']);
        }
      }
      
      // 解析分析结果
      if (copyDetail.analysis_result) {
        try {
          const analysisData = JSON.parse(copyDetail.analysis_result);
          setAnalysis(analysisData);
          
          // 如果有用户选择的数据，使用用户选择的；否则使用分析出的全部数据
          if (copyDetail.user_selected_data) {
            const selectedData = JSON.parse(copyDetail.user_selected_data);
            setSelectedKeywords(selectedData.keywords || []);
            setSelectedSellingPoints(selectedData.sellingPoints || []);
            setSelectedReviewInsights(selectedData.reviewInsights || []);
            setSelectedImageInsights(selectedData.imageInsights || []);
          } else {
            // 默认全选
            setSelectedKeywords(analysisData.keywords?.map((k: Keyword) => k.original) || []);
            setSelectedSellingPoints(analysisData.sellingPoints?.map((p: BilingualText) => p.original) || []);
            setSelectedReviewInsights(analysisData.reviewInsights?.map((i: BilingualText) => i.original) || []);
            setSelectedImageInsights(analysisData.imageInsights?.map((i: BilingualText) => i.original) || []);
          }
        } catch (e) {
          console.error('Failed to parse analysis result:', e);
        }
      }
      
      // 解析产品详情
      if (copyDetail.product_details) {
        try {
          const details = JSON.parse(copyDetail.product_details);
          setProductDetails(details);
        } catch (e) {
          console.error('Failed to parse product details:', e);
        }
      }
      
      // 解析生成的文案
      if (copyDetail.generated_copy) {
        try {
          const copy = JSON.parse(copyDetail.generated_copy);
          setResult(copy);
        } catch (e) {
          console.error('Failed to parse generated copy:', e);
        }
      }
      
      // 根据任务状态设置步骤
      if (detail.task_status === 'pending') {
        setStep('competitors');
      } else if (detail.task_status === 'ongoing') {
        if (copyDetail.analysis_result && !copyDetail.generated_copy) {
          setStep('configuration');
        } else if (copyDetail.generated_copy) {
          setStep('result');
        } else {
          setStep('competitors');
        }
      } else if (detail.task_status === 'completed') {
        setStep('result');
      }
      
    } catch (error: any) {
      console.error('Failed to load task:', error);
      alert('加载任务失败: ' + (error.message || '未知错误'));
    } finally {
      setIsLoadingTask(false);
    }
  };

  const handleAddUrl = () => setCompetitorUrls([...competitorUrls, '']);
  
  const handleRemoveUrl = (index: number) => {
    const newUrls = competitorUrls.filter((_, i) => i !== index);
    setCompetitorUrls(newUrls.length ? newUrls : ['']);
  };
  
  const handleUrlChange = (index: number, value: string) => {
    const newUrls = [...competitorUrls];
    newUrls[index] = value;
    setCompetitorUrls(newUrls);
  };

  const handleAnalyze = async () => {
    const validUrls = competitorUrls.filter(url => url.trim() !== '');
    if (validUrls.length === 0) return;

    setIsAnalyzing(true);
    try {
      const response = await apiClient.analyzeCompetitors(validUrls, selectedModel, taskName || undefined);
      setAnalysis(response.data);
      setTaskId(response.task_id);
      setSelectedKeywords(response.data.keywords.map((k: Keyword) => k.original));
      setSelectedSellingPoints(response.data.sellingPoints.map((p: BilingualText) => p.original));
      setStep('configuration');
    } catch (error: any) {
      alert('分析失败: ' + error.message);
    } finally {
      setIsAnalyzing(false);
    }
  };

  const toggleKeyword = (keyword: string) => {
    setSelectedKeywords(prev => 
      prev.includes(keyword) ? prev.filter(k => k !== keyword) : [...prev, keyword]
    );
  };

  const toggleSellingPoint = (point: string) => {
    setSelectedSellingPoints(prev => 
      prev.includes(point) ? prev.filter(p => p !== point) : [...prev, point]
    );
  };

  const toggleReviewInsight = (insight: string) => {
    setSelectedReviewInsights(prev => 
      prev.includes(insight) ? prev.filter(i => i !== insight) : [...prev, insight]
    );
  };

  const toggleImageInsight = (insight: string) => {
    setSelectedImageInsights(prev => 
      prev.includes(insight) ? prev.filter(i => i !== insight) : [...prev, insight]
    );
  };

  const handleGenerate = async () => {
    if (!taskId) return;
    
    setIsGenerating(true);
    try {
      const response = await apiClient.generateCopy({
        task_id: taskId,
        selectedKeywords,
        selectedSellingPoints,
        selectedReviewInsights,
        selectedImageInsights,
        productDetails,
        model: selectedModel,
      });
      setResult(response.data);
      setStep('result');
    } catch (error: any) {
      alert('生成失败: ' + error.message);
    } finally {
      setIsGenerating(false);
    }
  };

  const copyToClipboard = (text: string, field: string) => {
    navigator.clipboard.writeText(text);
    setCopiedField(field);
    setTimeout(() => setCopiedField(null), 2000);
  };

  const handleCopyAll = () => {
    if (!result) return;
    const allText = `
【产品标题】
${result.title}

【5点描述】
${result.bulletPoints.join('\n')}

【产品描述】
${result.description}

【搜索词】
${result.searchTerms}
    `.trim();
    copyToClipboard(allText, 'all');
  };

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      {isLoadingTask && (
        <div className="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center">
          <div className="bg-white rounded-2xl p-8 flex flex-col items-center gap-4">
            <Loader2 className="animate-spin text-orange-500" size={48} />
            <p className="text-lg font-medium">正在加载任务数据...</p>
          </div>
        </div>
      )}
      
      <div className="mb-8">
        <div className="flex items-center gap-3 mb-2">
          <div className="w-10 h-10 bg-gradient-to-r from-orange-500 to-pink-500 rounded-xl flex items-center justify-center text-white">
            <Sparkles size={20} />
          </div>
          <h2 className="text-3xl font-bold text-gray-800">文案生成</h2>
          {taskId && <span className="text-sm text-gray-500">任务ID: {taskId}</span>}
        </div>
        <p className="text-gray-600">分析竞品并生成高转化率的产品文案</p>
      </div>

      <div className="mb-6 flex items-center gap-4 text-sm font-medium">
        <span className={`${step === 'competitors' ? 'text-orange-600' : 'text-gray-400'}`}>1. 竞品分析</span>
        <ArrowRight size={14} className="text-gray-300" />
        <span className={`${step === 'configuration' ? 'text-orange-600' : 'text-gray-400'}`}>2. 文案配置</span>
        <ArrowRight size={14} className="text-gray-300" />
        <span className={`${step === 'result' ? 'text-orange-600' : 'text-gray-400'}`}>3. 生成结果</span>
      </div>

      <AnimatePresence mode="wait">
        {step === 'competitors' && (
          <motion.div
            key="competitors"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            className="bg-white rounded-2xl shadow-lg p-8 space-y-6"
          >
            <div>
              <h3 className="text-xl font-bold mb-4">输入竞品链接</h3>
              <p className="text-gray-600 text-sm mb-6">输入Amazon竞品链接，AI将提取关键词和核心卖点</p>
            </div>

            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-2">任务名称（可选）</label>
              <input
                type="text"
                placeholder="例如：春季新品文案"
                value={taskName}
                onChange={(e) => setTaskName(e.target.value)}
                className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-orange-500 focus:border-transparent outline-none"
              />
            </div>

            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-2">选择AI模型</label>
              <div className="grid grid-cols-3 gap-3">
                <button
                  type="button"
                  onClick={() => setSelectedModel('gemini')}
                  className={`px-4 py-3 rounded-xl border-2 transition-all font-medium ${
                    selectedModel === 'gemini'
                      ? 'border-orange-500 bg-orange-50 text-orange-700'
                      : 'border-gray-200 hover:border-orange-300'
                  }`}
                >
                  Google Gemini
                </button>
                <button
                  type="button"
                  onClick={() => setSelectedModel('gpt')}
                  className={`px-4 py-3 rounded-xl border-2 transition-all font-medium ${
                    selectedModel === 'gpt'
                      ? 'border-orange-500 bg-orange-50 text-orange-700'
                      : 'border-gray-200 hover:border-orange-300'
                  }`}
                >
                  OpenAI GPT
                </button>
                <button
                  type="button"
                  onClick={() => setSelectedModel('deepseek')}
                  className={`px-4 py-3 rounded-xl border-2 transition-all font-medium ${
                    selectedModel === 'deepseek'
                      ? 'border-orange-500 bg-orange-50 text-orange-700'
                      : 'border-gray-200 hover:border-orange-300'
                  }`}
                >
                  DeepSeek
                </button>
              </div>
            </div>

            <div className="space-y-4">
              {competitorUrls.map((url, index) => (
                <div key={index} className="flex gap-3">
                  <input
                    type="text"
                    placeholder="https://www.amazon.com/dp/..."
                    value={url}
                    onChange={(e) => handleUrlChange(index, e.target.value)}
                    className="flex-1 px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-orange-500 focus:border-transparent outline-none"
                  />
                  <button
                    onClick={() => handleRemoveUrl(index)}
                    className="p-3 text-gray-400 hover:text-red-500 hover:bg-red-50 rounded-xl transition-colors"
                  >
                    <Trash2 size={20} />
                  </button>
                </div>
              ))}
              <button
                onClick={handleAddUrl}
                className="w-full py-3 border-2 border-dashed border-gray-200 rounded-xl text-gray-500 hover:text-orange-500 hover:border-orange-500 hover:bg-orange-50 transition-all flex items-center justify-center gap-2 font-medium"
              >
                <Plus size={18} />
                添加更多链接
              </button>
            </div>

            <button
              onClick={handleAnalyze}
              disabled={isAnalyzing || competitorUrls.every(u => !u.trim())}
              className="w-full py-4 bg-gradient-to-r from-orange-600 to-pink-600 text-white rounded-xl font-bold text-lg hover:from-orange-700 hover:to-pink-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all flex items-center justify-center gap-3 shadow-lg"
            >
              {isAnalyzing ? (
                <>
                  <Loader2 className="animate-spin" />
                  正在分析中...
                </>
              ) : (
                <>
                  开始分析竞品
                  <ArrowRight size={20} />
                </>
              )}
            </button>
          </motion.div>
        )}

        {step === 'configuration' && analysis && (
          <motion.div
            key="configuration"
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: -20 }}
            className="grid grid-cols-1 lg:grid-cols-2 gap-6"
          >
            {/* Analysis Results */}
            <div className="bg-white rounded-2xl shadow-lg p-6 space-y-6">
              <h3 className="text-xl font-bold mb-4">竞品分析结果</h3>
              
              {/* Keywords */}
              <div>
                <h4 className="text-sm font-bold text-gray-500 uppercase mb-3">核心关键词</h4>
                {(['core', 'attribute', 'extension'] as const).map(cat => {
                  const catKeywords = analysis.keywords.filter(k => k.category === cat);
                  if (catKeywords.length === 0) return null;
                  return (
                    <div key={cat} className="mb-4">
                      <div className="text-xs text-gray-400 mb-2">
                        {cat === 'core' ? '核心词' : cat === 'attribute' ? '属性词' : '拓展词'}
                      </div>
                      <div className="flex flex-wrap gap-2">
                        {catKeywords.map((item, i) => (
                          <button
                            key={i}
                            onClick={() => toggleKeyword(item.original)}
                            className={`px-3 py-1.5 rounded-lg text-xs font-medium transition-all border ${
                              selectedKeywords.includes(item.original)
                                ? 'bg-orange-500 text-white border-orange-500'
                                : 'bg-white text-gray-600 border-gray-200 hover:border-orange-500'
                            }`}
                          >
                            <div>{item.original}</div>
                            <div className="text-[9px] opacity-70">{item.translation}</div>
                          </button>
                        ))}
                      </div>
                    </div>
                  );
                })}
              </div>

              {/* Selling Points */}
              <div>
                <h4 className="text-sm font-bold text-gray-500 uppercase mb-3">产品卖点</h4>
                <div className="space-y-2">
                  {analysis.sellingPoints.map((item, i) => (
                    <button
                      key={i}
                      onClick={() => toggleSellingPoint(item.original)}
                      className={`w-full p-3 rounded-xl text-left transition-all border flex items-start gap-3 ${
                        selectedSellingPoints.includes(item.original)
                          ? 'bg-orange-50 border-orange-200'
                          : 'bg-white border-gray-200 hover:border-orange-500'
                      }`}
                    >
                      <div className={`mt-0.5 w-4 h-4 rounded-full border flex items-center justify-center ${
                        selectedSellingPoints.includes(item.original) ? 'bg-orange-500 border-orange-500' : 'border-gray-300'
                      }`}>
                        {selectedSellingPoints.includes(item.original) && <Check size={10} className="text-white" />}
                      </div>
                      <div className="flex-1">
                        <div className="font-bold text-sm">{item.original}</div>
                        <div className="text-xs text-gray-500">{item.translation}</div>
                      </div>
                    </button>
                  ))}
                </div>
              </div>
            </div>

            {/* Product Details Form */}
            <div className="bg-white rounded-2xl shadow-lg p-6 space-y-6">
              <h3 className="text-xl font-bold mb-4">产品信息</h3>
              
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-xs font-bold text-gray-500 uppercase mb-2 block">尺寸</label>
                  <input
                    type="text"
                    placeholder="10 x 5 x 2 inch"
                    value={productDetails.size}
                    onChange={(e) => setProductDetails({ ...productDetails, size: e.target.value })}
                    className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-orange-500 focus:border-transparent outline-none"
                  />
                </div>
                <div>
                  <label className="text-xs font-bold text-gray-500 uppercase mb-2 block">颜色</label>
                  <input
                    type="text"
                    placeholder="黑色"
                    value={productDetails.color}
                    onChange={(e) => setProductDetails({ ...productDetails, color: e.target.value })}
                    className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-orange-500 focus:border-transparent outline-none"
                  />
                </div>
                <div>
                  <label className="text-xs font-bold text-gray-500 uppercase mb-2 block">数量/规格</label>
                  <input
                    type="text"
                    placeholder="2件装"
                    value={productDetails.quantity}
                    onChange={(e) => setProductDetails({ ...productDetails, quantity: e.target.value })}
                    className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-orange-500 focus:border-transparent outline-none"
                  />
                </div>
                <div>
                  <label className="text-xs font-bold text-gray-500 uppercase mb-2 block">材质</label>
                  <input
                    type="text"
                    placeholder="铝合金"
                    value={productDetails.material}
                    onChange={(e) => setProductDetails({ ...productDetails, material: e.target.value })}
                    className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-orange-500 focus:border-transparent outline-none"
                  />
                </div>
                <div className="col-span-2">
                  <label className="text-xs font-bold text-gray-500 uppercase mb-2 block">主要功能</label>
                  <input
                    type="text"
                    placeholder="描述产品主要功能..."
                    value={productDetails.function}
                    onChange={(e) => setProductDetails({ ...productDetails, function: e.target.value })}
                    className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-orange-500 focus:border-transparent outline-none"
                  />
                </div>
                <div>
                  <label className="text-xs font-bold text-gray-500 uppercase mb-2 block">使用场景</label>
                  <input
                    type="text"
                    placeholder="办公室、户外"
                    value={productDetails.scenario}
                    onChange={(e) => setProductDetails({ ...productDetails, scenario: e.target.value })}
                    className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-orange-500 focus:border-transparent outline-none"
                  />
                </div>
                <div>
                  <label className="text-xs font-bold text-gray-500 uppercase mb-2 block">目标人群</label>
                  <input
                    type="text"
                    placeholder="专业人士"
                    value={productDetails.audience}
                    onChange={(e) => setProductDetails({ ...productDetails, audience: e.target.value })}
                    className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-orange-500 focus:border-transparent outline-none"
                  />
                </div>
                <div className="col-span-2">
                  <label className="text-xs font-bold text-gray-500 uppercase mb-2 block">核心关键词</label>
                  <textarea
                    rows={3}
                    placeholder="输入核心关键词，用逗号分隔..."
                    value={productDetails.keywords}
                    onChange={(e) => setProductDetails({ ...productDetails, keywords: e.target.value })}
                    className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-orange-500 focus:border-transparent outline-none resize-none"
                  />
                </div>
                <div className="col-span-2">
                  <label className="text-xs font-bold text-gray-500 uppercase mb-2 block">核心卖点</label>
                  <textarea
                    rows={4}
                    placeholder="输入核心卖点..."
                    value={productDetails.sellingPoints}
                    onChange={(e) => setProductDetails({ ...productDetails, sellingPoints: e.target.value })}
                    className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-orange-500 focus:border-transparent outline-none resize-none"
                  />
                </div>
              </div>
            </div>

            <div className="col-span-full flex gap-4">
              <button
                onClick={() => setStep('competitors')}
                className="flex-1 py-4 border border-gray-200 rounded-xl font-bold hover:bg-gray-50 transition-all flex items-center justify-center gap-2"
              >
                <ArrowLeft size={20} />
                返回
              </button>
              <button
                onClick={handleGenerate}
                disabled={isGenerating}
                className="flex-[2] py-4 bg-gradient-to-r from-orange-600 to-pink-600 text-white rounded-xl font-bold text-lg hover:from-orange-700 hover:to-pink-700 disabled:opacity-50 transition-all flex items-center justify-center gap-3 shadow-lg"
              >
                {isGenerating ? (
                  <>
                    <Loader2 className="animate-spin" />
                    正在生成文案...
                  </>
                ) : (
                  <>
                    生成文案
                    <Sparkles size={20} />
                  </>
                )}
              </button>
            </div>
          </motion.div>
        )}

        {step === 'result' && result && (
          <motion.div
            key="result"
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            className="space-y-6"
          >
            <div className="flex items-center justify-between">
              <h3 className="text-2xl font-bold">生成结果</h3>
              <div className="flex gap-3">
                <button
                  onClick={handleCopyAll}
                  className="px-4 py-2 bg-orange-500 text-white rounded-xl text-sm font-bold hover:bg-orange-600 transition-all flex items-center gap-2"
                >
                  {copiedField === 'all' ? <Check size={14} /> : <Copy size={14} />}
                  {copiedField === 'all' ? '已复制' : '复制全部'}
                </button>
                <button
                  onClick={() => setStep('configuration')}
                  className="px-4 py-2 border border-gray-200 rounded-xl text-sm font-bold hover:bg-gray-50 transition-all"
                >
                  重新调整
                </button>
              </div>
            </div>

            <div className="space-y-6">
              {/* Title */}
              <div className="bg-white border border-gray-200 rounded-2xl overflow-hidden">
                <div className="px-6 py-4 border-b border-gray-100 bg-gray-50 flex items-center justify-between">
                  <h4 className="text-xs font-bold uppercase text-gray-500">产品标题</h4>
                  <button 
                    onClick={() => copyToClipboard(result.title, 'title')}
                    className="text-gray-400 hover:text-orange-500 transition-colors flex items-center gap-1.5 text-xs font-bold"
                  >
                    {copiedField === 'title' ? <Check size={14} /> : <Copy size={14} />}
                    {copiedField === 'title' ? '已复制' : '复制'}
                  </button>
                </div>
                <div className="p-6">
                  <p className="text-lg font-bold">{result.title}</p>
                </div>
              </div>

              {/* Bullet Points */}
              <div className="bg-white border border-gray-200 rounded-2xl overflow-hidden">
                <div className="px-6 py-4 border-b border-gray-100 bg-gray-50 flex items-center justify-between">
                  <h4 className="text-xs font-bold uppercase text-gray-500">5点描述</h4>
                  <button 
                    onClick={() => copyToClipboard(result.bulletPoints.join('\n'), 'bullets')}
                    className="text-gray-400 hover:text-orange-500 transition-colors flex items-center gap-1.5 text-xs font-bold"
                  >
                    {copiedField === 'bullets' ? <Check size={14} /> : <Copy size={14} />}
                    {copiedField === 'bullets' ? '已复制' : '复制'}
                  </button>
                </div>
                <div className="p-6 space-y-4">
                  {result.bulletPoints.map((point, i) => (
                    <div key={i} className="flex gap-3">
                      <span className="text-orange-500 font-bold">•</span>
                      <p className="text-gray-700">{point}</p>
                    </div>
                  ))}
                </div>
              </div>

              {/* Description */}
              <div className="bg-white border border-gray-200 rounded-2xl overflow-hidden">
                <div className="px-6 py-4 border-b border-gray-100 bg-gray-50 flex items-center justify-between">
                  <h4 className="text-xs font-bold uppercase text-gray-500">产品描述</h4>
                  <button 
                    onClick={() => copyToClipboard(result.description, 'desc')}
                    className="text-gray-400 hover:text-orange-500 transition-colors flex items-center gap-1.5 text-xs font-bold"
                  >
                    {copiedField === 'desc' ? <Check size={14} /> : <Copy size={14} />}
                    {copiedField === 'desc' ? '已复制' : '复制'}
                  </button>
                </div>
                <div className="p-6">
                  <div className="prose prose-sm max-w-none text-gray-700">{result.description}</div>
                </div>
              </div>

              {/* Search Terms */}
              <div className="bg-white border border-gray-200 rounded-2xl overflow-hidden">
                <div className="px-6 py-4 border-b border-gray-100 bg-gray-50 flex items-center justify-between">
                  <h4 className="text-xs font-bold uppercase text-gray-500">搜索词</h4>
                  <button 
                    onClick={() => copyToClipboard(result.searchTerms, 'st')}
                    className="text-gray-400 hover:text-orange-500 transition-colors flex items-center gap-1.5 text-xs font-bold"
                  >
                    {copiedField === 'st' ? <Check size={14} /> : <Copy size={14} />}
                    {copiedField === 'st' ? '已复制' : '复制'}
                  </button>
                </div>
                <div className="p-6">
                  <p className="text-sm font-mono bg-gray-50 p-4 rounded-xl border border-gray-100 text-gray-600 break-all">
                    {result.searchTerms}
                  </p>
                </div>
              </div>
            </div>

            <button
              onClick={() => {
                setStep('competitors');
                setResult(null);
                setAnalysis(null);
                setTaskId(null);
              }}
              className="w-full py-4 bg-gradient-to-r from-orange-600 to-pink-600 text-white rounded-xl font-bold text-lg hover:from-orange-700 hover:to-pink-700 transition-all flex items-center justify-center gap-2"
            >
              开始新的生成
            </button>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
};
