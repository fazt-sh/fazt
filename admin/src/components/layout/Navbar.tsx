import { Moon, Sun, User, LogOut } from 'lucide-react';
import { useTheme } from '../../context/ThemeContext';
import { useAuth } from '../../context/AuthContext';
import { Dropdown, DropdownItem, DropdownDivider } from '../ui/Dropdown';
import { Button } from '../ui/Button';

export function Navbar() {
  const { theme, toggleTheme } = useTheme();
  const { user, logout } = useAuth();

  return (
    <nav className="h-16 bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between px-6">
      <div className="flex items-center gap-3">
        <div className="text-xl font-bold text-gray-900 dark:text-gray-50">
          Fazt<span className="text-primary">.sh</span>
        </div>
      </div>

      <div className="flex items-center gap-4">
        {/* Theme Toggle */}
        <Button
          variant="ghost"
          size="sm"
          onClick={toggleTheme}
          className="rounded-full p-2"
          aria-label="Toggle theme"
        >
          {theme === 'light' ? (
            <Moon className="h-5 w-5" />
          ) : (
            <Sun className="h-5 w-5" />
          )}
        </Button>

        {/* User Menu */}
        {user && (
          <Dropdown
            trigger={
              <button className="flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
                <div className="w-8 h-8 rounded-full bg-primary flex items-center justify-center text-white font-medium">
                  {user.username[0].toUpperCase()}
                </div>
                <span className="text-sm font-medium text-gray-700 dark:text-gray-200">
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
