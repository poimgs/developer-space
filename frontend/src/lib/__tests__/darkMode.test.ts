import { beforeEach, describe, expect, it } from 'vitest';
import { initTheme } from '../darkMode';

describe('darkMode', () => {
  beforeEach(() => {
    document.documentElement.classList.remove('dark');
  });

  describe('initTheme', () => {
    it('always applies dark class', () => {
      initTheme();
      expect(document.documentElement.classList.contains('dark')).toBe(true);
    });
  });
});
