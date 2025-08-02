import type { ChildPlaylist } from '../../types/playlist'
import { ChildPlaylistCard } from './ChildPlaylistCard'
import { LoadingSpinner } from '../../components/ui/LoadingSpinner'
import { Alert } from '../../components/ui/Alert'

interface ChildPlaylistListProps {
  childPlaylists: ChildPlaylist[]
  loading?: boolean
  error?: string | null
  onEdit?: (playlist: ChildPlaylist) => void
  onDelete?: (id: string) => void
  className?: string
}

export function ChildPlaylistList({ 
  childPlaylists, 
  loading = false,
  error = null,
  onEdit, 
  onDelete,
  className = ''
}: ChildPlaylistListProps) {
  // Loading state
  if (loading) {
    return (
      <div className={`flex justify-center items-center py-8 ${className}`}>
        <LoadingSpinner size="lg" />
      </div>
    )
  }

  // Error state
  if (error) {
    return (
      <div className={className}>
        <Alert type="error">
          <span>Failed to load child playlists: {error}</span>
        </Alert>
      </div>
    )
  }

  // Empty state
  if (!childPlaylists || childPlaylists.length === 0) {
    return (
      <div className={`text-center py-8 ${className}`}>
        <div className="hero-content text-center">
          <div className="max-w-md">
            <div className="text-6xl mb-4">🎵</div>
            <h3 className="text-lg font-semibold mb-2">No Child Playlists Yet</h3>
            <p className="text-base-content/70">
              Create your first child playlist to start organizing your music with custom filters and rules.
            </p>
          </div>
        </div>
      </div>
    )
  }

  // List with data
  return (
    <div className={`space-y-4 ${className}`}>
      {/* Status Summary */}
      <div className="flex gap-2 text-sm justify-end">
        <span className="badge badge-success badge-sm">
          {childPlaylists.filter(p => p.is_active).length} Active
        </span>
        <span className="badge badge-neutral badge-sm">
          {childPlaylists.filter(p => !p.is_active).length} Inactive
        </span>
      </div>

      {/* Playlist Grid */}
      <div className="grid gap-4 sm:grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
        {childPlaylists.map((playlist) => (
          <ChildPlaylistCard
            key={playlist.id}
            childPlaylist={playlist}
            onEdit={onEdit}
            onDelete={onDelete}
            loading={loading}
          />
        ))}
      </div>
    </div>
  )
}