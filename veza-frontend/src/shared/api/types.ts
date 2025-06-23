export interface ApiResponse<T = any> {
  success?: boolean;
  data?: T;
  message?: string;
  error?: string;
  errors?: string[];
  total?: number;
  page?: number;
  limit?: number;
  total_pages?: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
  totalPages: number;
}

// Structure sql.NullString de Go
export interface GoNullString {
  String: string;
  Valid: boolean;
}

// Structure sql.NullTime de Go  
export interface GoNullTime {
  Time: string;
  Valid: boolean;
}

// User compatible avec la structure Go du backend
export interface User {
  id: number;
  username: string;
  email: string;
  password_hash?: string; // Ne sera jamais retourné par l'API
  first_name?: GoNullString | string;
  last_name?: GoNullString | string;
  bio?: GoNullString | string;
  avatar?: GoNullString | string;
  role: string; // 'user', 'admin', 'super_admin'
  is_active: boolean;
  is_verified?: boolean;
  last_login_at?: GoNullTime | string;
  created_at: string;
  updated_at: string;
}

export interface LoginCredentials {
  email: string; // Backend utilise email, pas username
  password: string;
}

export interface RegisterData {
  username: string;
  email: string;
  password: string;
  first_name?: string;
  last_name?: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: User;
}

// Product compatible avec le backend
export interface Product {
  id: number;
  name: string;
  description?: string;
  price: number;
  currency: string; // 'EUR', 'USD', etc.
  category?: string;
  status: string; // 'available', 'sold', etc.
  seller_id: number;
  seller?: User;
  images?: string[];
  created_at: string;
  updated_at: string;
}

// ChatRoom compatible avec le backend
export interface ChatRoom {
  id: number;
  name: string;
  description?: string;
  is_private: boolean;
  creator_id?: number;
  member_count?: number;
  created_at: string;
  updated_at?: string;
}

// ChatMessage compatible avec le backend
export interface ChatMessage {
  id: number;
  content: string;
  sender_id: number;
  receiver_id?: number;
  room_id?: number;
  is_read: boolean;
  created_at: string;
  // Champs additionnels pour l'affichage
  sender?: User;
  sender_username?: string;
  type?: 'text' | 'image' | 'file' | 'system';
}

// Track compatible avec le backend
export interface Track {
  id: number;
  title: string;
  artist: string;
  genre?: string;
  duration?: number; // en secondes
  filename: string;
  file_size: number;
  content_type: string;
  is_public: boolean;
  uploader_id: number;
  uploader?: User;
  download_count: number;
  tags?: string[];
  created_at: string;
  updated_at: string;
}

// File compatible avec le backend
export interface FileUpload {
  id: number;
  filename: string;
  original_name: string;
  file_size: number;
  content_type: string;
  category?: string;
  description?: string;
  url: string;
  uploader_id: number;
  uploader?: User;
  created_at: string;
}

// Listing/Marketplace
export interface Listing {
  id: number;
  title: string;
  description?: string;
  type: 'offer' | 'request';
  category?: string;
  location?: string;
  price?: number;
  currency?: string;
  status: string; // 'active', 'closed', etc.
  user_id: number;
  user?: User;
  images?: string[];
  created_at: string;
  updated_at: string;
}

// Shared Resource
export interface SharedResource {
  id: number;
  name: string;
  description?: string;
  type: string;
  file_url?: string;
  file_size?: number;
  category?: string;
  tags?: string[];
  uploader_id: number;
  uploader?: User;
  download_count: number;
  created_at: string;
  updated_at: string;
}

// Recherche globale
export interface GlobalSearchResult {
  query: string;
  total_results: number;
  search_time: number;
  results: {
    users: User[];
    tracks: Track[];
    products: Product[];
    shared_resources: SharedResource[];
  };
}

// Tags
export interface Tag {
  id: number;
  name: string;
  usage_count: number;
  created_at: string;
}

// Statistiques Admin
export interface AdminDashboard {
  total_users: number;
  active_users: number;
  total_tracks: number;
  public_tracks: number;
  total_listings: number;
  active_listings: number;
  total_messages: number;
  storage_used: string;
  last_updated: string;
}

// Erreurs API
export interface ApiError {
  error: string;
  code?: string;
  details?: any;
  timestamp?: string;
}

// Pagination
export interface PaginationParams {
  page?: number;
  limit?: number;
  search?: string;
}

// Filtres
export interface TrackFilters extends PaginationParams {
  genre?: string;
  uploader?: number;
  tags?: string;
  is_public?: boolean;
}

export interface ProductFilters extends PaginationParams {
  category?: string;
  price_min?: number;
  price_max?: number;
  status?: string;
}

export interface UserFilters extends PaginationParams {
  role?: string;
  is_active?: boolean;
}

// Utilitaires pour extraire les valeurs sql.NullString
export const extractNullString = (value: GoNullString | string | undefined): string => {
  if (!value) return '';
  if (typeof value === 'string') return value;
  if (value && typeof value === 'object' && 'String' in value) {
    return value.Valid ? value.String : '';
  }
  return '';
};

export const extractNullTime = (value: GoNullTime | string | undefined): string | null => {
  if (!value) return null;
  if (typeof value === 'string') return value;
  if (value && typeof value === 'object' && 'Time' in value) {
    return value.Valid ? value.Time : null;
  }
  return null;
}; 