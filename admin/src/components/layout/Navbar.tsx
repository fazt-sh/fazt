import { Moon, Sun, User, LogOut } from 'lucide-react';
import { useTheme } from '../../context/ThemeContext';
import { useAuth } from '../../context/AuthContext';
import { Dropdown, DropdownItem, DropdownDivider } from '../ui/Dropdown';

export function Navbar() {
  const { theme, toggleTheme } = useTheme();
  const { user, logout } = useAuth();

  return (
    <nav className="h-14 glass border-b border-[rgb(var(--border-primary))] flex items-center justify-between px-6 sticky top-0 z-50">
      <div className="flex items-center gap-4">
        <img src="/logo.png" alt="Fazt" className="h-8 w-8 rounded-lg" />
        <div className="font-display text-lg text-[rgb(var(--text-primary))] tracking-tight">
          Fazt<span className="gradient-text">.sh</span>
        </div>
      </div>

      <div className="flex items-center gap-2">
        {/* Theme Toggle */}
        <button
          onClick={toggleTheme}
          className="p-2 rounded-lg text-[rgb(var(--text-secondary))] hover:text-[rgb(var(--text-primary))]
                     hover:bg-[rgb(var(--bg-hover))] transition-all duration-150"
          aria-label="Toggle theme"
        >
          {theme === 'light' ? (
            <Moon className="h-[18px] w-[18px]" strokeWidth={2} />
          ) : (
            <Sun className="h-[18px] w-[18px]" strokeWidth={2} />
          )}
        </button>

        {/* User Menu */}
        {user && (
          <Dropdown
            trigger={
              <button className="flex items-center gap-2.5 px-2.5 py-1.5 rounded-lg
                               text-[rgb(var(--text-secondary))] hover:text-[rgb(var(--text-primary))]
                               hover:bg-[rgb(var(--bg-hover))] transition-all duration-150
                               border border-transparent hover:border-[rgb(var(--border-primary))]">
                <div className="w-7 h-7 rounded-full bg-gradient-to-br from-[rgb(var(--accent-start))] to-[rgb(var(--accent-mid))]
                              flex items-center justify-center text-white text-xs font-semibold
                              shadow-sm">
                  {user.username[0].toUpperCase()}
                </div>
                <span className="text-[13px] font-medium">
                  {user.username}
                </span>
              </button>
            }
          >
            <DropdownItem icon={<User className="h-4 w-4" />}>
              Profile
            </DropdownItem>
            <DropdownDivider />
            <DropdownItem
              onClick={logout}
              icon={<LogOut className="h-4 w-4" />}
            >
              Logout
            </DropdownItem>
          </Dropdown>
        )}
      </div>
    </nav>
  );
}
