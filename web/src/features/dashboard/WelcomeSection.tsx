import type { User } from '../../types/auth'

interface WelcomeSectionProps {
  user: User
}

export function WelcomeSection({ user }: WelcomeSectionProps) {
  return (
    <div className="text-center mb-8">
      <h1 className="text-4xl font-bold mb-4">Welcome, {user.name}!</h1>
      <p className="text-lg">Ready to sync your playlists?</p>
    </div>
  )
}