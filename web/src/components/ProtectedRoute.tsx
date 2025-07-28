import type { ReactNode } from 'react'
import { useAuth } from '../hooks/useAuth'

interface ProtectedRouteProps {
  children: ReactNode
}

export function ProtectedRoute({ children }: ProtectedRouteProps) {
  const { isAuthenticated, isLoading } = useAuth()

  if (isLoading) {
    return (
      <div className="min-h-screen bg-base-100 flex items-center justify-center">
        <div className="loading loading-spinner loading-lg"></div>
      </div>
    )
  }

  if (!isAuthenticated) {
    return (
      <div className="min-h-screen bg-base-100 flex items-center justify-center">
        <div className="text-center max-w-md">
          <h1 className="text-3xl font-bold mb-4">Access Denied</h1>
          <p className="text-lg mb-6">You need to be logged in to access this page.</p>
          <a href="/auth/spotify/login" className="btn btn-primary">
            <img src="/spotify-icon-dark.svg" alt="Spotify" className="w-5 h-5" />
            Log in with Spotify
          </a>
        </div>
      </div>
    )
  }

  return <>{children}</>
}