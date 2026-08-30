import { create } from "zustand";
import { createJSONStorage, persist, type StateStorage } from "zustand/middleware";
import type { UserInfo } from "@/types/auth";

const STORAGE_KEY = "cankora-auth";

interface AuthState {
  user: UserInfo | null;
  accessToken: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  rememberMe: boolean;

  // Actions
  setAuth: (
    user: UserInfo,
    accessToken: string,
    refreshToken: string,
    rememberMe: boolean
  ) => void;
  setTokens: (accessToken: string, refreshToken: string) => void;
  clearAuth: () => void;
  hasPermission: (permission: string) => boolean;
  hasRole: (role: string) => boolean;
}

interface PersistedAuthState {
  state?: {
    rememberMe?: boolean;
  };
}

function parsePersistedAuth(value: string): PersistedAuthState | null {
  try {
    const parsed: unknown = JSON.parse(value);
    if (!parsed || typeof parsed !== "object") return null;
    return parsed as PersistedAuthState;
  } catch {
    return null;
  }
}

function writeAuthCookie(value: string, rememberMe: boolean) {
  if (typeof document === "undefined") return;

  const maxAge = rememberMe ? "; Max-Age=604800" : "";
  document.cookie = `${STORAGE_KEY}=${encodeURIComponent(value)}; Path=/; SameSite=Lax${maxAge}`;
}

function clearAuthCookie() {
  if (typeof document === "undefined") return;
  document.cookie = `${STORAGE_KEY}=; Path=/; Max-Age=0; SameSite=Lax`;
}

const authStorage: StateStorage = {
  getItem: (name) => {
    if (typeof window === "undefined") return null;
    return localStorage.getItem(name) ?? sessionStorage.getItem(name);
  },
  setItem: (name, value) => {
    if (typeof window === "undefined") return;

    const persistedAuth = parsePersistedAuth(value);
    if (persistedAuth?.state?.rememberMe) {
      localStorage.setItem(name, value);
      sessionStorage.removeItem(name);
      writeAuthCookie(value, true);
      return;
    }

    sessionStorage.setItem(name, value);
    localStorage.removeItem(name);
    writeAuthCookie(value, false);
  },
  removeItem: (name) => {
    if (typeof window === "undefined") return;
    localStorage.removeItem(name);
    sessionStorage.removeItem(name);
    clearAuthCookie();
  },
};

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false,
      rememberMe: true,

      setAuth: (user, accessToken, refreshToken, rememberMe) =>
        set({ user, accessToken, refreshToken, rememberMe, isAuthenticated: true }),

      setTokens: (accessToken, refreshToken) =>
        set({ accessToken, refreshToken }),

      clearAuth: () =>
        set({
          user: null,
          accessToken: null,
          refreshToken: null,
          isAuthenticated: false,
          rememberMe: true,
        }),

      hasPermission: (permission: string) => {
        const { user } = get();
        if (!user) return false;
        // Super admin check via role
        if (user.roles?.includes("SUPER_ADMIN") || user.roles?.includes("ADMIN")) {
          return true;
        }
        return user.permissions?.includes(permission) ?? false;
      },

      hasRole: (role: string) => {
        const { user } = get();
        return user?.roles?.includes(role) ?? false;
      },
    }),
    {
      name: STORAGE_KEY,
      storage: createJSONStorage(() => authStorage),
      // Only persist tokens + user, not derived state
      partialize: (state) => ({
        user: state.user,
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        isAuthenticated: state.isAuthenticated,
        rememberMe: state.rememberMe,
      }),
    }
  )
);
