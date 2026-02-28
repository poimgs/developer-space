import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { initTheme, isDarkMode, toggleTheme } from '../darkMode';

describe('darkMode', () => {
  const originalMatchMedia = window.matchMedia;

  beforeEach(() => {
    document.documentElement.classList.remove('dark');
    localStorage.clear();
  });

  afterEach(() => {
    window.matchMedia = originalMatchMedia;
    vi.restoreAllMocks();
  });

  describe('initTheme', () => {
    it('applies dark class when localStorage is "dark"', () => {
      localStorage.setItem('theme', 'dark');
      initTheme();
      expect(document.documentElement.classList.contains('dark')).toBe(true);
    });

    it('removes dark class when localStorage is "light"', () => {
      document.documentElement.classList.add('dark');
      localStorage.setItem('theme', 'light');
      initTheme();
      expect(document.documentElement.classList.contains('dark')).toBe(false);
    });

    it('applies dark class when system prefers dark', () => {
      window.matchMedia = vi.fn().mockReturnValue({ matches: true });
      initTheme();
      expect(document.documentElement.classList.contains('dark')).toBe(true);
    });

    it('does not apply dark when system prefers light and no storage', () => {
      window.matchMedia = vi.fn().mockReturnValue({ matches: false });
      initTheme();
      expect(document.documentElement.classList.contains('dark')).toBe(false);
    });
  });

  describe('toggleTheme', () => {
    it('toggles to dark and persists', () => {
      toggleTheme();
      expect(document.documentElement.classList.contains('dark')).toBe(true);
      expect(localStorage.getItem('theme')).toBe('dark');
    });

    it('toggles back to light and persists', () => {
      document.documentElement.classList.add('dark');
      toggleTheme();
      expect(document.documentElement.classList.contains('dark')).toBe(false);
      expect(localStorage.getItem('theme')).toBe('light');
    });
  });

  describe('isDarkMode', () => {
    it('returns true when dark class present', () => {
      document.documentElement.classList.add('dark');
      expect(isDarkMode()).toBe(true);
    });

    it('returns false when dark class absent', () => {
      expect(isDarkMode()).toBe(false);
    });
  });
});
