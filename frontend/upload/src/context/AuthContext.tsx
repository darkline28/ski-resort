import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from 'react';
import { authApi } from '@/api';
import { setUnauthorizedHandler, tokenStorage } from '@/api/client';
import type { LoginPayload, RegisterPayload, User } from '@/types';

interface AuthState {
  user: User | null;
  /** true, пока проверяем сохранённый токен при старте */
  initializing: boolean;
  login: (p: LoginPayload) => Promise<User>;
  register: (p: RegisterPayload) => Promise<User>;
  logout: () => Promise<void>;
  setUser: (u: User) => void;
}

const AuthContext = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [initializing, setInitializing] = useState(true);

  useEffect(() => {
    setUnauthorizedHandler(() => {
      tokenStorage.clear();
      setUser(null);
    });
    if (!tokenStorage.get()) {
      setInitializing(false);
      return;
    }
    authApi.me()
      .then(setUser)
      .catch(() => tokenStorage.clear())
      .finally(() => setInitializing(false));
    return () => setUnauthorizedHandler(null);
  }, []);

  const login = useCallback(async (p: LoginPayload) => {
    const res = await authApi.login(p);
    tokenStorage.set(res.token);
    setUser(res.user);
    return res.user;
  }, []);

  const register = useCallback(async (p: RegisterPayload) => {
    const res = await authApi.register(p);
    tokenStorage.set(res.token);
    setUser(res.user);
    return res.user;
  }, []);

  const logout = useCallback(async () => {
    try { await authApi.logout(); } catch { /* токен всё равно удаляем */ }
    tokenStorage.clear();
    setUser(null);
  }, []);

  const value = useMemo(
    () => ({ user, initializing, login, register, logout, setUser }),
    [user, initializing, login, register, logout],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used inside AuthProvider');
  return ctx;
}
