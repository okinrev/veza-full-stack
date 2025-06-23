import { useState } from 'react';
import { Plus, Trash2, Music, List, Pause, Play, X } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card } from '@/components/ui/card';
import { cn } from '@/shared/utils/helpers';
import type { Track } from '../services/audioService';

interface Playlist {
  id: number;
  name: string;
  tracks: Track[];
  created_at: string;
  updated_at: string;
}

interface PlaylistManagerProps {
  playlists: Playlist[];
  tracks: Track[];
  onPlay: (track: Track) => void;
  onPause: () => void;
  currentTrackId?: number;
  isPlaying: boolean;
  className?: string;
}

export const PlaylistManager = ({
  playlists,
  tracks,
  onPlay,
  onPause,
  currentTrackId,
  isPlaying,
  className
}: PlaylistManagerProps) => {
  const [newPlaylistName, setNewPlaylistName] = useState('');
  const [selectedPlaylistId, setSelectedPlaylistId] = useState<number | null>(null);
  const [isCreatingPlaylist, setIsCreatingPlaylist] = useState(false);

  const handleCreatePlaylist = () => {
    if (!newPlaylistName.trim()) return;
    // TODO: Implémenter la création de playlist
    setIsCreatingPlaylist(false);
    setNewPlaylistName('');
  };

  const handleDeletePlaylist = (playlistId: number) => {
    // TODO: Implémenter la suppression de playlist
  };

  const handleAddTrackToPlaylist = (playlistId: number, trackId: number) => {
    // TODO: Implémenter l'ajout de piste à la playlist
  };

  const handleRemoveTrackFromPlaylist = (playlistId: number, trackId: number) => {
    // TODO: Implémenter la suppression de piste de la playlist
  };

  const selectedPlaylist = playlists.find(p => p.id === selectedPlaylistId);

  return (
    <div className={cn('space-y-4', className)}>
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold">Playlists</h2>
        <Button
          variant="outline"
          size="sm"
          onClick={() => setIsCreatingPlaylist(true)}
        >
          <Plus className="h-4 w-4 mr-2" />
          Nouvelle playlist
        </Button>
      </div>

      {isCreatingPlaylist && (
        <div className="space-y-2">
          <Input
            type="text"
            placeholder="Nom de la playlist"
            value={newPlaylistName}
            onChange={(e) => setNewPlaylistName(e.target.value)}
          />
          <div className="flex gap-2">
            <Button onClick={handleCreatePlaylist}>Créer</Button>
            <Button
              variant="outline"
              onClick={() => setIsCreatingPlaylist(false)}
            >
              Annuler
            </Button>
          </div>
        </div>
      )}

      <div className="space-y-2">
        {playlists.map(playlist => (
          <div
            key={playlist.id}
            className={cn(
              'p-3 rounded-lg border cursor-pointer',
              selectedPlaylistId === playlist.id && 'border-primary'
            )}
            onClick={() => setSelectedPlaylistId(playlist.id)}
          >
            <div className="flex items-center justify-between">
              <span className="font-medium">{playlist.name}</span>
              <Button
                variant="ghost"
                size="icon"
                onClick={(e) => {
                  e.stopPropagation();
                  handleDeletePlaylist(playlist.id);
                }}
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
            <p className="text-sm text-gray-500">
              {playlist.tracks.length} pistes
            </p>
          </div>
        ))}
      </div>

      {selectedPlaylist && (
        <div className="space-y-4">
          <h3 className="font-medium">{selectedPlaylist.name}</h3>
          <div className="space-y-2">
            {selectedPlaylist.tracks.map(track => (
              <div
                key={track.id}
                className="flex items-center justify-between p-2 rounded hover:bg-gray-50"
              >
                <div className="flex-1">
                  <p className="font-medium">{track.title}</p>
                  <p className="text-sm text-gray-500">{track.artist}</p>
                </div>
                <div className="flex items-center gap-2">
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => onPlay(track)}
                  >
                    {currentTrackId === track.id && isPlaying ? (
                      <Pause className="h-4 w-4" />
                    ) : (
                      <Play className="h-4 w-4" />
                    )}
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => handleRemoveTrackFromPlaylist(selectedPlaylist.id, track.id)}
                  >
                    <X className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {selectedPlaylist && (
        <div className="space-y-4">
          <h3 className="font-medium">Ajouter des pistes</h3>
          <div className="space-y-2">
            {tracks
              .filter(track => !selectedPlaylist.tracks.some(t => t.id === track.id))
              .map(track => (
                <div
                  key={track.id}
                  className="flex items-center justify-between p-2 rounded hover:bg-gray-50"
                >
                  <div className="flex-1">
                    <p className="font-medium">{track.title}</p>
                    <p className="text-sm text-gray-500">{track.artist}</p>
                  </div>
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => handleAddTrackToPlaylist(selectedPlaylist.id, track.id)}
                  >
                    <Plus className="h-4 w-4" />
                  </Button>
                </div>
              ))}
          </div>
        </div>
      )}
    </div>
  );
}; 