import express, { Request, Response, NextFunction } from 'express';
import { GoogleGenAI, Type } from '@google/genai';
import dotenv from 'dotenv';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// 优先加载本目录的 .env，回退到 ../web/.env（monorepo 开发模式兼容）
dotenv.config({ path: path.resolve(__dirname, '.env') });
dotenv.config({ path: path.resolve(__dirname, '../web/.env') });

const app = express();
const PORT = process.env.PORT || 3001;

// 生产模式：托管前端静态文件
const distPath = path.resolve(__dirname, '../web/dist');
const isProduction = fs.existsSync(distPath);

app.use(express.json({ limit: '50mb' }));
app.use(express.urlencoded({ extended: true, limit: '50mb' }));

app.use((req: Request, res: Response, next: NextFunction) => {
  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Access-Control-Allow-Methods', 'GET,POST,OPTIONS');
  res.setHeader('Access-Control-Allow-Headers', 'Content-Type');
  if (req.method === 'OPTIONS') {
    res.sendStatus(200);
    return;
  }
  next();
});

if (isProduction) {
  app.use(express.static(distPath));
}

const getAI = () => new GoogleGenAI({ apiKey: process.env.GEMINI_API_KEY || '' });

function getMimeType(dataUrl: string): string {
  const match = dataUrl.match(/^data:([^;]+);/);
  return match ? match[1] : 'image/png';
}

function makeImagePart(dataUrl: string) {
  return {
    inlineData: {
      data: dataUrl.split(',')[1],
      mimeType: getMimeType(dataUrl),
    },
  };
}

// POST /api/analyze — 分析产品卖点
app.post('/api/analyze', async (req: Request, res: Response) => {
  try {
    const { keywords, sellingPoints, competitorLink, sku } = req.body;
    const ai = getAI();

    const prompt = `
    你是一个资深的亚马逊运营专家。请根据以下信息，提炼出9个最具吸引力的产品卖点（Selling Points）。
    SKU: ${sku || '未提供'}
    核心关键词: ${keywords}
    用户提供的卖点: ${sellingPoints}
    ${competitorLink ? `竞品参考: ${competitorLink}` : ''}

    请为每个卖点提供：
    1. 英文标题 (title) 和 英文描述 (description) - 用于生成图片。描述中必须包含指令，要求生成模型"严格保留原产品的纹理、材质和细节特征"，确保产品看起来真实且与原图一致。
    2. 中文标题 (title_cn) 和 中文描述 (description_cn) - 用于用户在网页上快速浏览。

    请以JSON格式返回，包含一个数组，每个元素包含上述四个字段。
  `;

    const response = await ai.models.generateContent({
      model: 'gemini-3.1-pro-preview',
      contents: prompt,
      config: {
        responseMimeType: 'application/json',
        responseSchema: {
          type: Type.OBJECT,
          properties: {
            sellingPoints: {
              type: Type.ARRAY,
              items: {
                type: Type.OBJECT,
                properties: {
                  title: { type: Type.STRING },
                  description: { type: Type.STRING },
                  title_cn: { type: Type.STRING },
                  description_cn: { type: Type.STRING },
                },
                required: ['title', 'description', 'title_cn', 'description_cn'],
              },
            },
          },
        },
      },
    });

    const data = JSON.parse(response.text || '{}');
    res.json({ data: data.sellingPoints });
  } catch (error) {
    console.error('[/api/analyze]', error);
    res.status(500).json({ error: String(error) });
  }
});

// POST /api/generate-image — 生成亚马逊产品图
app.post('/api/generate-image', async (req: Request, res: Response) => {
  try {
    const { prompt, aspectRatio, productImages, styleRefImage } = req.body as {
      prompt: string;
      aspectRatio: '1:1' | '4:5';
      productImages: string[];
      styleRefImage?: string;
    };
    const ai = getAI();

    const aspectHint = aspectRatio === '1:1'
      ? 'square format (1:1)'
      : 'portrait format (4:5)';
    const stylePrompt = styleRefImage
      ? ' Follow the style, composition, and lighting of the provided style reference image.'
      : '';
    const enhancedPrompt = `${prompt}.${stylePrompt} Generate in ${aspectHint}. Photorealistic, high quality, maintain the original product's texture, material, and fine details exactly as in the product images. Ensure the product appears consistent with all provided reference images. No distortion of product features.`;

    const productParts = productImages.map(makeImagePart);
    const styleParts = styleRefImage ? [makeImagePart(styleRefImage)] : [];

    const response = await ai.models.generateContent({
      model: 'gemini-3.1-flash-image-preview',
      contents: {
        parts: [...productParts, ...styleParts, { text: enhancedPrompt }],
      },
      config: {
        responseModalities: ['IMAGE', 'TEXT'],
      },
    });

    for (const part of response.candidates?.[0]?.content?.parts || []) {
      if (part.inlineData) {
        const mime = part.inlineData.mimeType || 'image/png';
        res.json({ data: `data:${mime};base64,${part.inlineData.data}` });
        return;
      }
    }
    res.json({ data: null });
  } catch (error) {
    console.error('[/api/generate-image]', error);
    res.status(500).json({ error: String(error) });
  }
});

// POST /api/edit-image — 编辑图片（去背景/扩图等）
app.post('/api/edit-image', async (req: Request, res: Response) => {
  try {
    const { baseImage, instruction, aspectRatio = '1:1' } = req.body as {
      baseImage: string;
      instruction: string;
      aspectRatio?: '1:1' | '4:5';
    };
    const ai = getAI();

    const aspectHint = aspectRatio === '1:1' ? 'square format (1:1)' : 'portrait format (4:5)';
    const fullInstruction = `${instruction} Output in ${aspectHint}. Maintain photorealistic quality and preserve all product details.`;

    const response = await ai.models.generateContent({
      model: 'gemini-3.1-flash-image-preview',
      contents: {
        parts: [
          makeImagePart(baseImage),
          { text: fullInstruction },
        ],
      },
      config: {
        responseModalities: ['IMAGE', 'TEXT'],
      },
    });

    for (const part of response.candidates?.[0]?.content?.parts || []) {
      if (part.inlineData) {
        const mime = part.inlineData.mimeType || 'image/png';
        res.json({ data: `data:${mime};base64,${part.inlineData.data}` });
        return;
      }
    }
    res.json({ data: null });
  } catch (error) {
    console.error('[/api/edit-image]', error);
    res.status(500).json({ error: String(error) });
  }
});

// POST /api/aplus-content — 生成 A+ 页面方案
app.post('/api/aplus-content', async (req: Request, res: Response) => {
  try {
    const { keywords, sellingPoints, sku, template = 'standard', refImages } = req.body as {
      keywords: string;
      sellingPoints: string[];
      sku?: string;
      template?: string;
      refImages?: string[];
    };
    const ai = getAI();

    const templateDescriptions: Record<string, string> = {
      standard: '标准5模块：包含品牌故事、核心功能、场景化展示、细节展示、对比表。',
      visual: '视觉导向4模块：包含超大首图、三列功能展示、大图场景、细节放大。',
      technical: '技术详尽6模块：包含顶部横幅、爆炸图展示、材质细节、使用指南、安全说明、对比表。',
      minimalist: '极简3模块：包含干净的顶部图、场景网格、核心参数。',
    };

    const imageParts = refImages?.map(makeImagePart) || [];

    const prompt = `
    你是一个亚马逊高级A+页面设计专家。请根据以下产品卖点和选定的模板，策划一套符合亚马逊"高级A+ (Premium A+)"要求的页面方案。
    SKU: ${sku || '未提供'}
    关键词: ${keywords}
    核心卖点: ${sellingPoints.join(', ')}
    选定模板: ${template} (${templateDescriptions[template] || templateDescriptions['standard']})

    ${imageParts.length > 0 ? '参考图片已提供，请在策划模块和生成图片提示词时，参考这些竞品或参考图的风格、排版和视觉逻辑。' : ''}

    要求：
    1. 展示逻辑清晰：根据模板要求，从品牌心智到核心功能，再到场景体验和细节参数。
    2. 文字简洁有力：包含必要的关键词、卖点和属性词，符合亚马逊合规要求。
    3. 视觉引导：为每个模块提供高质量的图片生成指令。指令中必须包含"严格保留原产品的纹理、材质和细节特征"。

    请以JSON格式返回，包含：
    - modules: 数组，每个模块包含 type, title, description, imagePrompt (用于生成该模块图片的提示词)。
  `;

    const response = await ai.models.generateContent({
      model: 'gemini-3.1-pro-preview',
      contents: {
        parts: [...imageParts, { text: prompt }],
      },
      config: {
        responseMimeType: 'application/json',
        responseSchema: {
          type: Type.OBJECT,
          properties: {
            modules: {
              type: Type.ARRAY,
              items: {
                type: Type.OBJECT,
                properties: {
                  type: { type: Type.STRING },
                  title: { type: Type.STRING },
                  description: { type: Type.STRING },
                  imagePrompt: { type: Type.STRING },
                },
                required: ['type', 'title', 'description', 'imagePrompt'],
              },
            },
          },
        },
      },
    });

    const data = JSON.parse(response.text || '{}');
    res.json({ data: data.modules });
  } catch (error) {
    console.error('[/api/aplus-content]', error);
    res.status(500).json({ error: String(error) });
  }
});

// 健康检查
app.get('/api/health', (_req: Request, res: Response) => {
  res.json({ status: 'ok', port: PORT });
});

// SPA 回退：所有非 /api 请求返回 index.html（生产模式）
if (isProduction) {
  app.get('*', (_req: Request, res: Response) => {
    res.sendFile(path.join(distPath, 'index.html'));
  });
}

app.listen(PORT, () => {
  const mode = isProduction ? '生产' : '开发';
  console.log(`✅ 后端服务已启动（${mode}模式），监听端口: ${PORT}`);
});
