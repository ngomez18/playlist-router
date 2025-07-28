import type { User } from '../../types/auth'
import { Avatar } from '../ui'

interface UserMenuProps {
  user: User
  onLogout: () => void
}

export function UserMenu({ user, onLogout }: UserMenuProps) {
  return (
    <div className="dropdown dropdown-end">
      <div
        tabIndex={0}
        className="btn btn-ghost btn-circle avatar"
      >
        <Avatar name={user.name} />
      </div>
      <ul
        tabIndex={0}
        className="mt-3 z-[1] p-2 shadow menu menu-sm dropdown-content bg-base-100 rounded-box w-52"
      >
        <li>
          <span className="text-sm opacity-70">{user.email}</span>
        </li>
        <li>
          <a onClick={onLogout}>Logout</a>
        </li>
      </ul>
    </div>
  )
}