import { LucideIcon } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface EmptyStateProps {
  icon?: LucideIcon;
  title: string;
  description?: string;
  action?: {
    label: string;
    onClick: () => void;
  };
}

export const EmptyState = ({
  icon: Icon,
  title,
  description,
  action,
}: EmptyStateProps) => {
  return (
    <div className="flex min-h-[400px] flex-col items-center justify-center p-4 text-center">
      {Icon && (
        <div className="mb-4 rounded-full bg-gray-100 p-3">
          <Icon className="h-6 w-6 text-gray-600" />
        </div>
      )}
      <h3 className="mb-2 text-lg font-medium text-gray-900">{title}</h3>
      {description && (
        <p className="mb-6 max-w-sm text-sm text-gray-500">{description}</p>
      )}
      {action && (
        <Button onClick={action.onClick}>{action.label}</Button>
      )}
    </div>
  );
}; 