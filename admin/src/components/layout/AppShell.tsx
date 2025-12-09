import { useState } from 'react';
import { Outlet } from 'react-router-dom';
import { Navbar } from './Navbar';
import { Sidebar } from './Sidebar';

export function AppShell() {
  const [mobileSidebarOpen, setMobileSidebarOpen] = useState(false);

  return (
    <div className="h-full w-full flex flex-col overflow-hidden">
      <Navbar onMenuClick={() => setMobileSidebarOpen(true)} />
      <div className="flex-1 flex overflow-hidden">
        {/* Single Sidebar that adapts to screen size */}
        <Sidebar
          isOpen={mobileSidebarOpen}
          onClose={() => setMobileSidebarOpen(false)}
        />

        <main className="flex-1 overflow-y-auto bg-[rgb(var(--bg-base))]">
          <div className="p-8 max-w-[1600px]">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  );
}
