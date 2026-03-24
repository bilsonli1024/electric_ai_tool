import React, { useState } from 'react';
import { apiClient } from '../services/api';
import { hashPassword, validateEmail, validatePassword } from '../utils/crypto';

interface AuthProps {
  onAuthSuccess: () => void;
}

type ViewMode = 'login' | 'register' | 'forgot' | 'reset';

export const Auth: React.FC<AuthProps> = ({ onAuthSuccess }) => {
  const [viewMode, setViewMode] = useState<ViewMode>('login');
  const [formData, setFormData] = useState({
    loginId: '',
    email: '',
    password: '',
    confirmPassword: '',
    resetToken: '',
    verificationCode: '',
  });
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loading, setLoading] = useState(false);
  const [codeSent, setCodeSent] = useState(false);
  const [countdown, setCountdown] = useState(0);

  // 检查URL参数，如果有reset token则自动进入重置密码模式
  React.useEffect(() => {
    const urlParams = new URLSearchParams(window.location.search);
    const token = urlParams.get('reset_token');
    if (token) {
      setViewMode('reset');
      setFormData(prev => ({ ...prev, resetToken: token }));
    }
  }, []);

  // 倒计时
  React.useEffect(() => {
    if (countdown > 0) {
      const timer = setTimeout(() => setCountdown(countdown - 1), 1000);
      return () => clearTimeout(timer);
    }
  }, [countdown]);

  const sendVerificationCode = async () => {
    if (!validateEmail(formData.email)) {
      setError('请输入有效的邮箱地址');
      return;
    }

    setError('');
    setLoading(true);
    try {
      await apiClient.sendVerificationCode(formData.email, 'register');
      setSuccess('验证码已发送到您的邮箱');
      setCodeSent(true);
      setCountdown(60);
    } catch (err: any) {
      setError(err.message || '发送验证码失败');
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess('');
    setLoading(true);

    try {
      if (viewMode === 'register') {
        if (!validateEmail(formData.email)) {
          throw new Error('请输入有效的邮箱地址');
        }

        const validation = validatePassword(formData.password);
        if (!validation.valid) {
          throw new Error(validation.message);
        }

        if (formData.password !== formData.confirmPassword) {
          throw new Error('两次输入的密码不一致');
        }

        if (!formData.verificationCode) {
          throw new Error('请输入验证码');
        }

        const passwordHash = await hashPassword(formData.password);
        await apiClient.register({
          email: formData.email,
          password_hash: passwordHash,
          verification_code: formData.verificationCode,
        });
        setSuccess('注册成功！请等待管理员审批后登录');
        setTimeout(() => {
          setViewMode('login');
        }, 2000);
      } else if (viewMode === 'login') {
        if (!formData.loginId || !formData.password) {
          throw new Error('请输入用户ID/邮箱和密码');
        }

        const passwordHash = await hashPassword(formData.password);
        await apiClient.login({
          login_id: formData.loginId,
          password_hash: passwordHash,
        });
        onAuthSuccess();
      } else if (viewMode === 'forgot') {
        if (!validateEmail(formData.email)) {
          throw new Error('请输入有效的邮箱地址');
        }

        await apiClient.forgotPassword(formData.email);
        setSuccess('重置链接已发送到您的邮箱（如果该邮箱存在）');
        setTimeout(() => setViewMode('reset'), 2000);
      } else if (viewMode === 'reset') {
        if (!formData.resetToken) {
          throw new Error('请输入重置令牌');
        }

        const validation = validatePassword(formData.password);
        if (!validation.valid) {
          throw new Error(validation.message);
        }

        if (formData.password !== formData.confirmPassword) {
          throw new Error('两次输入的密码不一致');
        }

        const passwordHash = await hashPassword(formData.password);
        await apiClient.resetPassword(formData.resetToken, passwordHash);
        setSuccess('密码重置成功，请登录');
        setTimeout(() => setViewMode('login'), 2000);
      }
    } catch (err: any) {
      setError(err.message || '操作失败');
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    });
  };

  const getTitle = () => {
    switch (viewMode) {
      case 'register': return '注册账号';
      case 'forgot': return '忘记密码';
      case 'reset': return '重置密码';
      default: return '用户登录';
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-indigo-500 via-purple-500 to-pink-500">
      <div className="bg-white rounded-2xl shadow-2xl p-8 w-full max-w-md">
        <h2 className="text-3xl font-bold text-center mb-6 text-gray-800">
          {getTitle()}
        </h2>

        <form onSubmit={handleSubmit} className="space-y-4">
          {viewMode === 'login' && (
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                用户ID / 邮箱
              </label>
              <input
                type="text"
                name="loginId"
                value={formData.loginId}
                onChange={handleChange}
                required
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                placeholder="请输入用户ID或邮箱"
              />
            </div>
          )}

          {(viewMode === 'register' || viewMode === 'forgot') && (
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                邮箱 *
              </label>
              <input
                type="email"
                name="email"
                value={formData.email}
                onChange={handleChange}
                required
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                placeholder="请输入邮箱"
              />
            </div>
          )}

          {viewMode === 'register' && (
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                验证码 *
              </label>
              <div className="flex gap-2">
                <input
                  type="text"
                  name="verificationCode"
                  value={formData.verificationCode}
                  onChange={handleChange}
                  required
                  maxLength={6}
                  className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  placeholder="请输入6位验证码"
                />
                <button
                  type="button"
                  onClick={sendVerificationCode}
                  disabled={countdown > 0 || loading}
                  className="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:bg-gray-400 disabled:cursor-not-allowed whitespace-nowrap"
                >
                  {countdown > 0 ? `${countdown}秒后重试` : '发送验证码'}
                </button>
              </div>
            </div>
          )}

          {viewMode === 'reset' && (
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                重置令牌
              </label>
              <input
                type="text"
                name="resetToken"
                value={formData.resetToken}
                onChange={handleChange}
                required
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                placeholder="请输入邮件中的令牌"
              />
            </div>
          )}

          {(viewMode === 'login' || viewMode === 'register' || viewMode === 'reset') && (
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                密码 {viewMode === 'register' || viewMode === 'reset' ? '*' : ''}
              </label>
              <input
                type="password"
                name="password"
                value={formData.password}
                onChange={handleChange}
                required
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                placeholder={viewMode === 'register' || viewMode === 'reset' ? '至少8位，包含大小写字母和数字' : '请输入密码'}
              />
            </div>
          )}

          {(viewMode === 'register' || viewMode === 'reset') && (
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                确认密码
              </label>
              <input
                type="password"
                name="confirmPassword"
                value={formData.confirmPassword}
                onChange={handleChange}
                required
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                placeholder="请再次输入密码"
              />
            </div>
          )}

          {error && (
            <div className="text-red-500 text-sm text-center bg-red-50 p-2 rounded">
              {error}
            </div>
          )}

          {success && (
            <div className="text-green-600 text-sm text-center bg-green-50 p-2 rounded">
              {success}
            </div>
          )}

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-indigo-600 text-white py-2 rounded-lg hover:bg-indigo-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed font-medium"
          >
            {loading ? '处理中...' : viewMode === 'login' ? '登录' : viewMode === 'register' ? '注册' : viewMode === 'forgot' ? '发送重置邮件' : '重置密码'}
          </button>
        </form>

        <div className="mt-6 space-y-2">
          {viewMode === 'login' && (
            <>
              <div className="text-center">
                <button
                  onClick={() => {
                    setViewMode('register');
                    setError('');
                    setSuccess('');
                  }}
                  className="text-indigo-600 hover:text-indigo-800 text-sm"
                >
                  没有账号？立即注册
                </button>
              </div>
              <div className="text-center">
                <button
                  onClick={() => {
                    setViewMode('forgot');
                    setError('');
                    setSuccess('');
                  }}
                  className="text-gray-600 hover:text-gray-800 text-sm"
                >
                  忘记密码？
                </button>
              </div>
            </>
          )}

          {viewMode === 'register' && (
            <div className="text-center">
              <button
                onClick={() => {
                  setViewMode('login');
                  setError('');
                  setSuccess('');
                }}
                className="text-indigo-600 hover:text-indigo-800 text-sm"
              >
                已有账号？立即登录
              </button>
            </div>
          )}

          {viewMode === 'forgot' && (
            <div className="text-center">
              <button
                onClick={() => {
                  setViewMode('reset');
                  setError('');
                  setSuccess('');
                }}
                className="text-gray-600 hover:text-gray-800 text-sm"
              >
                已有重置令牌？直接重置
              </button>
            </div>
          )}

          {(viewMode === 'forgot' || viewMode === 'reset') && (
            <div className="text-center">
              <button
                onClick={() => {
                  setViewMode('login');
                  setError('');
                  setSuccess('');
                }}
                className="text-indigo-600 hover:text-indigo-800 text-sm"
              >
                返回登录
              </button>
            </div>
          )}
        </div>

        {viewMode === 'register' && (
          <div className="mt-4 text-xs text-gray-500 bg-gray-50 p-3 rounded">
            <p className="font-medium mb-1">密码要求：</p>
            <ul className="list-disc list-inside space-y-0.5">
              <li>至少8位字符</li>
              <li>包含大写字母</li>
              <li>包含小写字母</li>
              <li>包含数字</li>
            </ul>
          </div>
        )}

        {viewMode === 'forgot' && (
          <div className="mt-4 text-xs text-gray-500 bg-blue-50 p-3 rounded border border-blue-200">
            <p className="font-medium mb-1 text-blue-800">提示：</p>
            <p className="text-blue-700">
              开发环境下，重置令牌会打印在服务器日志中。生产环境将通过邮件发送。
            </p>
          </div>
        )}

        {viewMode === 'reset' && (
          <div className="mt-4 text-xs text-gray-500 bg-yellow-50 p-3 rounded border border-yellow-200">
            <p className="font-medium mb-1 text-yellow-800">说明：</p>
            <ul className="list-disc list-inside space-y-0.5 text-yellow-700">
              <li>令牌有效期为1小时</li>
              <li>每个令牌只能使用一次</li>
              <li>如果没收到邮件，请检查服务器日志</li>
            </ul>
          </div>
        )}
      </div>
    </div>
  );
};
