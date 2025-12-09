import { Outlet } from 'react-router-dom';
import { Navbar } from './Navbar';
import { Sidebar } from './Sidebar';

export function AppShell() {
  return (
    <div className="h-full w-full flex flex-col overflow-hidden">
      <Navbar />
      <div className="flex-1 flex overflow-hidden">
        <Sidebar />
        <main className="flex-1 overflow-y-auto bg-[rgb(var(--bg-base))]">
          <div className="p-8 max-w-[1600px]">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  );
}
