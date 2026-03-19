/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useRef, useEffect } from 'react';
import { 
  Upload, 
  Link as LinkIcon, 
  Tag, 
  CheckCircle2, 
  Download, 
  RefreshCw, 
  Image as ImageIcon,
  ChevronRight,
  ChevronLeft,
  Loader2,
  Trash2,
  Plus,
  Key,
  Hash,
  Edit,
  Monitor,
  Smartphone,
  Layout,
  Sparkles,
  ArrowRight,
  List,
  History,
  Clock,
  Type
} from 'lucide-react';
import { motion, AnimatePresence } from 'motion/react';
import { analyzeSellingPoints, generateAmazonImage, generateAPlusContent } from './services/gemini';
import confetti from 'canvas-confetti';
import JSZip from 'jszip';
import { saveAs } from 'file-saver';
import { ImageEditor } from './components/ImageEditor';

type Step = 'input' | 'selling-points' | 'generate' | 'aplus' | 'aplus-config';

interface Task {
  id: string;
  sku: string;
  keywords: string;
  sellingPoints: string;
  baseImages: string[];
  customImageText?: string;
  textTargetImageIds?: number[];
  isExclusiveText?: boolean;
  generationMode: 'single' | 'set';
  singleRefImage?: string | null;
  aplusTemplate?: string;
  aplusRefImages?: string[];
  images: GeneratedImage[];
  aplusModules: APlusModule[];
  status: 'idle' | 'running' | 'completed' | 'error';
  createdAt: number;
}

interface SellingPoint {
  title: string;
  description: string;
  title_cn: string;
  description_cn: string;
}

interface APlusModule {
  type: string;
  title: string;
  description: string;
  imagePrompt: string;
  url?: string;
  status: 'idle' | 'generating' | 'done' | 'error';
}

interface GeneratedImage {
  id: number;
  url: string;
  type: string;
  prompt: string;
  status: 'idle' | 'generating' | 'done' | 'error';
}

const AVAILABLE_IMAGE_TYPES = [
  { type: '场景图 (室内使用)', prompt: 'Lifestyle image of the product being used in a modern home setting, cinematic lighting, realistic, preserving original product texture and materials.' },
  { type: '场景图 (户外/特定环境)', prompt: 'Lifestyle image of the product in its natural environment, high quality, professional photography, maintain original product material details.' },
  { type: '场景图 (多角度展示)', prompt: 'Professional product photography from a dynamic angle in a stylish environment, premium feel, high fidelity to original product texture.' },
  { type: '场景图 (细节氛围)', prompt: 'Atmospheric shot of the product highlighting its design and aesthetic in a real-world context, photorealistic materials.' },
  { type: '细节图 (特写)', prompt: 'Close-up macro shot of the product showing high-quality materials and texture, professional studio lighting, exact material reproduction.' },
  { type: '功能图 (信息图)', prompt: 'Infographic style image showing product features, clean layout, modern design, realistic product representation.' },
  { type: '尺寸图', prompt: 'Product image with dimension lines and text showing size, professional and clear, maintaining product visual integrity.' },
];

export default function App() {
  const [hasKey, setHasKey] = useState<boolean | null>(null);
  const [step, setStep] = useState<Step>('input');
  const [loading, setLoading] = useState(false);
  
  useEffect(() => {
    const checkKey = async () => {
      if (typeof window !== 'undefined' && (window as any).aistudio) {
        try {
          const has = await (window as any).aistudio.hasSelectedApiKey();
          setHasKey(has);
        } catch (e) {
          setHasKey(true);
        }
      } else {
        setHasKey(true);
      }
    };
    checkKey();
  }, []);

  const handleAiError = (error: any) => {
    console.error('AI Error:', error);
    let errorMessage = '';
    
    if (typeof error === 'string') {
      errorMessage = error;
    } else if (error?.message) {
      errorMessage = error.message;
    } else if (error?.error?.message) {
      errorMessage = error.error.message;
    } else {
      errorMessage = JSON.stringify(error);
    }
    
    const lowerMessage = errorMessage.toLowerCase();
    if (lowerMessage.includes('leaked') || 
        lowerMessage.includes('permission_denied') || 
        lowerMessage.includes('api key') ||
        lowerMessage.includes('requested entity was not found')) {
      
      setHasKey(false);
      // Trigger key selection dialog
      handleSelectKey();
    } else {
      alert('处理失败: ' + (errorMessage || '未知错误'));
    }
  };

  const handleSelectKey = async () => {
    if ((window as any).aistudio) {
      try {
        await (window as any).aistudio.openSelectKey();
        setHasKey(true); // Assume success to mitigate race condition
      } catch (e) {
        console.error(e);
      }
    }
  };
  
  // Form State
  const [competitorLink, setCompetitorLink] = useState('');
  const [sku, setSku] = useState('');
  const [keywords, setKeywords] = useState('');
  const [userSellingPoints, setUserSellingPoints] = useState('');
  const [baseImages, setBaseImages] = useState<string[]>([]);
  const [aspectRatio, setAspectRatio] = useState<'1:1' | '4:5'>('1:1');
  const [customImageText, setCustomImageText] = useState('');
  const [textTargetImageIds, setTextTargetImageIds] = useState<number[]>([]);
  const [isExclusiveText, setIsExclusiveText] = useState(false);
  const [generationMode, setGenerationMode] = useState<'single' | 'set'>('set');
  const [selectedSingleImageId, setSelectedSingleImageId] = useState<number>(1);
  const [singleRefImage, setSingleRefImage] = useState<string | null>(null);
  const [aplusTemplate, setAplusTemplate] = useState<string>('standard');
  const [aplusRefImages, setAplusRefImages] = useState<string[]>([]);

  // AI State
  const [aiSellingPoints, setAiSellingPoints] = useState<SellingPoint[]>([]);
  const [selectedPoints, setSelectedPoints] = useState<number[]>([]);
  const [editingImageIndex, setEditingImageIndex] = useState<number | null>(null);
  const [regenerateDirection, setRegenerateDirection] = useState<{ id: number, text: string } | null>(null);
  const [aplusModules, setAplusModules] = useState<APlusModule[]>([]);
  const [aplusView, setAplusView] = useState<'desktop' | 'mobile'>('desktop');
  const [isGeneratingAPlus, setIsGeneratingAPlus] = useState(false);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [activeTaskId, setActiveTaskId] = useState<string | null>(null);
  const [showTaskList, setShowTaskList] = useState(false);
  const [hoveredBaseImage, setHoveredBaseImage] = useState<string | null>(null);
  const [mousePos, setMousePos] = useState({ x: 0, y: 0 });

  const [imageConfigs, setImageConfigs] = useState<{[key: string]: number}>({});

  const [images, setImages] = useState<GeneratedImage[]>(
    AVAILABLE_IMAGE_TYPES.map((t, i) => ({
      id: i + 1,
      url: '',
      type: t.type,
      prompt: t.prompt,
      status: 'idle'
    }))
  );

  const fileInputRef = useRef<HTMLInputElement>(null);
  const singleRefImageInputRef = useRef<HTMLInputElement>(null);
  const regenFileInputRef = useRef<HTMLInputElement>(null);
  const aplusRefImagesInputRef = useRef<HTMLInputElement>(null);

  const handleDeleteTask = (taskId: string, e: React.MouseEvent) => {
    e.stopPropagation();
    setTasks(prev => prev.filter(t => t.id !== taskId));
    if (activeTaskId === taskId) {
      setActiveTaskId(null);
      setStep('input');
    }
  };

  const handleDownloadTask = async (task: Task, e: React.MouseEvent) => {
    e.stopPropagation();
    const zip = new JSZip();
    const folder = zip.folder(`${task.sku || 'Product'}_Images`);
    
    task.images.forEach((img, i) => {
      if (img.url) {
        const base64 = img.url.split(',')[1];
        folder?.file(`image_${img.id}_${img.type.replace(/\s+/g, '_')}.jpg`, base64, { base64: true });
      }
    });

    if (task.aplusModules && task.aplusModules.length > 0) {
      const aplusFolder = zip.folder(`${task.sku || 'Product'}_APlus`);
      task.aplusModules.forEach((m, i) => {
        if (m.url) {
          const base64 = m.url.split(',')[1];
          aplusFolder?.file(`aplus_module_${i+1}.jpg`, base64, { base64: true });
        }
      });
    }
    
    const content = await zip.generateAsync({ type: 'blob' });
    saveAs(content, `${task.sku || 'Product'}_Task_Assets.zip`);
  };

  const handleImageUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (files) {
      Array.from(files).forEach((file: File) => {
        const reader = new FileReader();
        reader.onloadend = () => {
          setBaseImages(prev => [...prev, reader.result as string]);
        };
        reader.readAsDataURL(file);
      });
    }
  };

  const getPreparedImages = () => {
    const hasCustomConfig = (Object.values(imageConfigs) as number[]).some(count => count > 0);
    
    const configs = hasCustomConfig 
      ? imageConfigs 
      : AVAILABLE_IMAGE_TYPES.reduce((acc, t) => ({ ...acc, [t.type]: 1 }), {} as {[key: string]: number});

    const newImages: GeneratedImage[] = [];
    let idCounter = 1;

    // We want to maintain a consistent order based on AVAILABLE_IMAGE_TYPES
    AVAILABLE_IMAGE_TYPES.forEach((typeInfo) => {
      const count = configs[typeInfo.type] || 0;
      if (count > 0) {
        for (let i = 0; i < count; i++) {
          newImages.push({
            id: idCounter++,
            url: '',
            type: count > 1 ? `${typeInfo.type} ${i + 1}` : typeInfo.type,
            prompt: typeInfo.prompt,
            status: 'idle'
          });
        }
      }
    });

    return newImages;
  };

  useEffect(() => {
    if (generationMode === 'set') {
      setImages(getPreparedImages());
    } else {
      // In single mode, show all available types once for selection
      setImages(AVAILABLE_IMAGE_TYPES.map((t, i) => ({
        id: i + 1,
        url: '',
        type: t.type,
        prompt: t.prompt,
        status: 'idle'
      })));
    }
  }, [imageConfigs, generationMode]);

  const startAnalysis = async () => {
    if (!keywords || !userSellingPoints) {
      alert('请填写核心关键词和卖点信息');
      return;
    }
    setLoading(true);
    try {
      const points = await analyzeSellingPoints(keywords, userSellingPoints, competitorLink, sku);
      setAiSellingPoints(points);
      setSelectedPoints([0, 1, 2, 3, 4]);
      setStep('generate');
      generateAllImages();
    } catch (error: any) {
      handleAiError(error);
    } finally {
      setLoading(false);
    }
  };

  const togglePoint = (index: number) => {
    if (selectedPoints.includes(index)) {
      setSelectedPoints(selectedPoints.filter(i => i !== index));
    } else if (selectedPoints.length < 5) {
      setSelectedPoints([...selectedPoints, index]);
    }
  };

  const startGeneration = async () => {
    setStep('generate');
    generateAllImages();
  };

  const generateAllImages = async () => {
    if (baseImages.length === 0) {
      alert('请至少上传一张产品图');
      return;
    }

    // Determine which images to generate
    const imagesToGenerate = generationMode === 'set' 
      ? images.map(img => ({ ...img, status: 'generating' as const }))
      : images.filter(img => img.id === selectedSingleImageId).map(img => ({ ...img, status: 'generating' as const }));

    // Create a new task
    const newTask: Task = {
      id: Math.random().toString(36).substr(2, 9),
      sku: sku || '未命名产品',
      keywords,
      sellingPoints: userSellingPoints,
      baseImages: [...baseImages],
      customImageText,
      textTargetImageIds,
      isExclusiveText,
      generationMode,
      singleRefImage: generationMode === 'single' ? singleRefImage : null,
      aplusTemplate,
      aplusRefImages: [...aplusRefImages],
      images: imagesToGenerate,
      aplusModules: [],
      status: 'running',
      createdAt: Date.now()
    };

    setTasks(prev => [newTask, ...prev]);
    setActiveTaskId(newTask.id);
    setImages(newTask.images);

    const selectedText = selectedPoints.map(idx => aiSellingPoints[idx].title).join(', ');

    // Parallel execution of multiple generation tasks
    const generationPromises = newTask.images.map(async (img, index) => {
      try {
        let finalPrompt = `${img.prompt}. Product features: ${selectedText}. Keywords: ${keywords}`;
        
        // Add custom text if applicable
        if (customImageText && textTargetImageIds.includes(img.id)) {
          if (isExclusiveText) {
            finalPrompt = `${img.prompt}. IMPORTANT: ONLY include the following text in the image: "${customImageText}". Do not add any other text. Product features: ${selectedText}. Keywords: ${keywords}`;
          } else {
            finalPrompt = `${img.prompt}. IMPORTANT: Include the following text in the image: "${customImageText}". Ensure the text is clear, professional, and compliant with Amazon policies. Product features: ${selectedText}. Keywords: ${keywords}`;
          }
          finalPrompt += " Ensure all text is free of infringing or sensitive words and complies with US law and Amazon requirements.";
        }

        // Use singleRefImage if in single mode and it's provided
        const styleRef = (generationMode === 'single' && singleRefImage) ? singleRefImage : undefined;
        const url = await generateAmazonImage(finalPrompt, aspectRatio, baseImages, styleRef);
        
        setImages(prev => {
          const next = [...prev];
          if (next[index]) {
            next[index].url = url || '';
            next[index].status = url ? 'done' : 'error';
          }
          return next;
        });

        setTasks(prev => prev.map(t => t.id === newTask.id ? {
          ...t,
          images: t.images.map((ti, i) => i === index ? { ...ti, url: url || '', status: url ? 'done' : 'error' } : ti)
        } : t));

      } catch (error: any) {
        handleAiError(error);
        setImages(prev => {
          const next = [...prev];
          if (next[index]) next[index].status = 'error';
          return next;
        });
        setTasks(prev => prev.map(t => t.id === newTask.id ? {
          ...t,
          images: t.images.map((ti, i) => i === index ? { ...ti, status: 'error' } : ti)
        } : t));
      }
    });

    await Promise.all(generationPromises);
    
    setTasks(prev => prev.map(t => t.id === newTask.id ? { ...t, status: 'completed' } : t));

    if (newTask.images.every(img => img.status === 'done')) {
      confetti({
        particleCount: 100,
        spread: 70,
        origin: { y: 0.6 }
      });
    }
  };

  const regenerateImage = async (id: number, customDirection?: string, refImage?: string) => {
    const index = images.findIndex(img => img.id === id);
    if (index === -1) return;

    const updatedImages = [...images];
    updatedImages[index].status = 'generating';
    setImages([...updatedImages]);

    try {
      const selectedText = selectedPoints.map(idx => aiSellingPoints[idx].title).join(', ');
      const directionPrompt = customDirection ? ` Direction: ${customDirection}.` : '';
      let finalPrompt = `${updatedImages[index].prompt}.${directionPrompt} Product features: ${selectedText}. Keywords: ${keywords}`;
      
      // Add custom text if applicable
      if (customImageText && textTargetImageIds.includes(id)) {
        if (isExclusiveText) {
          finalPrompt = `${updatedImages[index].prompt}.${directionPrompt} IMPORTANT: ONLY include the following text in the image: "${customImageText}". Do not add any other text. Product features: ${selectedText}. Keywords: ${keywords}`;
        } else {
          finalPrompt = `${updatedImages[index].prompt}.${directionPrompt} IMPORTANT: Include the following text in the image: "${customImageText}". Ensure the text is clear, professional, and compliant with Amazon policies. Product features: ${selectedText}. Keywords: ${keywords}`;
        }
        finalPrompt += " Ensure all text is free of infringing or sensitive words and complies with US law and Amazon requirements.";
      }

      // Use the provided refImage as a style reference, and baseImages as product reference
      const url = await generateAmazonImage(finalPrompt, aspectRatio, baseImages, refImage);
      
      if (url) {
        updatedImages[index].url = url;
        updatedImages[index].status = 'done';
      } else {
        updatedImages[index].status = 'error';
      }
      } catch (error: any) {
      updatedImages[index].status = 'error';
      handleAiError(error);
    }
    setImages([...updatedImages]);
    setRegenerateDirection(null);
  };

  const downloadAll = async () => {
    const zip = new JSZip();
    const folderName = `${sku || 'Product'}AI前台图`;
    const folder = zip.folder(folderName);

    if (!folder) return;

    const downloadPromises = images.map(async (img, index) => {
      if (img.url) {
        // Handle data URL
        const response = await fetch(img.url);
        const blob = await response.blob();
        const fileName = `${img.type.replace(/\s+/g, '_')}_${index + 1}.jpg`;
        folder.file(fileName, blob);
      }
    });

    await Promise.all(downloadPromises);

    const content = await zip.generateAsync({ type: 'blob' });
    saveAs(content, `${folderName}.zip`);
  };

  const downloadSingle = (url: string, type: string) => {
    saveAs(url, `${type.replace(/\s+/g, '_')}.jpg`);
  };

  const handleGenerateAPlus = async () => {
    setIsGeneratingAPlus(true);
    setStep('aplus');
    try {
      const selectedText = selectedPoints.map(idx => aiSellingPoints[idx].title);
      const modules = await generateAPlusContent(keywords, selectedText, sku, aplusTemplate, aplusRefImages);
      
      const initialModules = modules.map((m: any) => ({
        ...m,
        status: 'generating'
      }));
      setAplusModules(initialModules);

      // Generate images for each module
      for (let i = 0; i < initialModules.length; i++) {
        try {
          // A+ Premium images are usually wider, but we'll use 1:1 or 4:5 for now as per our generator
          const url = await generateAmazonImage(initialModules[i].imagePrompt, '1:1', baseImages);
          setAplusModules(prev => {
            const next = [...prev];
            next[i].url = url || '';
            next[i].status = url ? 'done' : 'error';
            return next;
          });
        } catch (e) {
          setAplusModules(prev => {
            const next = [...prev];
            next[i].status = 'error';
            return next;
          });
        }
      }
    } catch (error) {
      handleAiError(error);
    } finally {
      setIsGeneratingAPlus(false);
    }
  };

  const handleSaveEditedImage = (newUrl: string) => {
    if (editingImageIndex === null) return;
    const updatedImages = [...images];
    updatedImages[editingImageIndex].url = newUrl;
    setImages(updatedImages);
    setEditingImageIndex(null);
  };

  const totalConfiguredImages = (Object.values(imageConfigs) as number[]).reduce((a, b) => a + b, 0);
  const displayCount = totalConfiguredImages > 0 ? totalConfiguredImages : 7;

  if (hasKey === null) {
    return (
      <div className="min-h-screen bg-[#F8F9FA] flex items-center justify-center">
        <Loader2 className="animate-spin text-orange-500" size={40} />
      </div>
    );
  }

  if (hasKey === false) {
    return (
      <div className="min-h-screen bg-[#F8F9FA] flex items-center justify-center p-6">
        <div className="bg-white p-8 rounded-2xl shadow-sm border border-gray-200 text-center max-w-md w-full space-y-6">
          <div className="w-16 h-16 bg-orange-100 rounded-full flex items-center justify-center mx-auto text-orange-500">
            <Key size={32} />
          </div>
          <div>
            <h2 className="text-2xl font-bold mb-2">需要配置 API Key</h2>
            <p className="text-gray-500 text-sm leading-relaxed">
              为了生成高质量的亚马逊产品图，本应用使用了高级图像生成模型 (Gemini 3.1 Flash Image)。这需要您提供自己的 Gemini API Key（需关联已启用计费的 Google Cloud 项目）。
            </p>
          </div>
          <a 
            href="https://ai.google.dev/gemini-api/docs/billing" 
            target="_blank" 
            rel="noreferrer"
            className="text-orange-500 text-sm hover:underline block"
          >
            了解计费详情
          </a>
          <button 
            onClick={handleSelectKey} 
            className="w-full bg-orange-500 hover:bg-orange-600 text-white font-bold py-3 rounded-xl transition-all"
          >
            选择 API Key
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#F8F9FA] text-[#1A1A1A] font-sans selection:bg-orange-100">
      {/* Header */}
      <header className="sticky top-0 z-50 bg-white/80 backdrop-blur-md border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 bg-orange-500 rounded-lg flex items-center justify-center text-white font-bold">A</div>
            <h1 className="text-xl font-semibold tracking-tight">Amazon Image Pro</h1>
          </div>
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-4 text-sm text-gray-500 mr-8">
              <span className={step === 'input' ? 'text-orange-600 font-medium' : ''}>1. 输入信息</span>
              <ChevronRight size={14} />
              <span className={step === 'selling-points' ? 'text-orange-600 font-medium' : ''}>2. 卖点确认</span>
              <ChevronRight size={14} />
              <span className={step === 'generate' ? 'text-orange-600 font-medium' : ''}>3. 生成图片</span>
            </div>
            <button 
              onClick={() => setShowTaskList(!showTaskList)}
              className="relative p-2 hover:bg-gray-100 rounded-full transition-colors"
              title="任务中心"
            >
              <History size={20} className="text-gray-600" />
              {tasks.some(t => t.status === 'running') && (
                <span className="absolute top-0 right-0 w-3 h-3 bg-orange-500 border-2 border-white rounded-full animate-pulse" />
              )}
            </button>
          </div>
        </div>
      </header>

      {/* Task List Sidebar Overlay */}
      <AnimatePresence>
        {showTaskList && (
          <>
            <motion.div 
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              onClick={() => setShowTaskList(false)}
              className="fixed inset-0 bg-black/20 backdrop-blur-sm z-[60]"
            />
            <motion.div 
              initial={{ x: '100%' }}
              animate={{ x: 0 }}
              exit={{ x: '100%' }}
              className="fixed right-0 top-0 bottom-0 w-80 bg-white shadow-2xl z-[70] border-l border-gray-200 flex flex-col"
            >
              <div className="p-6 border-b border-gray-100 flex items-center justify-between">
                <h3 className="font-bold text-lg flex items-center gap-2">
                  <History size={20} className="text-orange-500" />
                  任务中心
                </h3>
                <button onClick={() => setShowTaskList(false)} className="text-gray-400 hover:text-black">
                  <ChevronRight size={24} />
                </button>
              </div>
              <div className="flex-1 overflow-y-auto p-4 space-y-4">
                {tasks.length === 0 ? (
                  <div className="text-center py-20 text-gray-400 space-y-2">
                    <Clock size={40} className="mx-auto opacity-20" />
                    <p className="text-sm">暂无历史任务</p>
                  </div>
                ) : (
                  tasks.map(task => (
                    <div 
                      key={task.id}
                      onClick={() => {
                        setActiveTaskId(task.id);
                        setSku(task.sku);
                        setKeywords(task.keywords);
                        setUserSellingPoints(task.sellingPoints);
                        setBaseImages(task.baseImages);
                        setCustomImageText(task.customImageText || '');
                        setTextTargetImageIds(task.textTargetImageIds || []);
                        setIsExclusiveText(task.isExclusiveText || false);
                        setGenerationMode(task.generationMode || 'set');
                        setAplusTemplate(task.aplusTemplate || 'standard');
                        setAplusRefImages(task.aplusRefImages || []);
                        setImages(task.images);
                        setAplusModules(task.aplusModules);
                        setStep('generate');
                        setShowTaskList(false);
                      }}
                      className={`p-4 rounded-xl border transition-all cursor-pointer hover:shadow-md ${activeTaskId === task.id ? 'border-orange-500 bg-orange-50/30' : 'border-gray-100 bg-white'}`}
                    >
                      <div className="flex justify-between items-start mb-2">
                        <span className="font-bold text-sm truncate max-w-[150px]">{task.sku}</span>
                        <div className="flex items-center gap-1">
                          {task.status === 'completed' && (
                            <button 
                              onClick={(e) => handleDownloadTask(task, e)}
                              className="p-1.5 text-gray-400 hover:text-orange-500 hover:bg-orange-50 rounded-lg transition-all"
                              title="下载资产"
                            >
                              <Download size={14} />
                            </button>
                          )}
                          <button 
                            onClick={(e) => handleDeleteTask(task.id, e)}
                            className="p-1.5 text-gray-400 hover:text-red-500 hover:bg-red-50 rounded-lg transition-all"
                            title="删除任务"
                          >
                            <Trash2 size={14} />
                          </button>
                        </div>
                      </div>
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2 text-[10px] text-gray-400">
                          <Clock size={12} />
                          {new Date(task.createdAt).toLocaleString()}
                        </div>
                        <span className={`text-[10px] px-2 py-0.5 rounded-full font-bold uppercase ${
                          task.status === 'completed' ? 'bg-green-100 text-green-600' : 
                          task.status === 'running' ? 'bg-orange-100 text-orange-600 animate-pulse' : 
                          'bg-gray-100 text-gray-600'
                        }`}>
                          {task.status === 'completed' ? '已完成' : task.status === 'running' ? '进行中' : '等待中'}
                        </span>
                      </div>
                      <div className="mt-3 flex -space-x-2 overflow-hidden">
                        {task.images.slice(0, 4).map((img, i) => (
                          <div key={i} className="w-8 h-8 rounded-lg border-2 border-white bg-gray-100 overflow-hidden">
                            {img.url ? (
                              <img src={img.url} className="w-full h-full object-cover" />
                            ) : (
                              <div className="w-full h-full flex items-center justify-center">
                                <ImageIcon size={12} className="text-gray-300" />
                              </div>
                            )}
                          </div>
                        ))}
                        {task.images.length > 4 && (
                          <div className="w-8 h-8 rounded-lg border-2 border-white bg-gray-50 flex items-center justify-center text-[10px] text-gray-400 font-bold">
                            +{task.images.length - 4}
                          </div>
                        )}
                      </div>
                    </div>
                  ))
                )}
              </div>
              <div className="p-4 border-t border-gray-100">
                <button 
                  onClick={() => {
                    setStep('input');
                    setSku('');
                    setKeywords('');
                    setUserSellingPoints('');
                    setBaseImages([]);
                    setCustomImageText('');
                    setTextTargetImageIds([]);
                    setIsExclusiveText(false);
                    setGenerationMode('set');
                    setAplusTemplate('standard');
                    setAplusRefImages([]);
                    setImages([
                      { id: 1, url: '', type: '场景图 1 (室内使用)', prompt: 'Lifestyle image of the product being used in a modern home setting, cinematic lighting, realistic, preserving original product texture and materials.', status: 'idle' },
                      { id: 2, url: '', type: '场景图 2 (户外/特定环境)', prompt: 'Lifestyle image of the product in its natural environment, high quality, professional photography, maintain original product material details.', status: 'idle' },
                      { id: 3, url: '', type: '场景图 3 (多角度展示)', prompt: 'Professional product photography from a dynamic angle in a stylish environment, premium feel, high fidelity to original product texture.', status: 'idle' },
                      { id: 4, url: '', type: '场景图 4 (细节氛围)', prompt: 'Atmospheric shot of the product highlighting its design and aesthetic in a real-world context, photorealistic materials.', status: 'idle' },
                      { id: 5, url: '', type: '细节图 (特写)', prompt: 'Close-up macro shot of the product showing high-quality materials and texture, professional studio lighting, exact material reproduction.', status: 'idle' },
                      { id: 6, url: '', type: '功能图 (信息图)', prompt: 'Infographic style image showing product features, clean layout, modern design, realistic product representation.', status: 'idle' },
                      { id: 7, url: '', type: '尺寸图', prompt: 'Product image with dimension lines and text showing size, professional and clear, maintaining product visual integrity.', status: 'idle' },
                    ]);
                    setActiveTaskId(null);
                    setShowTaskList(false);
                  }}
                  className="w-full py-3 bg-black text-white rounded-xl font-bold flex items-center justify-center gap-2 hover:bg-gray-800 transition-all"
                >
                  <Plus size={18} />
                  新建任务
                </button>
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>

      <main className="max-w-4xl mx-auto px-6 py-12">
        <AnimatePresence mode="wait">
          {step === 'input' && (
            <motion.div 
              key="input"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
              className="space-y-8"
            >
              <div className="text-center space-y-2">
                <h2 className="text-3xl font-bold">开始您的产品视觉之旅</h2>
                <p className="text-gray-500">提供基本信息，AI 将为您打造专业级的亚马逊前台图片</p>
              </div>

              <div className="bg-white rounded-2xl shadow-sm border border-gray-200 p-8 space-y-6">
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  <div className="space-y-2">
                    <label className="text-sm font-semibold flex items-center gap-2">
                      <Hash size={16} className="text-gray-400" />
                      SKU
                    </label>
                    <input 
                      type="text" 
                      placeholder="例如: CHAIR-001-BLK"
                      className="w-full px-4 py-3 rounded-xl border border-gray-200 focus:ring-2 focus:ring-orange-500 focus:border-transparent outline-none transition-all"
                      value={sku}
                      onChange={(e) => setSku(e.target.value)}
                    />
                  </div>
                  <div className="space-y-2">
                    <label className="text-sm font-semibold flex items-center gap-2">
                      <LinkIcon size={16} className="text-gray-400" />
                      竞品链接 (可选)
                    </label>
                    <input 
                      type="text" 
                      placeholder="https://amazon.com/..."
                      className="w-full px-4 py-3 rounded-xl border border-gray-200 focus:ring-2 focus:ring-orange-500 focus:border-transparent outline-none transition-all"
                      value={competitorLink}
                      onChange={(e) => setCompetitorLink(e.target.value)}
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
                      className="w-full px-4 py-3 rounded-xl border border-gray-200 focus:ring-2 focus:ring-orange-500 focus:border-transparent outline-none transition-all"
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
                    className="w-full px-4 py-3 rounded-xl border border-gray-200 focus:ring-2 focus:ring-orange-500 focus:border-transparent outline-none transition-all resize-none"
                    value={userSellingPoints}
                    onChange={(e) => setUserSellingPoints(e.target.value)}
                  />
                </div>

                {/* Custom Image Text Section */}
                <div className="bg-gray-50 rounded-2xl p-6 border border-gray-100 space-y-4">
                  <div className="flex items-center justify-between">
                    <h3 className="text-sm font-bold flex items-center gap-2">
                      <Type size={16} className="text-orange-500" />
                      图片自定义文字 (可选)
                    </h3>
                    <div className="flex items-center gap-2">
                      <span className="text-xs text-gray-500">仅显示此文字</span>
                      <button 
                        onClick={() => setIsExclusiveText(!isExclusiveText)}
                        className={`w-10 h-5 rounded-full transition-all relative ${isExclusiveText ? 'bg-orange-500' : 'bg-gray-300'}`}
                      >
                        <div className={`absolute top-1 w-3 h-3 bg-white rounded-full transition-all ${isExclusiveText ? 'right-1' : 'left-1'}`} />
                      </button>
                    </div>
                  </div>
                  
                  <input 
                    type="text" 
                    placeholder="输入必须出现在图片中的文字内容..."
                    className="w-full px-4 py-3 rounded-xl border border-gray-200 focus:ring-2 focus:ring-orange-500 outline-none transition-all bg-white"
                    value={customImageText}
                    onChange={(e) => setCustomImageText(e.target.value)}
                  />

                  <div className="space-y-2">
                    <p className="text-[10px] font-bold text-gray-400 uppercase tracking-wider">选择应用此文字的图片类型</p>
                    <div className="flex flex-wrap gap-2">
                      {images.map(img => (
                        <button
                          key={img.id}
                          onClick={() => {
                            setTextTargetImageIds(prev => 
                              prev.includes(img.id) 
                                ? prev.filter(id => id !== img.id) 
                                : [...prev, img.id]
                            );
                          }}
                          className={`px-3 py-1.5 rounded-lg text-xs font-medium transition-all border ${
                            textTargetImageIds.includes(img.id)
                              ? 'bg-orange-500 border-orange-500 text-white'
                              : 'bg-white border-gray-200 text-gray-600 hover:border-orange-200'
                          }`}
                        >
                          {img.type}
                        </button>
                      ))}
                    </div>
                  </div>
                  <p className="text-[10px] text-gray-400 leading-tight">
                    * AI 将确保文字符合美国法律及亚马逊平台要求，不含侵权或敏感词汇。若不填写，AI 将自动匹配。
                  </p>
                </div>

                <div className="space-y-4">
                  <label className="text-sm font-semibold">产品白底图 (可多选)</label>
                  <div 
                    onClick={() => fileInputRef.current?.click()}
                    className="border-2 border-dashed border-gray-200 rounded-2xl p-8 text-center hover:border-orange-500 hover:bg-orange-50/50 transition-all cursor-pointer group"
                  >
                    <input 
                      type="file" 
                      ref={fileInputRef} 
                      className="hidden" 
                      accept="image/*" 
                      multiple
                      onChange={handleImageUpload} 
                    />
                    
                    {baseImages.length > 0 ? (
                      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                        {baseImages.map((img, idx) => (
                          <div 
                            key={idx} 
                            className="relative group/img"
                            onMouseEnter={(e) => {
                              setHoveredBaseImage(img);
                              setMousePos({ x: e.clientX, y: e.clientY });
                            }}
                            onMouseMove={(e) => {
                              setMousePos({ x: e.clientX, y: e.clientY });
                            }}
                            onMouseLeave={() => setHoveredBaseImage(null)}
                          >
                            <img src={img} alt={`Preview ${idx}`} className="aspect-square object-cover rounded-lg shadow-sm" />
                            <button 
                              onClick={(e) => { 
                                e.stopPropagation(); 
                                setBaseImages(prev => prev.filter((_, i) => i !== idx));
                                if (hoveredBaseImage === img) setHoveredBaseImage(null);
                              }}
                              className="absolute -top-2 -right-2 bg-red-500 text-white p-1 rounded-full shadow-lg opacity-0 group-hover/img:opacity-100 transition-opacity z-10"
                            >
                              <Trash2 size={12} />
                            </button>
                          </div>
                        ))}
                        <div className="aspect-square border-2 border-dashed border-gray-200 rounded-lg flex items-center justify-center text-gray-400 group-hover:border-orange-500 group-hover:text-orange-500 transition-all">
                          <Plus size={24} />
                        </div>
                      </div>
                    ) : (
                      <div className="space-y-4">
                        <div className="w-16 h-16 bg-orange-50 rounded-full flex items-center justify-center mx-auto group-hover:scale-110 transition-transform">
                          <Upload className="text-orange-500" size={24} />
                        </div>
                        <div>
                          <p className="font-bold text-gray-700">点击或拖拽上传产品图</p>
                          <p className="text-sm text-gray-500">支持多张上传，AI 将自动识别产品特征</p>
                        </div>
                      </div>
                    )}
                  </div>
                </div>

                {/* Image Type Configuration */}
                {generationMode === 'set' && (
                  <div className="bg-white rounded-2xl p-6 border border-gray-200 space-y-4">
                    <h3 className="text-sm font-bold flex items-center gap-2">
                      <List size={16} className="text-orange-500" />
                      图片类型配置 (可选)
                    </h3>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      {AVAILABLE_IMAGE_TYPES.map((typeInfo) => (
                        <div key={typeInfo.type} className="flex items-center justify-between p-3 rounded-xl border border-gray-100 bg-gray-50/50">
                          <span className="text-xs font-medium text-gray-700">{typeInfo.type}</span>
                          <div className="flex items-center gap-3">
                            <button 
                              onClick={() => {
                                const current = imageConfigs[typeInfo.type] || 0;
                                if (current > 0) {
                                  setImageConfigs({ ...imageConfigs, [typeInfo.type]: current - 1 });
                                }
                              }}
                              className="w-6 h-6 rounded-full bg-white border border-gray-200 flex items-center justify-center text-gray-500 hover:border-orange-500 hover:text-orange-500 transition-all"
                            >
                              -
                            </button>
                            <span className="text-sm font-bold w-4 text-center">{imageConfigs[typeInfo.type] || 0}</span>
                            <button 
                              onClick={() => {
                                const current = imageConfigs[typeInfo.type] || 0;
                                setImageConfigs({ ...imageConfigs, [typeInfo.type]: current + 1 });
                              }}
                              className="w-6 h-6 rounded-full bg-white border border-gray-200 flex items-center justify-center text-gray-500 hover:border-orange-500 hover:text-orange-500 transition-all"
                            >
                              +
                            </button>
                          </div>
                        </div>
                      ))}
                    </div>
                    <p className="text-[10px] text-gray-400">
                      * 若不填写，默认生成 7 张不同类型的图片。您可以增加或减少特定类型的张数。
                    </p>
                  </div>
                )}

                {/* Generation Mode Selection */}
                <div className="bg-white rounded-2xl p-6 border border-gray-200 space-y-4">
                  <h3 className="text-sm font-bold flex items-center gap-2">
                    <Sparkles size={16} className="text-orange-500" />
                    选择生成模式
                  </h3>
                  <div className="flex gap-4">
                    <button 
                      onClick={() => setGenerationMode('set')}
                      className={`flex-1 py-3 rounded-xl border-2 transition-all flex flex-col items-center gap-1 ${
                        generationMode === 'set' ? 'border-orange-500 bg-orange-50' : 'border-gray-100'
                      }`}
                    >
                      <span className="font-bold text-sm">全套生成</span>
                      <span className="text-[10px] text-gray-500 text-center">一次性生成 {displayCount} 张亚马逊前台图</span>
                    </button>
                    <button 
                      onClick={() => setGenerationMode('single')}
                      className={`flex-1 py-3 rounded-xl border-2 transition-all flex flex-col items-center gap-1 ${
                        generationMode === 'single' ? 'border-orange-500 bg-orange-50' : 'border-gray-100'
                      }`}
                    >
                      <span className="font-bold text-sm">单张生成</span>
                      <span className="text-[10px] text-gray-500 text-center">仅生成您指定类型的单张图片</span>
                    </button>
                  </div>

                  {generationMode === 'single' && (
                    <motion.div 
                      initial={{ opacity: 0, height: 0 }}
                      animate={{ opacity: 1, height: 'auto' }}
                      className="pt-2 space-y-2"
                    >
                      <p className="text-[10px] font-bold text-gray-400 uppercase tracking-wider">选择要生成的图片类型</p>
                      <div className="flex flex-wrap gap-2">
                        {images.map(img => (
                          <button
                            key={img.id}
                            onClick={() => setSelectedSingleImageId(img.id)}
                            className={`px-3 py-1.5 rounded-lg text-xs font-medium transition-all border ${
                              selectedSingleImageId === img.id
                                ? 'bg-orange-500 border-orange-500 text-white'
                                : 'bg-white border-gray-200 text-gray-600 hover:border-orange-200'
                            }`}
                          >
                            {img.type}
                          </button>
                        ))}
                      </div>

                      <div className="pt-2 space-y-2">
                        <p className="text-[10px] font-bold text-gray-400 uppercase tracking-wider">上传风格参考图 (可选)</p>
                        <div 
                          onClick={() => singleRefImageInputRef.current?.click()}
                          className="border-2 border-dashed border-gray-100 rounded-xl p-4 text-center hover:border-orange-500 hover:bg-orange-50/50 transition-all cursor-pointer group relative"
                        >
                          <input 
                            type="file" 
                            ref={singleRefImageInputRef} 
                            className="hidden" 
                            accept="image/*" 
                            onChange={(e) => {
                              const file = e.target.files?.[0];
                              if (file) {
                                const reader = new FileReader();
                                reader.onloadend = () => {
                                  setSingleRefImage(reader.result as string);
                                };
                                reader.readAsDataURL(file);
                              }
                            }} 
                          />
                          {singleRefImage ? (
                            <div className="relative inline-block">
                              <img src={singleRefImage} alt="Reference" className="h-20 w-20 object-cover rounded-lg shadow-sm" />
                              <button 
                                onClick={(e) => {
                                  e.stopPropagation();
                                  setSingleRefImage(null);
                                }}
                                className="absolute -top-2 -right-2 bg-red-500 text-white p-1 rounded-full shadow-lg"
                              >
                                <Trash2 size={10} />
                              </button>
                            </div>
                          ) : (
                            <div className="flex flex-col items-center gap-1">
                              <ImageIcon className="text-gray-300 group-hover:text-orange-500 transition-colors" size={20} />
                              <span className="text-[10px] text-gray-500">点击上传参考图，AI 将参考其构图与风格</span>
                            </div>
                          )}
                        </div>
                      </div>
                    </motion.div>
                  )}
                </div>

                <div className="pt-4">
                  <button 
                    onClick={startAnalysis}
                    disabled={loading}
                    className="w-full bg-orange-500 hover:bg-orange-600 text-white font-bold py-4 rounded-xl shadow-lg shadow-orange-500/20 transition-all flex items-center justify-center gap-2 disabled:opacity-50"
                  >
                    {loading ? <Loader2 className="animate-spin" /> : <ChevronRight />}
                    {generationMode === 'set' ? `一键生成 ${displayCount} 张前台图` : '生成单张前台图'}
                  </button>
                </div>
              </div>
            </motion.div>
          )}

          {step === 'selling-points' && (
            <motion.div 
              key="selling-points"
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -20 }}
              className="space-y-8"
            >
              <div className="text-center space-y-2">
                <h2 className="text-3xl font-bold">确认核心卖点</h2>
                <p className="text-gray-500">请从以下 9 个 AI 提炼的卖点中选择最多 5 个</p>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                {aiSellingPoints.map((point, idx) => (
                  <div 
                    key={idx}
                    onClick={() => togglePoint(idx)}
                    className={`p-6 rounded-2xl border-2 transition-all cursor-pointer relative group ${
                      selectedPoints.includes(idx) 
                        ? 'border-orange-500 bg-orange-50' 
                        : 'border-gray-100 bg-white hover:border-orange-200'
                    }`}
                  >
                    {selectedPoints.includes(idx) && (
                      <div className="absolute top-3 right-3 text-orange-500">
                        <CheckCircle2 size={20} />
                      </div>
                    )}
                    <h3 className="font-bold mb-1">{point.title_cn}</h3>
                    <p className="text-xs text-orange-600 font-medium mb-2 uppercase tracking-tight">{point.title}</p>
                    <p className="text-sm text-gray-500 leading-relaxed mb-2">{point.description_cn}</p>
                    <p className="text-[10px] text-gray-400 leading-tight italic">{point.description}</p>
                  </div>
                ))}
              </div>

              <div className="bg-white rounded-2xl p-8 border border-gray-200 space-y-6">
                <div className="space-y-4">
                  <h3 className="font-bold flex items-center gap-2">
                    <ImageIcon size={18} />
                    选择图片比例
                  </h3>
                  <div className="flex gap-4">
                    <button 
                      onClick={() => setAspectRatio('1:1')}
                      className={`flex-1 py-4 rounded-xl border-2 transition-all ${
                        aspectRatio === '1:1' ? 'border-orange-500 bg-orange-50' : 'border-gray-100'
                      }`}
                    >
                      <div className="font-bold">1:1 (1000x1000)</div>
                      <div className="text-xs text-gray-500">亚马逊标准正方形</div>
                    </button>
                    <button 
                      onClick={() => setAspectRatio('4:5')}
                      className={`flex-1 py-4 rounded-xl border-2 transition-all ${
                        aspectRatio === '4:5' ? 'border-orange-500 bg-orange-50' : 'border-gray-100'
                      }`}
                    >
                      <div className="font-bold">4:5 (1600x2000)</div>
                      <div className="text-xs text-gray-500">移动端优化长方形</div>
                    </button>
                  </div>
                </div>

                <div className="flex gap-4">
                  <button 
                    onClick={() => setStep('input')}
                    className="flex-1 py-4 rounded-xl border border-gray-200 font-bold hover:bg-gray-50 transition-all flex items-center justify-center gap-2"
                  >
                    <ChevronLeft size={18} />
                    返回修改
                  </button>
                  <button 
                    onClick={startGeneration}
                    disabled={selectedPoints.length === 0}
                    className="flex-[2] bg-orange-500 hover:bg-orange-600 text-white font-bold py-4 rounded-xl shadow-lg shadow-orange-500/20 transition-all flex items-center justify-center gap-2 disabled:opacity-50"
                  >
                    {generationMode === 'set' ? '生成 7 张前台图片' : '生成单张前台图片'}
                    <ChevronRight size={18} />
                  </button>
                </div>
              </div>
            </motion.div>
          )}

          {step === 'generate' && (
            <motion.div 
              key="generate"
              initial={{ opacity: 0, scale: 0.95 }}
              animate={{ opacity: 1, scale: 1 }}
              className="space-y-8"
            >
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-3xl font-bold">生成您的视觉资产</h2>
                  <p className="text-gray-500">AI 正在为您生成符合亚马逊标准的 {generationMode === 'set' ? '7 张图片' : '单张图片'}</p>
                </div>
                <button 
                  onClick={downloadAll}
                  disabled={images.some(img => img.status === 'generating')}
                  className="bg-black text-white px-6 py-3 rounded-xl font-bold flex items-center gap-2 hover:bg-gray-800 transition-all disabled:opacity-50"
                >
                  <Download size={18} />
                  一键下载全部 (JPG)
                </button>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {images.map((img) => (
                  <div key={img.id} className="bg-white rounded-2xl border border-gray-200 overflow-hidden group">
                    <div className={`aspect-[${aspectRatio === '1:1' ? '1/1' : '4/5'}] bg-gray-50 relative flex items-center justify-center overflow-hidden`}>
                      {img.status === 'generating' ? (
                        <div className="text-center space-y-4">
                          <Loader2 className="animate-spin text-orange-500 mx-auto" size={40} />
                          <p className="text-sm font-medium text-gray-400">正在生成 {img.type}...</p>
                        </div>
                      ) : img.url ? (
                        <img src={img.url} alt={img.type} className="w-full h-full object-cover" />
                      ) : (
                        <div className="text-gray-300">
                          <ImageIcon size={64} />
                        </div>
                      )}
                      
                      {img.status === 'done' && (
                        <div className="absolute inset-0 bg-black/60 opacity-0 group-hover:opacity-100 transition-opacity flex flex-col items-center justify-center gap-4 p-6">
                          <div className="w-full space-y-3">
                            <div className="relative">
                              <input 
                                type="text" 
                                placeholder="输入重新生成方向 (如: 换成木质背景)"
                                className="w-full bg-white/90 backdrop-blur pl-3 pr-10 py-2 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-orange-500"
                                onClick={(e) => e.stopPropagation()}
                                onKeyDown={(e) => {
                                  if (e.key === 'Enter') {
                                    regenerateImage(img.id, (e.target as HTMLInputElement).value);
                                  }
                                }}
                              />
                              <button 
                                onClick={(e) => {
                                  e.stopPropagation();
                                  regenFileInputRef.current?.click();
                                  // Store the current image ID to know which one we're uploading for
                                  (window as any).currentRegenId = img.id;
                                }}
                                className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-orange-500 transition-colors"
                                title="上传参考图"
                              >
                                <ImageIcon size={18} />
                              </button>
                              <input 
                                type="file"
                                ref={regenFileInputRef}
                                className="hidden"
                                accept="image/*"
                                onChange={(e) => {
                                  const file = e.target.files?.[0];
                                  if (file) {
                                    const reader = new FileReader();
                                    reader.onloadend = () => {
                                      const id = (window as any).currentRegenId;
                                      const input = e.target.parentElement?.querySelector('input[type="text"]') as HTMLInputElement;
                                      regenerateImage(id, input.value, reader.result as string);
                                    };
                                    reader.readAsDataURL(file);
                                  }
                                }}
                              />
                            </div>
                            <div className="flex gap-2 justify-center">
                              <button 
                                onClick={() => regenerateImage(img.id)}
                                className="bg-white text-black px-4 py-2 rounded-lg hover:scale-105 transition-transform flex items-center gap-2 font-bold text-sm"
                              >
                                <RefreshCw size={16} />
                                直接重试
                              </button>
                              <button 
                                onClick={() => downloadSingle(img.url, img.type)}
                                className="bg-white text-black px-4 py-2 rounded-lg hover:scale-105 transition-transform flex items-center gap-2 font-bold text-sm"
                              >
                                <Download size={16} />
                                下载
                              </button>
                              <button 
                                onClick={() => {
                                  const index = images.findIndex(i => i.id === img.id);
                                  setEditingImageIndex(index);
                                }}
                                className="bg-white text-black px-4 py-2 rounded-lg hover:scale-105 transition-transform flex items-center gap-2 font-bold text-sm"
                              >
                                <Edit size={16} />
                                编辑
                              </button>
                            </div>
                          </div>
                        </div>
                      )}
                    </div>
                    <div className="p-4 border-t border-gray-100 flex items-center justify-between">
                      <div>
                        <span className="text-xs font-bold text-orange-500 uppercase tracking-wider">Image {img.id}</span>
                        <h4 className="font-bold">{img.type}</h4>
                      </div>
                      {img.status === 'done' && (
                        <div className="bg-green-100 text-green-600 p-1 rounded-full">
                          <CheckCircle2 size={16} />
                        </div>
                      )}
                    </div>
                  </div>
                ))}
              </div>

              <div className="flex justify-center pt-8 gap-4">
                <button 
                  onClick={() => setStep('selling-points')}
                  className="text-gray-500 hover:text-black transition-colors flex items-center gap-2"
                >
                  <ChevronLeft size={18} />
                  返回修改卖点
                </button>
                <button 
                  onClick={() => setStep('aplus-config')}
                  className="bg-orange-100 text-orange-600 px-6 py-3 rounded-xl font-bold flex items-center gap-2 hover:bg-orange-200 transition-all"
                >
                  <Layout size={18} />
                  生成高级 A+ 页面
                  <ArrowRight size={18} />
                </button>
              </div>
            </motion.div>
          )}

          {step === 'aplus-config' && (
            <motion.div 
              key="aplus-config"
              initial={{ opacity: 0, scale: 0.95 }}
              animate={{ opacity: 1, scale: 1 }}
              className="max-w-3xl mx-auto space-y-8"
            >
              <div className="text-center space-y-2">
                <h2 className="text-3xl font-bold">A+ 页面配置</h2>
                <p className="text-gray-500">选择模板并提供参考图，AI 将为您策划专业级 A+ 内容</p>
              </div>

              <div className="bg-white rounded-3xl p-8 shadow-sm border border-gray-100 space-y-8">
                {/* Template Selection */}
                <div className="space-y-4">
                  <h3 className="text-lg font-bold flex items-center gap-2">
                    <Layout size={20} className="text-orange-500" />
                    选择 A+ 模板
                  </h3>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    {[
                      { id: 'standard', name: '标准全能型', desc: '5模块：品牌故事+核心功能+场景+细节+对比表', icon: <Layout size={20} /> },
                      { id: 'visual', name: '视觉冲击型', desc: '4模块：大图首图+三列功能+场景+细节放大', icon: <ImageIcon size={20} /> },
                      { id: 'technical', name: '技术参数型', desc: '6模块：横幅+爆炸图+材质+指南+安全+对比表', icon: <Hash size={20} /> },
                      { id: 'minimalist', name: '极简设计型', desc: '3模块：顶部图+场景网格+核心参数', icon: <Monitor size={20} /> }
                    ].map(t => (
                      <button
                        key={t.id}
                        onClick={() => setAplusTemplate(t.id)}
                        className={`p-4 rounded-2xl border-2 text-left transition-all flex gap-4 ${
                          aplusTemplate === t.id ? 'border-orange-500 bg-orange-50' : 'border-gray-100 hover:border-orange-200'
                        }`}
                      >
                        <div className={`w-12 h-12 rounded-xl flex items-center justify-center ${aplusTemplate === t.id ? 'bg-orange-500 text-white' : 'bg-gray-100 text-gray-400'}`}>
                          {t.icon}
                        </div>
                        <div className="flex-1">
                          <div className="font-bold text-sm">{t.name}</div>
                          <div className="text-[10px] text-gray-500 leading-tight mt-1">{t.desc}</div>
                        </div>
                      </button>
                    ))}
                  </div>
                </div>

                {/* Reference Images */}
                <div className="space-y-4">
                  <h3 className="text-lg font-bold flex items-center gap-2">
                    <History size={20} className="text-orange-500" />
                    竞品/参考 A+ 模板 (可选)
                  </h3>
                  <div 
                    onClick={() => aplusRefImagesInputRef.current?.click()}
                    className="border-2 border-dashed border-gray-200 rounded-2xl p-8 text-center hover:border-orange-500 hover:bg-orange-50/50 transition-all cursor-pointer group"
                  >
                    <input 
                      type="file" 
                      ref={aplusRefImagesInputRef} 
                      className="hidden" 
                      accept="image/*" 
                      multiple
                      onChange={(e) => {
                        const files = e.target.files;
                        if (files) {
                          Array.from(files).forEach((file: File) => {
                            const reader = new FileReader();
                            reader.onloadend = () => {
                              setAplusRefImages(prev => [...prev, reader.result as string]);
                            };
                            reader.readAsDataURL(file);
                          });
                        }
                      }} 
                    />
                    
                    {aplusRefImages.length > 0 ? (
                      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                        {aplusRefImages.map((img, idx) => (
                          <div key={idx} className="relative group/img">
                            <img src={img} alt={`Ref ${idx}`} className="aspect-square object-cover rounded-lg shadow-sm" />
                            <button 
                              onClick={(e) => { 
                                e.stopPropagation(); 
                                setAplusRefImages(prev => prev.filter((_, i) => i !== idx));
                              }}
                              className="absolute -top-2 -right-2 bg-red-500 text-white p-1 rounded-full shadow-lg opacity-0 group-hover/img:opacity-100 transition-opacity"
                            >
                              <Trash2 size={12} />
                            </button>
                          </div>
                        ))}
                        <div className="aspect-square border-2 border-dashed border-gray-200 rounded-lg flex items-center justify-center text-gray-400 group-hover:border-orange-500 group-hover:text-orange-500 transition-all">
                          <Plus size={24} />
                        </div>
                      </div>
                    ) : (
                      <div className="space-y-2">
                        <div className="w-12 h-12 bg-gray-100 rounded-xl flex items-center justify-center mx-auto text-gray-400 group-hover:scale-110 transition-transform">
                          <Plus size={24} />
                        </div>
                        <p className="text-sm font-medium text-gray-600">点击上传竞品 A+ 截图</p>
                        <p className="text-xs text-gray-400">AI 将参考其视觉风格和排版逻辑</p>
                      </div>
                    )}
                  </div>
                </div>

                <div className="flex gap-4 pt-4">
                  <button 
                    onClick={() => setStep('generate')}
                    className="flex-1 py-4 rounded-xl border border-gray-200 font-bold hover:bg-gray-50 transition-all flex items-center justify-center gap-2"
                  >
                    <ChevronLeft size={18} />
                    返回
                  </button>
                  <button 
                    onClick={handleGenerateAPlus}
                    className="flex-[2] bg-orange-500 hover:bg-orange-600 text-white font-bold py-4 rounded-xl shadow-lg shadow-orange-500/20 transition-all flex items-center justify-center gap-2"
                  >
                    <Sparkles size={18} />
                    开始生成 A+ 页面
                    <ArrowRight size={18} />
                  </button>
                </div>
              </div>
            </motion.div>
          )}

          {step === 'aplus' && (
            <motion.div 
              key="aplus"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="space-y-8 max-w-7xl mx-auto"
            >
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-3xl font-bold">高级 A+ 页面策划</h2>
                  <p className="text-gray-500">同步生成网页版与手机版预览，符合亚马逊 Premium A+ 规范</p>
                </div>
                <div className="flex gap-3">
                  <button 
                    onClick={() => setStep('generate')}
                    className="px-6 py-3 border border-gray-200 rounded-xl text-sm font-bold hover:bg-gray-50 transition-all"
                  >
                    返回前台图
                  </button>
                  <button 
                    onClick={() => {
                      const zip = new JSZip();
                      const desktopFolder = zip.folder(`${sku || 'Product'}_A+网页版`);
                      const mobileFolder = zip.folder(`${sku || 'Product'}_A+手机版`);
                      
                      aplusModules.forEach((m, i) => {
                        if (m.url) {
                          const base64 = m.url.split(',')[1];
                          desktopFolder?.file(`module_${i+1}_desktop.jpg`, base64, { base64: true });
                          mobileFolder?.file(`module_${i+1}_mobile.jpg`, base64, { base64: true });
                        }
                      });
                      
                      zip.generateAsync({ type: 'blob' }).then(content => {
                        saveAs(content, `${sku || 'Product'}_A+全套资产.zip`);
                      });
                    }}
                    className="px-6 py-3 bg-black text-white rounded-xl text-sm font-bold hover:bg-gray-800 transition-all flex items-center justify-center gap-2 shadow-lg"
                  >
                    <Download size={18} />
                    一键导出全套 A+ 资产
                  </button>
                </div>
              </div>

              <div className="grid grid-cols-1 lg:grid-cols-12 gap-8">
                {/* Desktop View */}
                <div className="lg:col-span-8 space-y-4">
                  <div className="flex items-center gap-2 text-gray-400 font-bold text-xs uppercase tracking-widest">
                    <Monitor size={14} />
                    网页版预览 (1464px)
                  </div>
                  <div className="bg-white shadow-xl rounded-2xl overflow-hidden border border-gray-100">
                    {aplusModules.length === 0 ? (
                      <div className="p-20 text-center space-y-4">
                        <Loader2 className="animate-spin text-orange-500 mx-auto" size={40} />
                        <p className="text-gray-500 font-medium">正在策划 A+ 模块内容...</p>
                      </div>
                    ) : (
                      <div className="space-y-0">
                        {aplusModules.map((module, idx) => (
                          <div key={idx} className="relative group border-b border-gray-50 last:border-0">
                            <div className="relative bg-gray-100 flex items-center justify-center overflow-hidden aspect-[1464/600]">
                              {module.status === 'generating' ? (
                                <div className="flex flex-col items-center gap-2">
                                  <Loader2 className="animate-spin text-orange-500" size={24} />
                                  <span className="text-[10px] text-gray-400">正在生成视觉资产...</span>
                                </div>
                              ) : module.url ? (
                                <img src={module.url} alt={module.title} className="w-full h-full object-cover" />
                              ) : (
                                <ImageIcon className="text-gray-300" size={48} />
                              )}
                              
                              <div className="absolute inset-0 bg-gradient-to-r from-black/70 via-black/20 to-transparent flex items-center px-16">
                                <div className="max-w-lg text-white space-y-3">
                                  <h3 className="text-3xl font-bold leading-tight">{module.title}</h3>
                                  <p className="text-base text-white/90 leading-relaxed font-medium">{module.description}</p>
                                </div>
                              </div>
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                </div>

                {/* Mobile View */}
                <div className="lg:col-span-4 space-y-4">
                  <div className="flex items-center gap-2 text-gray-400 font-bold text-xs uppercase tracking-widest">
                    <Smartphone size={14} />
                    手机版预览 (600px)
                  </div>
                  <div className="bg-white shadow-xl rounded-[2.5rem] overflow-hidden border-[8px] border-gray-900 aspect-[9/19] flex flex-col">
                    <div className="h-6 bg-gray-900 flex items-center justify-center">
                      <div className="w-16 h-1 bg-gray-800 rounded-full" />
                    </div>
                    <div className="flex-1 overflow-y-auto scrollbar-hide">
                      {aplusModules.length === 0 ? (
                        <div className="h-full flex items-center justify-center p-12 text-center">
                          <Loader2 className="animate-spin text-orange-500" size={32} />
                        </div>
                      ) : (
                        <div className="space-y-0">
                          {aplusModules.map((module, idx) => (
                            <div key={idx} className="border-b border-gray-100 last:border-0">
                              <div className="aspect-[4/3] bg-gray-100 relative overflow-hidden">
                                {module.url ? (
                                  <img src={module.url} alt={module.title} className="w-full h-full object-cover" />
                                ) : (
                                  <div className="w-full h-full flex items-center justify-center"><ImageIcon className="text-gray-300" size={32} /></div>
                                )}
                              </div>
                              <div className="p-5 space-y-2">
                                <h3 className="text-lg font-bold text-gray-900">{module.title}</h3>
                                <p className="text-sm text-gray-600 leading-relaxed">{module.description}</p>
                              </div>
                            </div>
                          ))}
                        </div>
                      )}
                    </div>
                    <div className="h-6 bg-gray-900" />
                  </div>
                </div>
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </main>

      {/* Footer */}
      <footer className="py-12 border-t border-gray-200 mt-12 bg-white">
        <div className="max-w-7xl mx-auto px-6 text-center space-y-4">
          <p className="text-sm text-gray-400">© 2026 Amazon Image Pro. 专业级亚马逊视觉解决方案.</p>
          <div className="flex justify-center gap-6 text-xs text-gray-400 uppercase tracking-widest font-bold">
            <a href="#" className="hover:text-orange-500">使用协议</a>
            <a href="#" className="hover:text-orange-500">隐私政策</a>
            <a href="#" className="hover:text-orange-500">联系支持</a>
          </div>
        </div>
      </footer>
      {/* Image Editor Modal */}
      {editingImageIndex !== null && (
        <ImageEditor 
          imageUrl={images[editingImageIndex].url}
          onSave={handleSaveEditedImage}
          onClose={() => setEditingImageIndex(null)}
        />
      )}

      {/* Hover Preview Overlay */}
      <AnimatePresence>
        {hoveredBaseImage && (
          <motion.div
            initial={{ opacity: 0, scale: 0.8 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0.8 }}
            style={{
              position: 'fixed',
              left: mousePos.x + 20,
              top: mousePos.y + 20,
              zIndex: 9999,
              pointerEvents: 'none'
            }}
            className="w-64 h-64 bg-white rounded-2xl shadow-2xl border-4 border-white overflow-hidden"
          >
            <img src={hoveredBaseImage} alt="Preview" className="w-full h-full object-contain" />
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
