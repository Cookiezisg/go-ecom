import { create } from "zustand";
import type { UserProfile } from "@/types";

interface AuthState {
  token: string;
  profile: UserProfile | null;
  setAuth: (token: string, profile: UserProfile) => void;
  logout: () => void;
}

const TOKEN_KEY = "go-ecom-user-token";
const PROFILE_KEY = "go-ecom-user-profile";

function loadProfile(): UserProfile | null {
  const raw = window.localStorage.getItem(PROFILE_KEY);
  if (!raw) {
    return null;
  }

  try {
    return JSON.parse(raw) as UserProfile;
  } catch {
    return null;
  }
}

export const useAuthStore = create<AuthState>((set) => ({
  token: window.localStorage.getItem(TOKEN_KEY) ?? "",
  profile: loadProfile(),
  setAuth: (token, profile) => {
    window.localStorage.setItem(TOKEN_KEY, token);
    window.localStorage.setItem(PROFILE_KEY, JSON.stringify(profile));
    set({ token, profile });
  },
  logout: () => {
    window.localStorage.removeItem(TOKEN_KEY);
    window.localStorage.removeItem(PROFILE_KEY);
    set({ token: "", profile: null });
  },
}));
