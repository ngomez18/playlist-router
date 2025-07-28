export interface User {
  id: string
  email: string
  username: string
  name: string
  created: string
  updated: string
}

export interface AuthState {
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
}