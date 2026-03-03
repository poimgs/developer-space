import { useState } from 'react';
import { Link, Outlet, useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import ThemeToggle from './ThemeToggle';

function NavLink({ to, children }: { to: string; children: React.ReactNode }) {
  const { pathname } = useLocation();
  const active = pathname === to || pathname.startsWith(to + '/');
  return (
    <Link
      to={to}
      className={`rounded-md px-3 py-2 text-sm font-medium transition-colors ${
        active
          ? 'bg-amber-50 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300'
          : 'text-stone-700 hover:bg-stone-100 dark:text-stone-300 dark:hover:bg-stone-800'
      }`}
    >
      {children}
    </Link>
  );
}

export default function Layout() {
  const { user, logout } = useAuth();
  const [menuOpen, setMenuOpen] = useState(false);

  return (
    <div className="min-h-screen">
      <nav className="border-b border-stone-200 bg-white dark:border-stone-700 dark:bg-stone-900">
        <div className="mx-auto flex h-14 max-w-5xl items-center justify-between px-4">
          <div className="flex items-center gap-4">
            <Link to="/" className="text-lg font-bold text-stone-900 dark:text-stone-100">
              Developer Space
            </Link>
            <div className="hidden items-center gap-1 md:flex">
              <NavLink to="/">Sessions</NavLink>
              {user?.is_admin && <NavLink to="/members">Members</NavLink>}
            </div>
          </div>

          <div className="hidden items-center gap-2 md:flex">
            <ThemeToggle />
            <NavLink to="/profile">Profile</NavLink>
            <button
              onClick={logout}
              className="rounded-md px-3 py-2 text-sm font-medium text-stone-700 transition-colors hover:bg-stone-100 dark:text-stone-300 dark:hover:bg-stone-800"
            >
              Logout
            </button>
          </div>

          {/* Mobile hamburger */}
          <button
            onClick={() => setMenuOpen(!menuOpen)}
            className="rounded-md p-2 text-stone-500 hover:bg-stone-100 md:hidden dark:text-stone-400 dark:hover:bg-stone-800"
            aria-label="Toggle menu"
          >
            <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              {menuOpen ? (
                <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
              ) : (
                <path strokeLinecap="round" strokeLinejoin="round" d="M4 6h16M4 12h16M4 18h16" />
              )}
            </svg>
          </button>
        </div>

        {/* Mobile menu overlay */}
        {menuOpen && (
          <div className="border-t border-stone-200 px-4 py-3 space-y-1 md:hidden dark:border-stone-700">
            <Link to="/" onClick={() => setMenuOpen(false)} className="block rounded-md px-3 py-2 text-sm font-medium text-stone-700 hover:bg-stone-100 dark:text-stone-300 dark:hover:bg-stone-800">
              Sessions
            </Link>
            {user?.is_admin && (
              <Link to="/members" onClick={() => setMenuOpen(false)} className="block rounded-md px-3 py-2 text-sm font-medium text-stone-700 hover:bg-stone-100 dark:text-stone-300 dark:hover:bg-stone-800">
                Members
              </Link>
            )}
            <Link to="/profile" onClick={() => setMenuOpen(false)} className="block rounded-md px-3 py-2 text-sm font-medium text-stone-700 hover:bg-stone-100 dark:text-stone-300 dark:hover:bg-stone-800">
              Profile
            </Link>
            <div className="flex items-center justify-between px-3 py-2">
              <span className="text-sm text-stone-500 dark:text-stone-400">Theme</span>
              <ThemeToggle />
            </div>
            <button
              onClick={() => { setMenuOpen(false); logout(); }}
              className="block w-full rounded-md px-3 py-2 text-left text-sm font-medium text-stone-700 hover:bg-stone-100 dark:text-stone-300 dark:hover:bg-stone-800"
            >
              Logout
            </button>
          </div>
        )}
      </nav>

      <main className="mx-auto max-w-5xl px-4 py-6">
        <Outlet />
      </main>
    </div>
  );
}
