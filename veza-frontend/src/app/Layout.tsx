import { Outlet } from 'react-router-dom';
import { Navbar } from '@/shared/components/Navbar';
import { Sidebar } from '@/shared/components/Sidebar';
import { Toaster } from '@/components/ui/toaster';

export function Layout() {
  return (
    <div className="min-h-screen bg-background">
      <Navbar />
      <div className="flex">
        <Sidebar />
        <main className="flex-1">
          <Outlet />
        </main>
      </div>
      <Toaster />
    </div>
  );
} 