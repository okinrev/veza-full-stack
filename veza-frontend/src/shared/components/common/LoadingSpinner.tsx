import { cn } from '@/shared/utils/helpers';

interface LoadingSpinnerProps {
  size?: 'sm' | 'md' | 'lg';
  fullScreen?: boolean;
  className?: string;
}

export const LoadingSpinner = ({
  size = 'md',
  fullScreen = false,
  className,
}: LoadingSpinnerProps) => {
  const sizeClasses = {
    sm: 'h-4 w-4',
    md: 'h-8 w-8',
    lg: 'h-12 w-12',
  };

  const spinner = (
    <div
      className={cn(
        'animate-spin rounded-full border-4 border-gray-200 border-t-primary',
        sizeClasses[size],
        className
      )}
    />
  );

  if (fullScreen) {
    return (
      <div className="fixed inset-0 z-50 flex items-center justify-center bg-white bg-opacity-80">
        {spinner}
      </div>
    );
  }

  return spinner;
}; 