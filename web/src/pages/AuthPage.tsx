import { useEffect } from "react"
import { LoginForm } from "../components/auth"
import { useAuth } from "../hooks/useAuth"

export function AuthPage() {
  const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || ""
  const loginUrl = `${apiBaseUrl}/auth/spotify/login`
  const { login } = useAuth()

  // Handle OAuth callback - extract token from URL params
  useEffect(() => {
    const urlParams = new URLSearchParams(window.location.search)
    const token = urlParams.get("token")
    
    if (token) {
      // Clean URL first
      window.history.replaceState({}, document.title, window.location.pathname)
      // Use the auth context login method which will validate the token
      login(token)
    }
  }, [login])

  return <LoginForm loginUrl={loginUrl} />
}