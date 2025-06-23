export interface Track {
  id: number;
  title: string;
  filename: string;
  artist: string;
  duration_seconds: number;
  tags: string[];
  is_public: boolean;
  uploader_id: number;
  created_at: string;
  album?: string;
  uploader_username?: string;
  play_count?: number;
  waveform_url?: string;
  file_url?: string;
}

export interface ApiResponse<T> {
  success: boolean;
  data: T;
  message: string;
}

export interface TracksResponse {
  tracks: Track[];
  total: number;
  page: number;
  limit: number;
} 