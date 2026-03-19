import { StrictMode, useEffect, useState } from 'react';
import { createRoot } from 'react-dom/client';
import App from './App.tsx';
import { MainApp } from './MainApp.tsx';
import './index.css';

function RootApp() {
  const [showOriginal, setShowOriginal] = useState(false);

  useEffect(() => {
    const container = document.getElementById('original-app-container');
    if (container && !showOriginal) {
      setShowOriginal(true);
      const originalRoot = createRoot(container);
      originalRoot.render(<App />);
    }
  }, []);

  return <MainApp />;
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <RootApp />
  </StrictMode>,
);
