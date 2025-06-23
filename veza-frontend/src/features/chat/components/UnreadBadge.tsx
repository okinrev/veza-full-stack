import React from 'react';

interface UnreadBadgeProps {
  count: number;
  size?: 'sm' | 'md' | 'lg';
}

export function UnreadBadge({ count, size = 'md' }: UnreadBadgeProps) {
  if (count === 0) return null;

  const sizeClasses = {
    sm: 'text-xs px-1.5 py-0.5 min-w-[16px] h-4',
    md: 'text-sm px-2 py-1 min-w-[20px] h-5',
    lg: 'text-base px-2.5 py-1.5 min-w-[24px] h-6'
  };

  return (
    <div className={`
      bg-red-500 text-white rounded-full flex items-center justify-center font-medium
      ${sizeClasses[size]}
    `}>
      {count > 99 ? '99+' : count}
    </div>
  );
}

export default UnreadBadge; 