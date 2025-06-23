interface ImageOptions {
  width?: number;
  height?: number;
  quality?: number;
  format?: 'webp' | 'jpeg' | 'png';
}

export function getOptimizedImageUrl(
  url: string,
  options: ImageOptions = {}
): string {
  // Si l'URL est déjà une URL d'image optimisée, la retourner telle quelle
  if (url.includes('?width=') || url.includes('?quality=')) {
    return url;
  }

  const params = new URLSearchParams();

  if (options.width) {
    params.append('width', options.width.toString());
  }

  if (options.height) {
    params.append('height', options.height.toString());
  }

  if (options.quality) {
    params.append('quality', options.quality.toString());
  }

  if (options.format) {
    params.append('format', options.format);
  }

  // Ajouter les paramètres d'optimisation à l'URL
  const separator = url.includes('?') ? '&' : '?';
  return `${url}${separator}${params.toString()}`;
}

export function getImagePlaceholder(width: number, height: number): string {
  return `data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='${width}' height='${height}' viewBox='0 0 ${width} ${height}'%3E%3Crect width='${width}' height='${height}' fill='%23f3f4f6'/%3E%3C/svg%3E`;
}

export function preloadImage(url: string): Promise<void> {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.onload = () => resolve();
    img.onerror = reject;
    img.src = url;
  });
}

export function getImageDimensions(url: string): Promise<{ width: number; height: number }> {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.onload = () => {
      resolve({
        width: img.naturalWidth,
        height: img.naturalHeight,
      });
    };
    img.onerror = reject;
    img.src = url;
  });
} 