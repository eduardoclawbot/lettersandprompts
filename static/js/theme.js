// Theme switcher - vanilla JS, no dependencies
(function() {
  'use strict';

  const STORAGE_KEY = 'theme';
  const THEME_LIGHT = 'light';
  const THEME_DARK = 'dark';

  // Get stored theme or default to dark
  function getStoredTheme() {
    try {
      return localStorage.getItem(STORAGE_KEY) || THEME_DARK;
    } catch (e) {
      return THEME_DARK;
    }
  }

  // Store theme preference
  function storeTheme(theme) {
    try {
      localStorage.setItem(STORAGE_KEY, theme);
    } catch (e) {
      // localStorage might be unavailable (private browsing, etc.)
    }
  }

  // Apply theme to document
  function applyTheme(theme) {
    if (theme === THEME_LIGHT) {
      document.documentElement.setAttribute('data-theme', 'light');
    } else {
      document.documentElement.removeAttribute('data-theme');
    }
  }

  // Toggle between themes
  function toggleTheme() {
    const currentTheme = document.documentElement.getAttribute('data-theme') === 'light' 
      ? THEME_LIGHT 
      : THEME_DARK;
    const newTheme = currentTheme === THEME_LIGHT ? THEME_DARK : THEME_LIGHT;
    
    applyTheme(newTheme);
    storeTheme(newTheme);
  }

  // Apply stored theme immediately (before page renders) to prevent flash
  applyTheme(getStoredTheme());

  // Set up toggle button when DOM is ready
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initToggle);
  } else {
    initToggle();
  }

  function initToggle() {
    const toggle = document.getElementById('theme-toggle');
    if (toggle) {
      toggle.addEventListener('click', toggleTheme);
    }
  }
})();
