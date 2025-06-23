import { Link, useNavigate } from "react-router-dom";
import { useAuthStore } from "@/shared/stores/authStore";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Bell, LogOut, Settings, User } from "lucide-react";

export function Navbar() {
  const navigate = useNavigate();
  const { user, logout } = useAuthStore();

  const handleLogout = () => {
    logout();
    navigate("/login");
  };

  return (
    <nav className="border-b">
      <div className="flex h-16 items-center px-4">
        <div className="flex items-center space-x-4">
          <Link to="/" className="text-xl font-bold">
            Talas
          </Link>
          <div className="hidden md:flex space-x-4">
            <Link to="/dashboard" className="text-sm font-medium transition-colors hover:text-primary">
              Tableau de bord
            </Link>
            <Link to="/chat" className="text-sm font-medium transition-colors hover:text-primary">
              Chat
            </Link>
            <Link to="/resources" className="text-sm font-medium transition-colors hover:text-primary">
              Ressources
            </Link>
            <Link to="/tracks" className="text-sm font-medium transition-colors hover:text-primary">
              Pistes
            </Link>
          </div>
        </div>
        <div className="ml-auto flex items-center space-x-4">
          <Button variant="ghost" size="icon" type="button">
            <Bell className="h-5 w-5" />
          </Button>
          {user ? (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" className="relative h-8 w-8 rounded-full" type="button">
                  <Avatar className="h-8 w-8">
                    <AvatarImage src={`https://avatar.vercel.sh/${user.username}`} alt={user.username} />
                    <AvatarFallback>{user.username?.charAt(0).toUpperCase()}</AvatarFallback>
                  </Avatar>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent className="w-56" align="end" forceMount>
                <DropdownMenuLabel className="font-normal">
                  <div className="flex flex-col space-y-1">
                    <p className="text-sm font-medium leading-none">{user.username}</p>
                    <p className="text-xs leading-none text-muted-foreground">{user.email}</p>
                  </div>
                </DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem onClick={() => navigate("/profile")}>
                  <User className="mr-2 h-4 w-4" />
                  <span>Profil</span>
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => navigate("/settings")}>
                  <Settings className="mr-2 h-4 w-4" />
                  <span>Paramètres</span>
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem onClick={handleLogout}>
                  <LogOut className="mr-2 h-4 w-4" />
                  <span>Déconnexion</span>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          ) : (
            <div className="flex items-center space-x-2">
              <Button variant="ghost" onClick={() => navigate("/login")} type="button">
                Connexion
              </Button>
              <Button onClick={() => navigate("/register")} type="button">
                Inscription
              </Button>
            </div>
          )}
        </div>
      </div>
    </nav>
  );
} 