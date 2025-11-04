import { useEffect, useState, useCallback } from 'react'
import type { ReactNode } from 'react'
import type { User } from '../types/auth'
import { getAuthToken, setAuthToken, removeAuthToken } from '../lib/auth'
import { AuthContext, type AuthContextType } from './auth-context'
import { apiClient } from '../lib/api'
import { fullStory } from "../lib/fullstory";

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const validateToken = useCallback(async () => {
    try {
      const userData = await apiClient.validateToken();
      setUser(userData);

      fullStory.identify(userData.id);

      setIsLoading(false);
    } catch {
      // Token is invalid, remove it
      removeAuthToken();
      setUser(null);
      setIsLoading(false);
    }
  }, []);

  const login = (token: string) => {
    setAuthToken(token);
    // Validate token and set user
    validateToken();
  };

  const logout = () => {
    removeAuthToken();
    setUser(null);
    window.location.reload();
  };

  useEffect(() => {
    // Check for existing token on mount and validate it
    const token = getAuthToken();
    if (token) {
      validateToken();
    } else {
      setIsLoading(false);
    }
  }, [validateToken]);

  const value: AuthContextType = {
    user,
    isAuthenticated: !!user,
    isLoading,
    login,
    logout,
    setUser,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

