import { useEffect } from "react"
import { LoginForm } from "../components/auth"
import { setAuthToken } from "../lib/auth"

export function AuthPage() {
  const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || ""
  const loginUrl = `${apiBaseUrl}/auth/spotify/login`

  // Handle OAuth callback - extract token from URL params
  useEffect(() => {
    const urlParams = new URLSearchParams(window.location.search)
    const token = urlParams.get("token")
    
    if (token) {
      // Store token and clean URL
      setAuthToken(token)
      window.history.replaceState({}, document.title, window.location.pathname)
      // Reload to trigger auth validation
      window.location.reload()
    }
  }, [])

  return <LoginForm loginUrl={loginUrl} />
}