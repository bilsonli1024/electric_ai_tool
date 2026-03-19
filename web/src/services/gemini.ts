const API_BASE = import.meta.env.VITE_API_BASE_URL || '';

async function post<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const err = await res.text();
    throw new Error(`后端请求失败 [${res.status}]: ${err}`);
  }
  const json = await res.json();
  if (json.error) throw new Error(json.error);
  return json.data as T;
}

export const analyzeSellingPoints = (
  keywords: string,
  sellingPoints: string,
  competitorLink?: string,
  sku?: string,
) =>
  post<{ title: string; description: string; title_cn: string; description_cn: string }[]>(
    '/api/analyze',
    { keywords, sellingPoints, competitorLink, sku },
  );

export const generateAmazonImage = (
  prompt: string,
  aspectRatio: '1:1' | '4:5',
  productImages: string[],
  styleRefImage?: string,
) =>
  post<string | null>('/api/generate-image', {
    prompt,
    aspectRatio,
    productImages,
    styleRefImage,
  });

export const editImage = (
  baseImage: string,
  instruction: string,
  aspectRatio: '1:1' | '4:5' = '1:1',
) => post<string | null>('/api/edit-image', { baseImage, instruction, aspectRatio });

export const generateAPlusContent = (
  keywords: string,
  sellingPoints: string[],
  sku?: string,
  template: string = 'standard',
  refImages?: string[],
) =>
  post<{ type: string; title: string; description: string; imagePrompt: string }[]>(
    '/api/aplus-content',
    { keywords, sellingPoints, sku, template, refImages },
  );
