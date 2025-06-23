import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

interface DebugPanelProps {
  onClearCache?: () => void;
  onMockData?: () => void;
}

interface PerformanceMetrics {
  memory?: number;
  fps: number;
  loadTime: number;
}

export function DebugPanel({ onClearCache, onMockData }: DebugPanelProps) {
  const [isVisible, setIsVisible] = useState(false);
  const [metrics, setMetrics] = useState<PerformanceMetrics>({
    fps: 0,
    loadTime: 0,
  });

  useEffect(() => {
    let frameCount = 0;
    let lastTime = window.performance.now();
    let animationFrameId: number;

    const measureFPS = () => {
      const currentTime = window.performance.now();
      frameCount++;

      if (currentTime - lastTime >= 1000) {
        setMetrics(prev => ({
          ...prev,
          fps: Math.round((frameCount * 1000) / (currentTime - lastTime))
        }));
        frameCount = 0;
        lastTime = currentTime;
      }

      animationFrameId = requestAnimationFrame(measureFPS);
    };

    // Mesurer le temps de chargement
    const loadTime = window.performance.now();
    setMetrics(prev => ({ ...prev, loadTime }));

    // Mesurer la mémoire si disponible (Chrome uniquement)
    if ('memory' in window.performance) {
      const memory = (window.performance as any).memory;
      if (memory && typeof memory.usedJSHeapSize === 'number') {
        setMetrics(prev => ({
          ...prev,
          memory: Math.round(memory.usedJSHeapSize / 1024 / 1024)
        }));
      }
    }

    // Démarrer la mesure FPS
    measureFPS();

    return () => {
      cancelAnimationFrame(animationFrameId);
    };
  }, []);

  if (!isVisible) {
    return (
      <Button
        variant="outline"
        size="sm"
        className="fixed bottom-4 right-4 z-50"
        onClick={() => setIsVisible(true)}
      >
        Debug
      </Button>
    );
  }

  return (
    <Card className="fixed bottom-4 right-4 w-80 z-50">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">Panneau de Débogage</CardTitle>
        <Button
          variant="ghost"
          size="sm"
          onClick={() => setIsVisible(false)}
        >
          ×
        </Button>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          <div className="grid grid-cols-2 gap-2 text-sm">
            <div>FPS:</div>
            <div className="font-mono">{metrics.fps}</div>
            <div>Temps de chargement:</div>
            <div className="font-mono">{metrics.loadTime.toFixed(0)}ms</div>
            {metrics.memory && (
              <>
                <div>Mémoire utilisée:</div>
                <div className="font-mono">{metrics.memory}MB</div>
              </>
            )}
          </div>

          <div className="pt-2 space-y-2">
            <Button
              variant="outline"
              size="sm"
              className="w-full"
              onClick={() => {
                localStorage.clear();
                onClearCache?.();
              }}
            >
              Vider le cache
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="w-full"
              onClick={() => {
                onMockData?.();
              }}
            >
              Charger données de test
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
} 