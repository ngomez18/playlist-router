import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient } from '../lib/api'
import { CreateBasePlaylistForm } from '../features/playlists/CreateBasePlaylistForm'
import { TrashIcon, ArrowLeftIcon } from '@heroicons/react/24/outline'
import { Button, Card, CardBody, CardTitle, CardActions, Alert, LoadingSpinner } from '../components/ui'
import type { BasePlaylist } from '../types/playlist'

interface BasePlaylistsPageProps {
  onBack: () => void
}

export function BasePlaylistsPage({ onBack }: BasePlaylistsPageProps) {
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)
  const queryClient = useQueryClient()

  const { data: basePlaylists, isLoading, error } = useQuery({
    queryKey: ['basePlaylists'],
    queryFn: () => apiClient.getUserBasePlaylists(),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => apiClient.deleteBasePlaylist(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['basePlaylists'] })
      setSuccessMessage('Base playlist deleted successfully!')
      setTimeout(() => setSuccessMessage(null), 5000)
    },
    onError: () => {
      setSuccessMessage('Failed to delete base playlist')
      setTimeout(() => setSuccessMessage(null), 5000)
    }
  })

  const handleCreateSuccess = () => {
    setShowCreateForm(false)
    queryClient.invalidateQueries({ queryKey: ['basePlaylists'] })
    setSuccessMessage('Base playlist created successfully!')
    setTimeout(() => setSuccessMessage(null), 5000)
  }

  const handleDelete = (playlist: BasePlaylist) => {
    if (confirm(`Are you sure you want to delete "${playlist.name}"?`)) {
      deleteMutation.mutate(playlist.id)
    }
  }

  if (showCreateForm) {
    return (
      <div className="max-w-md mx-auto">
        <CreateBasePlaylistForm 
          onSuccess={handleCreateSuccess}
          onCancel={() => setShowCreateForm(false)}
        />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Base Playlists</h1>
          <p className="text-gray-600 mt-2">Manage your source playlists that songs will be distributed from</p>
        </div>
        <Button onClick={onBack} variant="outline">
          <ArrowLeftIcon className="h-4 w-4" />
          Back
        </Button>
      </div>

      {successMessage && (
        <Alert type={successMessage.includes('success') ? 'success' : 'error'}>
          {successMessage}
        </Alert>
      )}

      <div className="flex justify-between items-center">
        <h2 className="text-xl font-semibold">Your Base Playlists</h2>
        {basePlaylists && basePlaylists.length > 0 && (
          <Button onClick={() => setShowCreateForm(true)}>
            Create New
          </Button>
        )}
      </div>

      {isLoading && <LoadingSpinner />}

      {error && (
        <Alert type="error">
          Failed to load base playlists. Please try again.
        </Alert>
      )}

      {basePlaylists && (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {basePlaylists.length === 0 ? (
            <div className="col-span-full text-center py-12">
              <p className="text-gray-600 mb-4">You don't have any base playlists yet.</p>
              <Button onClick={() => setShowCreateForm(true)}>
                Create Your First
              </Button>
            </div>
          ) : (
            basePlaylists.map((playlist) => (
              <Card key={playlist.id}>
                <CardBody>
                  <CardTitle>{playlist.name}</CardTitle>
                  <p className="text-sm text-gray-600 mb-2">
                    Spotify ID: {playlist.spotify_playlist_id}
                  </p>
                  <p className="text-sm">
                    Status: <span className={playlist.is_active ? 'text-green-600' : 'text-gray-500'}>
                      {playlist.is_active ? 'Active' : 'Inactive'}
                    </span>
                  </p>
                  <CardActions>
                    <Button 
                      onClick={() => handleDelete(playlist)}
                      variant="outline"
                      size="sm"
                      disabled={deleteMutation.isPending}
                    >
                      {deleteMutation.isPending ? '...' : <TrashIcon className="h-4 w-4" />}
                    </Button>
                  </CardActions>
                </CardBody>
              </Card>
            ))
          )}
        </div>
      )}
    </div>
  )
}