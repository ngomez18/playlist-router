import type { ChildPlaylist } from '../../types/playlist'
import { Card, CardBody, CardTitle, CardActions } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'

interface ChildPlaylistCardProps {
  childPlaylist: ChildPlaylist
  onEdit?: (playlist: ChildPlaylist) => void
  onDelete?: (id: string) => void
  loading?: boolean
}

export function ChildPlaylistCard({ 
  childPlaylist, 
  onEdit, 
  onDelete,
  loading = false
}: ChildPlaylistCardProps) {
  const createdDate = new Date(childPlaylist.created).toLocaleDateString()
  const hasFilters = childPlaylist.filter_rules && Object.keys(childPlaylist.filter_rules).length > 0

  return (
    <Card className="w-full">
      <CardBody>
        <CardTitle className="flex items-center justify-between">
          <span className="truncate">{childPlaylist.name}</span>
          <div className="flex items-center gap-2">
            {/* Status Badge */}
            <span className={`badge ${
              childPlaylist.is_active 
                ? 'badge-success' 
                : 'badge-neutral'
            }`}>
              {childPlaylist.is_active ? 'Active' : 'Inactive'}
            </span>
          </div>
        </CardTitle>
        
        {/* Description */}
        {childPlaylist.description && (
          <p className="text-sm text-base-content/70 line-clamp-2">
            {childPlaylist.description}
          </p>
        )}
        
        {/* Metadata */}
        <div className="text-xs text-base-content/50 mt-2">
          <p>Created: {createdDate}</p>
          <p className="truncate">Spotify ID: {childPlaylist.spotify_playlist_id}</p>
        </div>

        {/* Filter Rules Summary */}
        {hasFilters && (
          <div className="mt-3">
            <div className="collapse collapse-arrow border border-base-300 bg-base-200 rounded-lg">
              <input type="checkbox" className="peer" />
              <div className="collapse-title text-sm font-medium">
                Audio Filters ({Object.keys(childPlaylist.filter_rules || {}).length})
              </div>
              <div className="collapse-content">
                <div className="grid grid-cols-2 gap-2 text-xs">
                  {Object.entries(childPlaylist.filter_rules || {}).map(([key, value]) => (
                    <div key={key} className="flex justify-between">
                      <span className="capitalize">{key.replace('_', ' ')}:</span>
                      <span className="text-base-content/70">
                        {typeof value === 'object' && value !== null ? (
                          'min' in value || 'max' in value ? (
                            (() => {
                              const range = value as any
                              let minVal = range.min !== undefined ? range.min : 'N/A'
                              let maxVal = range.max !== undefined ? range.max : 'N/A'
                              
                              // Convert duration from ms to seconds for display
                              if (key === 'duration_ms') {
                                minVal = minVal !== 'N/A' ? Math.round(minVal / 1000) + 's' : 'N/A'
                                maxVal = maxVal !== 'N/A' ? Math.round(maxVal / 1000) + 's' : 'N/A'
                              }
                              
                              return `${minVal} - ${maxVal}`
                            })()
                          ) : (
                            'Custom'
                          )
                        ) : (
                          String(value)
                        )}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Actions */}
        {(onEdit || onDelete) && (
          <CardActions className="mt-4">
            {onEdit && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => onEdit(childPlaylist)}
                disabled={loading}
                className="btn-sm"
              >
                Edit
              </Button>
            )}
            {onDelete && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => onDelete(childPlaylist.id)}
                disabled={loading}
                className="btn-sm btn-error"
              >
                Delete
              </Button>
            )}
          </CardActions>
        )}
      </CardBody>
    </Card>
  )
}