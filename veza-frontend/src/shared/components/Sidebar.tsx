import { NavLink } from 'react-router-dom';
import { useAuthStore } from '@/features/auth/store/authStore';
import { cn } from '@/shared/lib/utils';
import {
  Home,
  MessageSquare,
  Music,
  Package,
  Share2,
  Settings,
  User,
} from 'lucide-react';

const navigation = [
  { name: 'Dashboard', href: '/dashboard', icon: Home },
  { name: 'Chat', href: '/chat', icon: MessageSquare },
  { name: 'Pistes', href: '/tracks', icon: Music },
  { name: 'Produits', href: '/products', icon: Package },
  { name: 'Ressources', href: '/resources', icon: Share2 },
  { name: 'Profil', href: '/profile', icon: User },
  { name: 'Paramètres', href: '/settings', icon: Settings },
];

export function Sidebar() {
  const { isAuthenticated } = useAuthStore();

  if (!isAuthenticated) return null;

  return (
    <div className="w-64 border-r bg-background">
      <div className="space-y-4 py-4">
        <div className="px-3 py-2">
          <div className="space-y-1">
            {navigation.map((item) => (
              <NavLink
                key={item.name}
                to={item.href}
                className={({ isActive }) =>
                  cn(
                    'flex items-center rounded-md px-3 py-2 text-sm font-medium',
                    isActive
                      ? 'bg-accent text-accent-foreground'
                      : 'hover:bg-accent hover:text-accent-foreground'
                  )
                }
              >
                <item.icon className="mr-2 h-4 w-4" />
                {item.name}
              </NavLink>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
} 