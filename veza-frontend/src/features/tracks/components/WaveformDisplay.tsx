import { useEffect, useRef, useState } from 'react';
import { cn } from '@/shared/utils/helpers';

interface WaveformDisplayProps {
  audioUrl: string;
  currentTime?: number;
  duration?: number;
  onSeek?: (time: number) => void;
  className?: string;
}

export const WaveformDisplay = ({
  audioUrl,
  currentTime = 0,
  duration = 0,
  onSeek,
  className,
}: WaveformDisplayProps) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [waveformData, setWaveformData] = useState<number[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadWaveform = async () => {
      try {
        setIsLoading(true);
        setError(null);

        // Créer un contexte audio
        const audioContext = new (window.AudioContext || (window as any).webkitAudioContext)();
        const response = await fetch(audioUrl);
        const arrayBuffer = await response.arrayBuffer();
        const audioBuffer = await audioContext.decodeAudioData(arrayBuffer);

        // Extraire les données de la forme d'onde
        const channelData = audioBuffer.getChannelData(0);
        const samples = 100; // Nombre de points à afficher
        const blockSize = Math.floor(channelData.length / samples);
        const waveform: number[] = [];

        for (let i = 0; i < samples; i++) {
          const start = blockSize * i;
          let sum = 0;
          for (let j = 0; j < blockSize; j++) {
            sum += Math.abs(channelData[start + j]);
          }
          waveform.push(sum / blockSize);
        }

        // Normaliser les données
        const max = Math.max(...waveform);
        const normalizedWaveform = waveform.map(value => value / max);
        setWaveformData(normalizedWaveform);
      } catch (err) {
        setError('Erreur lors du chargement de la forme d\'onde');
        console.error('Waveform loading error:', err);
      } finally {
        setIsLoading(false);
      }
    };

    loadWaveform();
  }, [audioUrl]);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas || waveformData.length === 0) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // Ajuster la taille du canvas
    const dpr = window.devicePixelRatio || 1;
    const rect = canvas.getBoundingClientRect();
    canvas.width = rect.width * dpr;
    canvas.height = rect.height * dpr;
    ctx.scale(dpr, dpr);

    // Effacer le canvas
    ctx.clearRect(0, 0, rect.width, rect.height);

    // Dessiner la forme d'onde
    const barWidth = rect.width / waveformData.length;
    const barGap = 1;
    const maxHeight = rect.height * 0.8;

    ctx.fillStyle = '#e5e7eb'; // Couleur de fond des barres

    waveformData.forEach((value, index) => {
      const x = index * (barWidth + barGap);
      const height = value * maxHeight;
      const y = (rect.height - height) / 2;

      ctx.fillRect(x, y, barWidth, height);
    });

    // Dessiner la progression
    if (duration > 0) {
      const progress = currentTime / duration;
      const progressX = rect.width * progress;

      ctx.fillStyle = '#3b82f6'; // Couleur de la progression
      ctx.fillRect(0, 0, progressX, rect.height);
    }
  }, [waveformData, currentTime, duration]);

  const handleClick = (e: React.MouseEvent<HTMLCanvasElement>) => {
    if (!onSeek || !duration) return;

    const canvas = canvasRef.current;
    if (!canvas) return;

    const rect = canvas.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const progress = x / rect.width;
    const time = progress * duration;

    onSeek(time);
  };

  if (error) {
    return (
      <div className={cn('text-sm text-red-600', className)}>
        {error}
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className={cn('animate-pulse bg-gray-200 rounded', className)} style={{ height: 100 }} />
    );
  }

  return (
    <canvas
      ref={canvasRef}
      onClick={handleClick}
      className={cn('w-full h-24 cursor-pointer', className)}
    />
  );
}; 