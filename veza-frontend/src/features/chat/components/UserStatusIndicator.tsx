import React from 'react';
import { Circle } from 'lucide-react';

interface UserStatusIndicatorProps {
  isOnline: boolean;
  lastSeen?: string;
  size?: 'sm' | 'md' | 'lg';
}

export function UserStatusIndicator({ isOnline, lastSeen, size = 'md' }: UserStatusIndicatorProps) {
  const sizeClasses = {
    sm: 'w-2 h-2',
    md: 'w-3 h-3',
    lg: 'w-4 h-4'
  };

  return (
    <div className="relative">
      <Circle 
        className={`${sizeClasses[size]} ${
          isOnline 
            ? 'fill-green-500 text-green-500' 
            : 'fill-gray-400 text-gray-400'
        }`} 
      />
    </div>
  );
}

export default UserStatusIndicator; 