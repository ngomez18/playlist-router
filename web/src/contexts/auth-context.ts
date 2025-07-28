import { createContext } from 'react'
import type { User, AuthState } from '../types/auth'

export interface AuthContextType extends AuthState {
  login: (token: string) => void
  logout: () => void
  setUser: (user: User | null) => void
}

export const AuthContext = createContext<AuthContextType | undefined>(undefined)