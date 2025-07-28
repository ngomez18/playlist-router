import { useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useAuth } from './useAuth'
import { apiClient } from '../lib/api'
import { getAuthToken } from '../lib/auth'

export function useAuthValidation() {
  const { setUser, logout } = useAuth()
  const token = getAuthToken()

  const { data: user, error, isLoading } = useQuery({
    queryKey: ['auth', 'validate'],
    queryFn: () => apiClient.validateToken(),
    enabled: !!token,
    retry: false,
    staleTime: 5 * 60 * 1000, // 5 minutes
  })

  useEffect(() => {
    if (user) {
      setUser(user)
    } else if (error && token) {
      // Token is invalid, logout (will reload page)
      logout()
    }
  }, [user, error, token, setUser, logout])

  return { isLoading }
}