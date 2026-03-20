import React, { useState } from 'react';
import { Search, X, CheckCircle2, Loader2, FileText } from 'lucide-react';
import { apiClient } from '../services/api';

interface CopywritingTask {
  id: number;
  task_name: string;
  generated_copy: string;
  created_at: string;
}

interface Props {
  onSelectCopywriting?: (task: CopywritingTask) => void;
}

export const CopywritingSelector: React.FC<Props> = ({ onSelectCopywriting }) => {
  const [searchKeyword, setSearchKeyword] = useState('');
  const [isSearching, setIsSearching] = useState(false);
  const [searchResults, setSearchResults] = useState<CopywritingTask[]>([]);
  const [selectedTask, setSelectedTask] = useState<CopywritingTask | null>(null);

  const handleSearch = async () => {
    if (!searchKeyword.trim()) return;

    setIsSearching(true);
    try {
      const response = await apiClient.searchCopywritingTasks(searchKeyword, 10);
      setSearchResults(response.data);
    } catch (error: any) {
      alert('搜索失败: ' + error.message);
    } finally {
      setIsSearching(false);
    }
  };

  const handleSelect = (task: CopywritingTask) => {
    setSelectedTask(task);
    if (onSelectCopywriting) {
      onSelectCopywriting(task);
    }
  };

  const handleClear = () => {
    setSelectedTask(null);
    setSearchResults([]);
    setSearchKeyword('');
  };

  return (
    <div className="bg-blue-50 border border-blue-200 rounded-xl p-6">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-sm font-bold text-blue-900 flex items-center gap-2">
          <FileText size={16} />
          引用已生成的文案
        </h3>
        {selectedTask && (
          <button
            onClick={handleClear}
            className="text-xs text-blue-600 hover:text-blue-800 flex items-center gap-1"
          >
            <X size={14} />
            清除
          </button>
        )}
      </div>

      {!selectedTask ? (
        <>
          <div className="flex gap-2 mb-4">
            <input
              type="text"
              placeholder="搜索文案任务名称或内容..."
              value={searchKeyword}
              onChange={(e) => setSearchKeyword(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
              className="flex-1 px-4 py-2 border border-blue-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none text-sm"
            />
            <button
              onClick={handleSearch}
              disabled={isSearching}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-all flex items-center gap-2 text-sm"
            >
              {isSearching ? <Loader2 size={16} className="animate-spin" /> : <Search size={16} />}
              搜索
            </button>
          </div>

          {searchResults.length > 0 && (
            <div className="space-y-2 max-h-64 overflow-y-auto">
              {searchResults.map((task) => {
                let copy;
                try {
                  copy = JSON.parse(task.generated_copy);
                } catch {
                  copy = null;
                }

                return (
                  <button
                    key={task.id}
                    onClick={() => handleSelect(task)}
                    className="w-full text-left p-3 bg-white rounded-lg hover:bg-blue-50 border border-blue-100 transition-all"
                  >
                    <div className="font-medium text-sm text-blue-900 mb-1">{task.task_name}</div>
                    {copy && (
                      <div className="text-xs text-gray-600 line-clamp-2">{copy.title}</div>
                    )}
                    <div className="text-xs text-gray-400 mt-1">
                      {new Date(task.created_at).toLocaleDateString()}
                    </div>
                  </button>
                );
              })}
            </div>
          )}
        </>
      ) : (
        <div className="bg-white rounded-lg p-4 border border-blue-200">
          <div className="flex items-start gap-3">
            <CheckCircle2 size={20} className="text-green-500 flex-shrink-0 mt-1" />
            <div className="flex-1 min-w-0">
              <div className="font-medium text-sm text-blue-900 mb-1">{selectedTask.task_name}</div>
              {(() => {
                try {
                  const copy = JSON.parse(selectedTask.generated_copy);
                  return (
                    <div className="text-xs text-gray-600 space-y-2">
                      <div className="line-clamp-2">{copy.title}</div>
                      {copy.bulletPoints && copy.bulletPoints.length > 0 && (
                        <div className="text-xs text-gray-500">
                          • {copy.bulletPoints[0].substring(0, 50)}...
                        </div>
                      )}
                    </div>
                  );
                } catch {
                  return null;
                }
              })()}
            </div>
          </div>
        </div>
      )}

      <p className="text-xs text-blue-600 mt-3">
        * 引用文案后，标题和关键词将自动填充到下方表单
      </p>
    </div>
  );
};
