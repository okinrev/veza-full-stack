import { useState } from 'react';
import { Outlet } from 'react-router-dom';
import { Header } from './Header';
import { Sidebar } from './Sidebar';
import { cn } from '@/shared/utils/helpers';

export const PageLayout = () => {
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [mobileSidebarOpen, setMobileSidebarOpen] = useState(false);

  return (
    <div className="min-h-screen bg-gray-50">
      <Header 
        onMenuClick={() => setMobileSidebarOpen(!mobileSidebarOpen)}
        onToggleSidebar={() => setSidebarOpen(!sidebarOpen)}
      />
      
      <div className="flex h-[calc(100vh-64px)]">
        <Sidebar 
          open={sidebarOpen}
          mobileOpen={mobileSidebarOpen}
          onMobileClose={() => setMobileSidebarOpen(false)}
        />
        
        <main
          className={cn(
            'flex-1 overflow-y-auto transition-all duration-300',
            sidebarOpen ? 'lg:ml-64' : 'lg:ml-16'
          )}
        >
          <div className="container mx-auto px-4 py-6">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  );
}; 