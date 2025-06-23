import { useState, useEffect, useCallback } from 'react';
import { useNotifications } from '@/shared/hooks/useNotifications';
import { audioService } from '../services/audioService';
import { Track } from '../types';
import { TrackList } from '../components/TrackList';
import { TrackUpload } from '../components/TrackUpload';
import { AudioPlayer } from '../components/AudioPlayer';
import { WaveformDisplay } from '../components/WaveformDisplay';
import { PlaylistManager } from '../components/PlaylistManager';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Search, Upload, Filter, Music, Users, Tag } from 'lucide-react';
import { cn } from '@/shared/utils/helpers';

export const TracksPage = () => {
  const [tracks, setTracks] = useState<Track[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [currentTrack, setCurrentTrack] = useState<Track | null>(null);
  const [isPlaying, setIsPlaying] = useState(false);
  const [showUpload, setShowUpload] = useState(false);
  const [currentTime, setCurrentTime] = useState(0);
  const [filters, setFilters] = useState({
    genre: '',
    artist: '',
    tags: [] as string[]
  });
  const [stats, setStats] = useState({
    totalTracks: 0,
    totalPlays: 0,
    totalDuration: 0
  });
  const [totalTracks, setTotalTracks] = useState(0);
  const { showNotification } = useNotifications();

  const loadTracks = useCallback(async () => {
    try {
      setLoading(true);
      console.log('Loading tracks with filters:', filters);
      const response = await audioService.getTracks({
        search: searchQuery,
        artist: filters.artist,
        tags: filters.tags
      });
      console.log('Received tracks response:', response);
      
      setTracks(response.tracks);
      setTotalTracks(response.total);
      
      // Calculer les statistiques
      const stats = {
        totalTracks: response.total,
        totalPlays: response.tracks.reduce((sum, track) => sum + (track.play_count || 0), 0),
        totalDuration: response.tracks.reduce((sum, track) => sum + (track.duration_seconds || 0), 0)
      };
      console.log('Calculated stats:', stats);
      setStats(stats);
    } catch (error) {
      console.error('Error loading tracks:', error);
      showNotification(
        error instanceof Error ? error.message : 'Erreur lors du chargement des pistes',
        'error'
      );
    } finally {
      setLoading(false);
    }
  }, [searchQuery, filters, showNotification]);

  useEffect(() => {
    loadTracks();
  }, [loadTracks]);

  const handleSearch = async () => {
    try {
      setLoading(true);
      const response = await audioService.getTracks({ 
        search: searchQuery,
        artist: filters.artist,
        tags: filters.tags
      });
      setTracks(response.tracks);
    } catch (err) {
      showNotification('Erreur lors de la recherche', 'error');
    } finally {
      setLoading(false);
    }
  };

  const handlePlay = async (track: Track) => {
    try {
      if (currentTrack?.id === track.id) {
        setIsPlaying(!isPlaying);
      } else {
        console.log('Getting stream URL for track:', track);
        const { url } = await audioService.getStreamUrl(track.id);
        console.log('Using file URL:', url);
        
        if (!url) {
          throw new Error('URL du fichier non disponible');
        }

        // Vérifier si le fichier existe
        try {
          console.log('Checking if file exists:', url);
          const response = await fetch(url, { 
            method: 'HEAD',
            headers: {
              'Accept': 'audio/mpeg,audio/*;q=0.9,*/*;q=0.8',
              'Authorization': `Bearer ${localStorage.getItem('access_token')}`
            }
          });
          console.log('File check response:', response.status, response.statusText);
          
          if (!response.ok) {
            throw new Error(`Fichier non trouvé (${response.status}): ${response.statusText}`);
          }

          // Vérifier le type de contenu
          const contentType = response.headers.get('content-type');
          console.log('Content-Type:', contentType);
          if (!contentType?.includes('audio/')) {
            throw new Error(`Type de fichier invalide: ${contentType}`);
          }
        } catch (err) {
          console.error('Error checking file:', err);
          throw new Error(`Impossible d'accéder au fichier audio: ${err instanceof Error ? err.message : 'Erreur inconnue'}`);
        }

        setCurrentTrack({ ...track, file_url: url });
        setIsPlaying(true);
        setCurrentTime(0);
        await audioService.incrementPlayCount(track.id);
      }
    } catch (err) {
      console.error('Error playing track:', err);
      showNotification(
        err instanceof Error ? err.message : 'Erreur lors de la lecture de la piste',
        'error'
      );
    }
  };

  const handlePause = () => {
    setIsPlaying(false);
  };

  const handleTimeUpdate = (time: number) => {
    setCurrentTime(time);
  };

  const handleUploadComplete = () => {
    setShowUpload(false);
    loadTracks();
    showNotification('Piste téléchargée avec succès', 'success');
  };

  const formatDuration = (seconds: number) => {
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = Math.floor(seconds % 60);
    return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`;
  };

  if (loading) {
    return <div>Chargement...</div>;
  }

  if (error) {
    return <div className="text-red-500">{error}</div>;
  }

  return (
    <div className="container mx-auto p-4">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold">Pistes Audio</h1>
          <div className="text-sm text-gray-500 mt-1">
            {totalTracks} pistes • {stats.totalPlays} écoutes • {formatDuration(stats.totalDuration)}
          </div>
        </div>
        <Button onClick={() => setShowUpload(true)}>
          <Upload className="w-4 h-4 mr-2" />
          Télécharger
        </Button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className="md:col-span-2">
          <Input
            type="text"
            placeholder="Rechercher des pistes..."
            value={searchQuery}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSearchQuery(e.target.value)}
            className="w-full"
          />
        </div>
        <div>
          <Input
            type="text"
            placeholder="Artiste..."
            value={filters.artist}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setFilters({ ...filters, artist: e.target.value })}
          />
        </div>
        <div>
          <Button onClick={handleSearch} className="w-full">
            <Search className="w-4 h-4 mr-2" />
            Rechercher
          </Button>
        </div>
      </div>

      {currentTrack && currentTrack.file_url && (
        <div className="mb-6">
          <AudioPlayer
            audioUrl={currentTrack.file_url}
            title={currentTrack.title}
            artist={currentTrack.artist}
            isPlaying={isPlaying}
            onPlay={() => setIsPlaying(true)}
            onPause={() => setIsPlaying(false)}
            onTimeUpdate={handleTimeUpdate}
          />
          <WaveformDisplay
            audioUrl={currentTrack.file_url}
            currentTime={currentTime}
            duration={currentTrack.duration_seconds}
            onSeek={(time) => setCurrentTime(time)}
          />
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="md:col-span-2">
          <TrackList
            tracks={tracks}
            onPlay={handlePlay}
            onPause={handlePause}
            currentTrackId={currentTrack?.id}
            isPlaying={isPlaying}
          />
        </div>
        <div>
          <PlaylistManager
            playlists={[]}
            tracks={tracks}
            onPlay={handlePlay}
            onPause={handlePause}
            currentTrackId={currentTrack?.id}
            isPlaying={isPlaying}
          />
        </div>
      </div>

      {showUpload && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
          <div className="bg-white p-6 rounded-lg w-full max-w-2xl">
            <TrackUpload onUploadComplete={handleUploadComplete} />
            <Button
              variant="outline"
              className="mt-4"
              onClick={() => setShowUpload(false)}
            >
              Annuler
            </Button>
          </div>
        </div>
      )}
    </div>
  );
};

export default TracksPage; 