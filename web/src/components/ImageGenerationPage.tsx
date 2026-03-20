import React, { useState } from 'react';
import { Upload, Tag, Hash, Sparkles, Loader2 } from 'lucide-react';
import { CopywritingSelector } from './CopywritingSelector';

export const ImageGenerationPage: React.FC = () => {
  const [sku, setSku] = useState('');
  const [keywords, setKeywords] = useState('');
  const [sellingPoints, setSellingPoints] = useState('');
  const [competitorLink, setCompetitorLink] = useState('');
  const [selectedCopywritingTaskId, setSelectedCopywritingTaskId] = useState<number | null>(null);

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

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
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
        <div className="lg:col-span-2 space-y-6">
          {/* Copywriting Selector */}
          <CopywritingSelector onSelectCopywriting={handleCopywritingSelect} />

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

            <div className="space-y-4">
              <label className="text-sm font-semibold">产品白底图 (可多选)</label>
              <div 
                className="border-2 border-dashed border-gray-200 rounded-2xl p-8 text-center hover:border-purple-500 hover:bg-purple-50/50 transition-all cursor-pointer group"
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
            </div>

            <button 
              className="w-full py-4 bg-gradient-to-r from-purple-600 to-indigo-600 text-white rounded-xl font-bold text-lg hover:from-purple-700 hover:to-indigo-700 transition-all flex items-center justify-center gap-3 shadow-lg"
            >
              开始生成图片
              <Sparkles size={20} />
            </button>
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
