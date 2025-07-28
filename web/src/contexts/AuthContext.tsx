import { useEffect, useState } from 'react'
import type { ReactNode } from 'react'
import type { User } from '../types/auth'
import { getAuthToken, setAuthToken, removeAuthToken } from '../lib/auth'
import { AuthContext, type AuthContextType } from './auth-context'

interface AuthProviderProps {
  children: ReactNode
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  const login = (token: string) => {
    setAuthToken(token)
    // User will be set when the token is validated
  }

  const logout = () => {
    removeAuthToken()
    setUser(null)
    window.location.reload()
  }

  useEffect(() => {
    // Check for existing token on mount
    const token = getAuthToken()
    if (token) {
      // Token validation will be handled by the API client
      setIsLoading(false)
    } else {
      setIsLoading(false)
    }
  }, [])

  const value: AuthContextType = {
    user,
    isAuthenticated: !!user,
    isLoading,
    login,
    logout,
    setUser,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

