import { useState, useCallback } from 'react';

const STORAGE_KEY = 'bp-admin-key';

export function useAuth() {
  const [key, setKey] = useState<string>(() => localStorage.getItem(STORAGE_KEY) || '');

  const login = useCallback((adminKey: string) => {
    localStorage.setItem(STORAGE_KEY, adminKey);
    setKey(adminKey);
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem(STORAGE_KEY);
    setKey('');
  }, []);

  return { isAuthenticated: !!key, login, logout };
}
