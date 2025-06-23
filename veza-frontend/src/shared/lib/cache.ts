interface CacheItem<T> {
  data: T;
  timestamp: number;
  expiresIn: number;
}

class CacheService {
  private cache: Map<string, CacheItem<any>>;
  private readonly DEFAULT_EXPIRY = 5 * 60 * 1000; // 5 minutes

  constructor() {
    this.cache = new Map();
  }

  set<T>(key: string, data: T, expiresIn: number = this.DEFAULT_EXPIRY): void {
    this.cache.set(key, {
      data,
      timestamp: Date.now(),
      expiresIn,
    });
  }

  get<T>(key: string): T | null {
    const item = this.cache.get(key);
    if (!item) return null;

    const isExpired = Date.now() - item.timestamp > item.expiresIn;
    if (isExpired) {
      this.cache.delete(key);
      return null;
    }

    return item.data as T;
  }

  delete(key: string): void {
    this.cache.delete(key);
  }

  clear(): void {
    this.cache.clear();
  }

  // Méthodes utilitaires
  has(key: string): boolean {
    const item = this.cache.get(key);
    if (!item) return false;

    const isExpired = Date.now() - item.timestamp > item.expiresIn;
    if (isExpired) {
      this.cache.delete(key);
      return false;
    }

    return true;
  }

  size(): number {
    return this.cache.size;
  }

  keys(): string[] {
    return Array.from(this.cache.keys());
  }
}

export const cacheService = new CacheService();

// Clés de cache communes
export const CACHE_KEYS = {
  // Products
  PRODUCTS: 'products',
  PRODUCT: (id: number) => `product-${id}`,
  CATEGORIES: 'categories',
  
  // Users
  USER: (id: number) => `user-${id}`,
  USERS: 'users',
  
  // Chat
  ROOMS: 'chat-rooms',
  MESSAGES: (roomId: string) => `messages-${roomId}`,
  DM_MESSAGES: (userId: number) => `dm-messages-${userId}`,
  
  // Tracks
  TRACKS: 'tracks',
  TRACK: (id: number) => `track-${id}`,
  GENRES: 'genres',
} as const; 