import React, { useEffect, useState } from 'react';
import Toast from './Toast';

interface ToastItem {
  id: number;
  message: string;
  type: 'success' | 'error' | 'warning' | 'info';
}

const ErrorToastContainer: React.FC = () => {
  const [toasts, setToasts] = useState<ToastItem[]>([]);

  useEffect(() => {
    const handleApiError = (event: Event) => {
      const customEvent = event as CustomEvent<{ message: string; type?: 'success' | 'error' | 'warning' | 'info' }>;
      const { message, type = 'error' } = customEvent.detail;
      
      const newToast: ToastItem = {
        id: Date.now(),
        message,
        type,
      };
      
      setToasts(prev => [...prev, newToast]);
    };

    window.addEventListener('api-error', handleApiError);

    return () => {
      window.removeEventListener('api-error', handleApiError);
    };
  }, []);

  const handleClose = (id: number) => {
    setToasts(prev => prev.filter(toast => toast.id !== id));
  };

  return (
    <div className="fixed top-0 right-0 z-[9999] p-4 space-y-2">
      {toasts.map(toast => (
        <Toast
          key={toast.id}
          message={toast.message}
          type={toast.type}
          onClose={() => handleClose(toast.id)}
        />
      ))}
    </div>
  );
};

export default ErrorToastContainer;
