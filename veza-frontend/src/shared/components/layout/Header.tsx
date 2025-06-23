import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Bell, Menu, Search, User } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { useAuth } from '@/features/auth/hooks/useAuth';

interface HeaderProps {
  onMenuClick: () => void;
  onToggleSidebar: () => void;
}

export const Header = ({ onMenuClick, onToggleSidebar }: HeaderProps) => {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [searchQuery, setSearchQuery] = useState('');

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (searchQuery.trim()) {
      navigate(`/search?q=${encodeURIComponent(searchQuery)}`);
    }
  };

  return (
    <header className="sticky top-0 z-50 h-16 bg-white border-b border-gray-200">
      <div className="flex h-full items-center justify-between px-4">
        <div className="flex items-center gap-4">
          <Button
            variant="ghost"
            size="icon"
            onClick={onMenuClick}
            className="lg:hidden"
          >
            <Menu className="h-5 w-5" />
          </Button>
          
          <Button
            variant="ghost"
            size="icon"
            onClick={onToggleSidebar}
            className="hidden lg:flex"
          >
            <Menu className="h-5 w-5" />
          </Button>

          <Link to="/" className="flex items-center gap-2">
            <div className="h-8 w-8 rounded-lg bg-gradient-to-r from-blue-600 to-purple-600 flex items-center justify-center">
              <span className="text-white font-bold">T</span>
            </div>
            <span className="text-xl font-bold">Talas</span>
          </Link>
        </div>

        <form onSubmit={handleSearch} className="hidden md:flex flex-1 max-w-md mx-4">
          <Input
            type="search"
            placeholder="Rechercher..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            icon={<Search className="h-4 w-4" />}
            className="w-full"
          />
        </form>

        <div className="flex items-center gap-2">
          <Button variant="ghost" size="icon" className="relative">
            <Bell className="h-5 w-5" />
            <span className="absolute top-1 right-1 h-2 w-2 rounded-full bg-red-500" />
          </Button>

          <div className="relative group">
            <Button variant="ghost" size="icon">
              <User className="h-5 w-5" />
            </Button>
            
            <div className="absolute right-0 mt-2 w-48 bg-white rounded-lg shadow-lg border opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all">
              <div className="p-4 border-b">
                <p className="font-semibold">{user?.username}</p>
                <p className="text-sm text-gray-500">{user?.email}</p>
              </div>
              <div className="p-2">
                <Link
                  to="/profile"
                  className="block px-4 py-2 text-sm hover:bg-gray-100 rounded"
                >
                  Mon profil
                </Link>
                <Link
                  to="/settings"
                  className="block px-4 py-2 text-sm hover:bg-gray-100 rounded"
                >
                  Paramètres
                </Link>
                <button
                  onClick={logout}
                  className="w-full text-left px-4 py-2 text-sm hover:bg-gray-100 rounded text-red-600"
                >
                  Déconnexion
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </header>
  );
}; 