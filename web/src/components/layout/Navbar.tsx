import type { User } from '../../types/auth'
import { UserMenu } from '../auth'

interface NavbarProps {
  user: User
  onLogout: () => void
}

export function Navbar({ user, onLogout }: NavbarProps) {
  return (
    <div className="navbar bg-base-200">
      <div className="flex-1">
        <a className="btn btn-ghost text-xl">PlaylistSync</a>
      </div>
      <div className="flex-none gap-2">
        <UserMenu user={user} onLogout={onLogout} />
      </div>
    </div>
  )
}