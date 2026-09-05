"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";

import { useIdleLogout } from "@/features/auth/useIdleLogout";
import { clearAllQueuedTransfers } from "@/features/wallet/queue";
import { api, ApiError } from "@/lib/api";
import { clearAllTransferDrafts, getAccessToken, setAccessToken } from "@/lib/storage";
import type { MePayload, TokenPayload, UserPublic } from "@/lib/types";

type AuthContextValue = {
  user: UserPublic | null;
  message: string;
  ready: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, confirmPassword: string) => Promise<void>;
  logout: () => void;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<UserPublic | null>(null);
  const [message, setMessage] = useState("");
  const [ready, setReady] = useState(false);

  const logout = useCallback(() => {
    setAccessToken(null);
    clearAllTransferDrafts();
    void clearAllQueuedTransfers();
    setUser(null);
    setMessage("");
  }, []);

  const applySession = useCallback((payload: TokenPayload) => {
    setAccessToken(payload.accessToken);
    setUser(payload.user);
  }, []);

  const hydrate = useCallback(async () => {
    if (!getAccessToken()) {
      setReady(true);
      return;
    }

    try {
      const me = await api<MePayload>("/api/v1/me");
      setUser({ id: me.id, email: me.email });
      setMessage(me.message);
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        logout();
      }
    } finally {
      setReady(true);
    }
  }, [logout]);

  useEffect(() => {
    void hydrate();
  }, [hydrate]);

  const refreshSession = useCallback(async () => {
    if (!getAccessToken()) {
      return;
    }

    try {
      const payload = await api<TokenPayload>("/api/v1/auth/refresh", { method: "POST" });
      applySession(payload);
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        logout();
      }
    }
  }, [applySession, logout]);

  useIdleLogout({
    enabled: Boolean(user),
    onIdle: logout,
    onActivity: refreshSession,
  });

  const login = useCallback(
    async (email: string, password: string) => {
      const payload = await api<TokenPayload>("/api/v1/auth/login", {
        method: "POST",
        auth: false,
        body: { email, password },
      });
      applySession(payload);
      const me = await api<MePayload>("/api/v1/me");
      setMessage(me.message);
    },
    [applySession],
  );

  const register = useCallback(
    async (email: string, password: string, confirmPassword: string) => {
      const payload = await api<TokenPayload>("/api/v1/auth/register", {
        method: "POST",
        auth: false,
        body: { email, password, confirmPassword },
      });
      applySession(payload);
      const me = await api<MePayload>("/api/v1/me");
      setMessage(me.message);
    },
    [applySession],
  );

  const value = useMemo(
    () => ({ user, message, ready, login, register, logout }),
    [user, message, ready, login, register, logout],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within AuthProvider");
  }

  return ctx;
}
