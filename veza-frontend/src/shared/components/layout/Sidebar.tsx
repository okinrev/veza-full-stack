import { NavLink } from 'react-router-dom';
import { cn } from '@/shared/utils/helpers';
import {
  Home,
  MessageSquare,
  Package,
  Music,
  FolderOpen,
  ShoppingBag,
  Users,
  Settings,
  Shield,
  X,
} from 'lucide-react';
import { useAuth } from '@/features/auth/hooks/useAuth';
import { Button } from '@/components/ui/button';

interface SidebarProps {
  open: boolean;
  mobileOpen: boolean;
  onMobileClose: () => void;
}

const navigation = [
  { name: 'Dashboard', href: '/dashboard', icon: Home },
  { name: 'Chat', href: '/chat', icon: MessageSquare },
  { name: 'Produits', href: '/products', icon: Package },
  { name: 'Pistes Audio', href: '/tracks', icon: Music },
  { name: 'Ressources', href: '/resources', icon: FolderOpen },
  { name: 'Marketplace', href: '/marketplace', icon: ShoppingBag },
  { name: 'Utilisateurs', href: '/users', icon: Users },
];

const adminNavigation = [
  { name: 'Administration', href: '/admin', icon: Shield },
];

export const Sidebar = ({ open, mobileOpen, onMobileClose }: SidebarProps) => {
  const { user } = useAuth();
  const isAdmin = user?.role === 'admin';

  return (
    <>
      {/* Mobile backdrop */}
      {mobileOpen && (
        <div
          className="fixed inset-0 z-40 bg-black bg-opacity-50 lg:hidden"
          onClick={onMobileClose}
        />
      )}

      {/* Sidebar */}
      <aside
        className={cn(
          'fixed inset-y-0 left-0 z-40 flex flex-col bg-white border-r border-gray-200 transition-all duration-300',
          open ? 'w-64' : 'w-16',
          mobileOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'
        )}
      >
        <div className="flex h-16 items-center justify-between px-4 border-b lg:hidden">
          <span className="text-xl font-bold">Menu</span>
          <Button variant="ghost" size="icon" onClick={onMobileClose}>
            <X className="h-5 w-5" />
          </Button>
        </div>

        <nav className="flex-1 space-y-1 p-2 mt-5">
          {navigation.map((item) => (
            <NavLink
              key={item.name}
              to={item.href}
              className={({ isActive }) =>
                cn(
                  'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
                  isActive
                    ? 'bg-gray-100 text-gray-900'
                    : 'text-gray-700 hover:bg-gray-50 hover:text-gray-900',
                  !open && 'justify-center'
                )
              }
            >
              <item.icon className="h-5 w-5 flex-shrink-0" />
              {open && <span>{item.name}</span>}
            </NavLink>
          ))}

          {isAdmin && (
            <>
              <div className="my-4 border-t" />
              {adminNavigation.map((item) => (
                <NavLink
                  key={item.name}
                  to={item.href}
                  className={({ isActive }) =>
                    cn(
                      'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
                      isActive
                        ? 'bg-gray-100 text-gray-900'
                        : 'text-gray-700 hover:bg-gray-50 hover:text-gray-900',
                      !open && 'justify-center'
                    )
                  }
                >
                  <item.icon className="h-5 w-5 flex-shrink-0" />
                  {open && <span>{item.name}</span>}
                </NavLink>
              ))}
            </>
          )}
        </nav>

        <div className="border-t p-2">
          <NavLink
            to="/settings"
            className={({ isActive }) =>
              cn(
                'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
                isActive
                  ? 'bg-gray-100 text-gray-900'
                  : 'text-gray-700 hover:bg-gray-50 hover:text-gray-900',
                !open && 'justify-center'
              )
            }
          >
            <Settings className="h-5 w-5 flex-shrink-0" />
            {open && <span>Paramètres</span>}
          </NavLink>
        </div>
      </aside>
    </>
  );
}; 