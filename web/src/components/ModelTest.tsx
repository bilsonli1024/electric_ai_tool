import React, { useState } from 'react';
import { apiClient } from '../services/api';

interface ModelTestResult {
  success: boolean;
  model: string;
  response?: string;
  error?: string;
  response_time_ms: number;
}

interface EmailTestResult {
  success: boolean;
  message?: string;
  code?: string;
  email?: string;
  error?: string;
}

export const ModelTest: React.FC = () => {
  const [testResults, setTestResults] = useState<{ [key: string]: ModelTestResult | null }>({
    gemini: null,
    gpt: null,
    deepseek: null,
  });
  const [testPrompts, setTestPrompts] = useState<{ [key: string]: string }>({
    gemini: '请用中文回复"连接成功"',
    gpt: 'Please respond with "Connection successful"',
    deepseek: '请用中文回复"连接成功"',
  });
  const [loading, setLoading] = useState<{ [key: string]: boolean }>({
    gemini: false,
    gpt: false,
    deepseek: false,
    email: false,
  });
  const [testingAll, setTestingAll] = useState(false);
  const [emailTestResult, setEmailTestResult] = useState<EmailTestResult | null>(null);
  const [testEmail, setTestEmail] = useState('');

  const testModel = async (model: string) => {
    setLoading({ ...loading, [model]: true });
    setTestResults({ ...testResults, [model]: null });

    try {
      const result = await apiClient.testModel(model, testPrompts[model]);
      setTestResults({ ...testResults, [model]: result });
    } catch (err: any) {
      setTestResults({
        ...testResults,
        [model]: {
          success: false,
          model,
          error: err.message,
          response_time_ms: 0,
        },
      });
    } finally {
      setLoading({ ...loading, [model]: false });
    }
  };

  const testAllModels = async () => {
    setTestingAll(true);
    setTestResults({
      gemini: null,
      gpt: null,
      deepseek: null,
    });

    try {
      const response = await apiClient.testAllModels('请测试连接');
      const resultsMap: { [key: string]: ModelTestResult } = {};
      response.results.forEach((result: ModelTestResult) => {
        resultsMap[result.model] = result;
      });
      setTestResults(resultsMap);
    } catch (err: any) {
      console.error('Test all models failed:', err);
    } finally {
      setTestingAll(false);
    }
  };

  const testEmailVerification = async () => {
    setLoading({ ...loading, email: true });
    setEmailTestResult(null);

    try {
      const response = await fetch('http://43.160.241.164:4002/api/auth/test-send-verification-code', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email: testEmail }),
      });

      const data = await response.json();
      
      if (response.ok) {
        setEmailTestResult({
          success: true,
          message: data.message,
          code: data.code,
          email: data.email,
        });
      } else {
        setEmailTestResult({
          success: false,
          error: data.error || '测试失败',
        });
      }
    } catch (err: any) {
      setEmailTestResult({
        success: false,
        error: err.message || '网络请求失败',
      });
    } finally {
      setLoading({ ...loading, email: false });
    }
  };

  const getStatusIcon = (result: ModelTestResult | null) => {
    if (!result) return null;
    if (result.success) {
      return (
        <svg className="w-6 h-6 text-green-500" fill="currentColor" viewBox="0 0 20 20">
          <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
        </svg>
      );
    }
    return (
      <svg className="w-6 h-6 text-red-500" fill="currentColor" viewBox="0 0 20 20">
        <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
      </svg>
    );
  };

  const modelConfigs = [
    {
      key: 'gemini',
      name: 'Google Gemini',
      description: 'Google的多模态AI模型，支持文本分析和图片生成',
      color: 'from-blue-500 to-indigo-500',
    },
    {
      key: 'gpt',
      name: 'OpenAI GPT',
      description: 'OpenAI的语言模型，支持文本分析',
      color: 'from-green-500 to-teal-500',
    },
    {
      key: 'deepseek',
      name: 'DeepSeek',
      description: '深度求索AI模型，支持文本分析',
      color: 'from-purple-500 to-pink-500',
    },
  ];

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <div className="mb-8">
        <h2 className="text-3xl font-bold text-gray-800 mb-2">联通性测试</h2>
        <p className="text-gray-600">测试AI模型和邮件服务的连接状态和响应速度</p>
      </div>

      {/* 邮箱验证码测试卡片 */}
      <div className="mb-8 bg-white rounded-xl shadow-lg overflow-hidden">
        <div className="h-2 bg-gradient-to-r from-orange-500 to-red-500"></div>
        
        <div className="p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-xl font-bold text-gray-800">📧 邮箱验证码测试</h3>
            {emailTestResult && (
              emailTestResult.success ? (
                <svg className="w-6 h-6 text-green-500" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                </svg>
              ) : (
                <svg className="w-6 h-6 text-red-500" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                </svg>
              )
            )}
          </div>

          <p className="text-sm text-gray-600 mb-4">测试邮箱验证码生成和发送功能（验证码将显示在后端日志）</p>

          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              测试邮箱地址
            </label>
            <input
              type="email"
              value={testEmail}
              onChange={(e) => setTestEmail(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-orange-500 focus:border-transparent text-sm"
              placeholder="输入邮箱地址，如 test@example.com"
            />
          </div>

          <button
            onClick={testEmailVerification}
            disabled={loading.email || !testEmail}
            className="w-full px-4 py-2 bg-gradient-to-r from-orange-500 to-red-500 text-white rounded-lg hover:opacity-90 transition-opacity disabled:opacity-50 disabled:cursor-not-allowed font-medium"
          >
            {loading.email ? (
              <span className="flex items-center justify-center">
                <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                测试中...
              </span>
            ) : (
              '测试邮箱验证码'
            )}
          </button>

          {emailTestResult && (
            <div className={`mt-4 p-4 rounded-lg ${emailTestResult.success ? 'bg-green-50 border border-green-200' : 'bg-red-50 border border-red-200'}`}>
              <div className="mb-2">
                <span className={`text-sm font-medium ${emailTestResult.success ? 'text-green-800' : 'text-red-800'}`}>
                  {emailTestResult.success ? '测试成功' : '测试失败'}
                </span>
              </div>

              {emailTestResult.success && (
                <div className="space-y-2">
                  <div className="text-sm text-green-700 bg-white p-2 rounded border border-green-100">
                    <strong>邮箱:</strong> {emailTestResult.email}
                  </div>
                  <div className="text-sm text-green-700 bg-white p-2 rounded border border-green-100">
                    <strong>验证码:</strong> <span className="font-mono text-lg font-bold">{emailTestResult.code}</span>
                  </div>
                  <div className="text-xs text-green-600 mt-2">
                    💡 提示：请查看后端服务器日志获取完整的验证码发送信息
                  </div>
                </div>
              )}

              {!emailTestResult.success && emailTestResult.error && (
                <div className="text-sm text-red-700 mt-2">
                  <strong>错误:</strong> {emailTestResult.error}
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      <h3 className="text-2xl font-bold text-gray-800 mb-4">AI模型测试</h3>

      <div className="mb-6">
        <button
          onClick={testAllModels}
          disabled={testingAll}
          className="px-6 py-3 bg-gradient-to-r from-indigo-600 to-purple-600 text-white rounded-lg hover:from-indigo-700 hover:to-purple-700 transition-all disabled:opacity-50 disabled:cursor-not-allowed font-medium shadow-lg"
        >
          {testingAll ? (
            <span className="flex items-center">
              <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              测试所有模型中...
            </span>
          ) : (
            '一键测试所有AI模型'
          )}
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {modelConfigs.map((config) => {
          const result = testResults[config.key];
          const isLoading = loading[config.key];

          return (
            <div key={config.key} className="bg-white rounded-xl shadow-lg overflow-hidden">
              <div className={`h-2 bg-gradient-to-r ${config.color}`}></div>
              
              <div className="p-6">
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-xl font-bold text-gray-800">{config.name}</h3>
                  {result && getStatusIcon(result)}
                </div>

                <p className="text-sm text-gray-600 mb-4">{config.description}</p>

                <div className="mb-4">
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    测试提示词
                  </label>
                  <textarea
                    value={testPrompts[config.key]}
                    onChange={(e) =>
                      setTestPrompts({ ...testPrompts, [config.key]: e.target.value })
                    }
                    rows={3}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-sm"
                    placeholder="输入测试提示词"
                  />
                </div>

                <button
                  onClick={() => testModel(config.key)}
                  disabled={isLoading}
                  className={`w-full px-4 py-2 bg-gradient-to-r ${config.color} text-white rounded-lg hover:opacity-90 transition-opacity disabled:opacity-50 disabled:cursor-not-allowed font-medium`}
                >
                  {isLoading ? (
                    <span className="flex items-center justify-center">
                      <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" fill="none" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                      </svg>
                      测试中...
                    </span>
                  ) : (
                    '测试连通性'
                  )}
                </button>

                {result && (
                  <div className={`mt-4 p-4 rounded-lg ${result.success ? 'bg-green-50 border border-green-200' : 'bg-red-50 border border-red-200'}`}>
                    <div className="flex items-center justify-between mb-2">
                      <span className={`text-sm font-medium ${result.success ? 'text-green-800' : 'text-red-800'}`}>
                        {result.success ? '连接成功' : '连接失败'}
                      </span>
                      <span className="text-xs text-gray-600">
                        {result.response_time_ms}ms
                      </span>
                    </div>

                    {result.success && result.response && (
                      <div className="text-sm text-green-700 bg-white p-2 rounded border border-green-100">
                        <strong>响应:</strong> {result.response}
                      </div>
                    )}

                    {!result.success && result.error && (
                      <div className="text-sm text-red-700 mt-2">
                        <strong>错误:</strong> {result.error}
                      </div>
                    )}
                  </div>
                )}
              </div>
            </div>
          );
        })}
      </div>

      <div className="mt-8 bg-blue-50 border-l-4 border-blue-500 p-4 rounded">
        <div className="flex">
          <div className="flex-shrink-0">
            <svg className="h-5 w-5 text-blue-500" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
            </svg>
          </div>
          <div className="ml-3">
            <p className="text-sm text-blue-700">
              <strong>提示:</strong> 如果某个模型连接失败，请检查.env文件中对应的API Key配置是否正确。
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};
