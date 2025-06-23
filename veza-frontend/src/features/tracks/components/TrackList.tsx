import { useState } from 'react';
import { Play, Pause, MoreVertical, Heart } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { cn } from '@/shared/utils/helpers';
import type { Track } from '../services/audioService';

interface TrackListProps {
  tracks: Track[];
  onPlay: (track: Track) => void;
  onPause: () => void;
  currentTrackId?: number;
  isPlaying?: boolean;
  className?: string;
}

export const TrackList = ({
  tracks,
  onPlay,
  onPause,
  currentTrackId,
  isPlaying,
  className,
}: TrackListProps) => {
  const [hoveredTrackId, setHoveredTrackId] = useState<number | null>(null);

  const formatDuration = (seconds: number) => {
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = Math.floor(seconds % 60);
    return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`;
  };

  return (
    <div className={cn('space-y-2', className)}>
      {tracks.map((track) => {
        const isCurrentTrack = track.id === currentTrackId;
        const isHovered = track.id === hoveredTrackId;

        return (
          <Card
            key={track.id}
            className={cn(
              'p-4 transition-colors',
              isCurrentTrack && 'bg-gray-50',
              isHovered && 'bg-gray-50'
            )}
            onMouseEnter={() => setHoveredTrackId(track.id)}
            onMouseLeave={() => setHoveredTrackId(null)}
          >
            <div className="flex items-center gap-4">
              <Button
                variant="ghost"
                size="icon"
                onClick={() => {
                  if (isCurrentTrack && isPlaying) {
                    onPause();
                  } else {
                    onPlay(track);
                  }
                }}
                className="w-10 h-10"
              >
                {isCurrentTrack && isPlaying ? (
                  <Pause className="h-5 w-5" />
                ) : (
                  <Play className="h-5 w-5" />
                )}
              </Button>

              <div className="flex-1 min-w-0">
                <h3 className="font-medium truncate">{track.title}</h3>
                <p className="text-sm text-gray-500 truncate">
                  {track.artist}
                  {track.album && ` • ${track.album}`}
                </p>
              </div>

              <div className="flex items-center gap-4">
                <span className="text-sm text-gray-500">
                  {formatDuration(track.duration_seconds)}
                </span>

                <div className="flex items-center gap-2">
                  <Button
                    variant="ghost"
                    size="icon"
                    className="w-8 h-8"
                  >
                    <Heart className="h-4 w-4" />
                  </Button>

                  <Button
                    variant="ghost"
                    size="icon"
                    className="w-8 h-8"
                  >
                    <MoreVertical className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            </div>

            {(track.tags || []).length > 0 && (
              <div className="mt-2 flex flex-wrap gap-2">
                {(track.tags || []).map((tag) => (
                  <span
                    key={tag}
                    className="px-2 py-1 text-xs bg-gray-100 rounded-full"
                  >
                    {tag}
                  </span>
                ))}
              </div>
            )}
          </Card>
        );
      })}
    </div>
  );
}; 