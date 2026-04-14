import { create } from "zustand";

interface AdminAuthState {
  token: string;
  username: string;
  setAuth: (token: string, username: string) => void;
  logout: () => void;
}

const TOKEN_KEY = "go-ecom-admin-token";
const USERNAME_KEY = "go-ecom-admin-username";

export const useAdminAuthStore = create<AdminAuthState>((set) => ({
  token: window.localStorage.getItem(TOKEN_KEY) ?? "",
  username: window.localStorage.getItem(USERNAME_KEY) ?? "",
  setAuth: (token, username) => {
    window.localStorage.setItem(TOKEN_KEY, token);
    window.localStorage.setItem(USERNAME_KEY, username);
    set({ token, username });
  },
  logout: () => {
    window.localStorage.removeItem(TOKEN_KEY);
    window.localStorage.removeItem(USERNAME_KEY);
    set({ token: "", username: "" });
  },
}));
