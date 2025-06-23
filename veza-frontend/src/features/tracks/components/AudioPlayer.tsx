import { useEffect, useRef, useState } from 'react';
import { Play, Pause, Volume2, VolumeX, SkipBack, SkipForward } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { cn } from '@/shared/utils/helpers';

interface AudioPlayerProps {
  audioUrl?: string;
  title: string;
  artist: string;
  isPlaying: boolean;
  onPlay: () => void;
  onPause: () => void;
  onTimeUpdate: (time: number) => void;
  className?: string;
}

export const AudioPlayer = ({
  audioUrl,
  title,
  artist,
  isPlaying,
  onPlay,
  onPause,
  onTimeUpdate,
  className
}: AudioPlayerProps) => {
  const audioRef = useRef<HTMLAudioElement>(null);

  useEffect(() => {
    if (audioRef.current) {
      if (isPlaying) {
        audioRef.current.play();
      } else {
        audioRef.current.pause();
      }
    }
  }, [isPlaying]);

  useEffect(() => {
    const audio = audioRef.current;
    if (!audio) return;

    const handleTimeUpdate = () => {
      onTimeUpdate(audio.currentTime);
    };

    audio.addEventListener('timeupdate', handleTimeUpdate);
    return () => {
      audio.removeEventListener('timeupdate', handleTimeUpdate);
    };
  }, [onTimeUpdate]);

  return (
    <div className={cn('flex items-center gap-4 p-4 bg-gray-50 rounded-lg', className)}>
      <audio ref={audioRef} src={audioUrl} />
      <div className="flex-1">
        <h3 className="font-medium">{title}</h3>
        <p className="text-sm text-gray-500">{artist}</p>
      </div>
      <Button
        variant="ghost"
        size="icon"
        onClick={isPlaying ? onPause : onPlay}
      >
        {isPlaying ? <Pause className="h-6 w-6" /> : <Play className="h-6 w-6" />}
      </Button>
    </div>
  );
}; 