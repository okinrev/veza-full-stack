import { apiClient } from '@/shared/api/client';
import type { Track, TrackFilters, PaginatedResponse } from '@/shared/api/types';

export interface TracksResponse {
  tracks: Track[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export interface UploadTrackData {
  title: string;
  artist: string;
  genre?: string;
  is_public: boolean;
  tags?: string[];
}

export class AudioService {
  // Récupérer les pistes avec filtres
  async getTracks(filters?: TrackFilters): Promise<TracksResponse> {
    const params = new URLSearchParams();
    
    if (filters?.page) params.append('page', filters.page.toString());
    if (filters?.limit) params.append('limit', filters.limit.toString());
    if (filters?.search) params.append('search', filters.search);
    if (filters?.genre) params.append('genre', filters.genre);
    if (filters?.uploader) params.append('uploader', filters.uploader.toString());
    if (filters?.tags) params.append('tags', filters.tags);
    if (filters?.is_public !== undefined) params.append('is_public', filters.is_public.toString());

    const queryString = params.toString();
    const url = `/api/v1/tracks${queryString ? `?${queryString}` : ''}`;
    
    return apiClient.get<TracksResponse>(url);
  }

  // Récupérer une piste spécifique
  async getTrack(id: number): Promise<Track> {
    return apiClient.get<Track>(`/api/v1/tracks/${id}`);
  }

  // Upload d'une nouvelle piste
  async uploadTrack(
    file: File, 
    metadata: UploadTrackData, 
    onProgress?: (progress: number) => void
  ): Promise<Track> {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('title', metadata.title);
    formData.append('artist', metadata.artist);
    if (metadata.genre) formData.append('genre', metadata.genre);
    formData.append('is_public', metadata.is_public.toString());
    if (metadata.tags && metadata.tags.length > 0) {
      formData.append('tags', metadata.tags.join(','));
    }

    return apiClient.uploadFile<Track>('/api/v1/tracks', file, {
      title: metadata.title,
      artist: metadata.artist,
      genre: metadata.genre,
      is_public: metadata.is_public,
      tags: metadata.tags?.join(',')
    }, onProgress);
  }

  // Mettre à jour les métadonnées d'une piste
  async updateTrack(id: number, metadata: Partial<UploadTrackData>): Promise<Track> {
    return apiClient.put<Track>(`/api/v1/tracks/${id}`, metadata);
  }

  // Supprimer une piste
  async deleteTrack(id: number): Promise<void> {
    return apiClient.delete(`/api/v1/tracks/${id}`);
  }

  // Obtenir l'URL de streaming pour une piste
  async getStreamUrl(id: number): Promise<{ url: string }> {
    return apiClient.get<{ url: string }>(`/api/v1/tracks/${id}/stream`);
  }

  // Récupérer les pistes de l'utilisateur connecté
  async getMyTracks(filters?: Omit<TrackFilters, 'uploader'>): Promise<TracksResponse> {
    const params = new URLSearchParams();
    
    if (filters?.page) params.append('page', filters.page.toString());
    if (filters?.limit) params.append('limit', filters.limit.toString());
    if (filters?.search) params.append('search', filters.search);
    if (filters?.genre) params.append('genre', filters.genre);
    if (filters?.tags) params.append('tags', filters.tags);

    const queryString = params.toString();
    const url = `/api/v1/tracks/my${queryString ? `?${queryString}` : ''}`;
    
    return apiClient.get<TracksResponse>(url);
  }

  // Incrémenter le nombre de lectures
  async incrementPlayCount(id: number): Promise<void> {
    return apiClient.post(`/api/v1/tracks/${id}/play`);
  }

  // Recherche de pistes
  async searchTracks(query: string, filters?: Omit<TrackFilters, 'search'>): Promise<TracksResponse> {
    const params = new URLSearchParams();
    params.append('q', query);
    
    if (filters?.page) params.append('page', filters.page.toString());
    if (filters?.limit) params.append('limit', filters.limit.toString());
    if (filters?.genre) params.append('genre', filters.genre);
    if (filters?.uploader) params.append('uploader', filters.uploader.toString());
    if (filters?.tags) params.append('tags', filters.tags);
    if (filters?.is_public !== undefined) params.append('is_public', filters.is_public.toString());

    return apiClient.get<TracksResponse>(`/api/v1/search/tracks?${params.toString()}`);
  }

  // Télécharger une piste
  async downloadTrack(id: number): Promise<Blob> {
    const response = await fetch(`/api/v1/tracks/${id}/download`, {
      headers: {
        'Authorization': `Bearer ${JSON.parse(localStorage.getItem('auth-store') || '{}').state?.accessToken}`
      }
    });
    
    if (!response.ok) {
      throw new Error('Erreur lors du téléchargement');
    }
    
    return response.blob();
  }

  // Obtenir les genres disponibles
  async getGenres(): Promise<string[]> {
    return apiClient.get<string[]>('/api/v1/tracks/genres');
  }

  // Obtenir les tags populaires
  async getPopularTags(limit = 20): Promise<Array<{name: string, count: number}>> {
    return apiClient.get(`/api/v1/tracks/tags/popular?limit=${limit}`);
  }
}

// Instance singleton
export const audioService = new AudioService();
export default audioService; 