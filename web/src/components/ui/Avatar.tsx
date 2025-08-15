import { getInitials } from '../../lib/utils'

interface AvatarProps {
  name: string
}

export function Avatar({ name }: AvatarProps) {
  return (
    <div className="avatar avatar-placeholder">
      <div className="bg-accent text-accent-content w-8 rounded-full">
        <span>{getInitials(name)}</span>
      </div>
    </div>
  );
}