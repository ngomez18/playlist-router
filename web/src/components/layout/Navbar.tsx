import type { User } from '../../types/auth'
import { UserMenu } from '../auth'

interface NavbarProps {
  user: User
  onLogout: () => void
  onNavigateHome?: () => void
}

export function Navbar({ user, onLogout, onNavigateHome }: NavbarProps) {
  return (
    <div className="navbar bg-base-200">
      <div className="flex-1">
        <button 
          className="btn btn-ghost text-xl"
          onClick={onNavigateHome}
        >
          PlaylistSync
        </button>
      </div>
      <div className="flex-none gap-2">
        <UserMenu user={user} onLogout={onLogout} />
      </div>
    </div>
  )
}