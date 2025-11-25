import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient } from '../lib/api'
import { CreateBasePlaylistForm } from '../features/playlists/CreateBasePlaylistForm'
import { ChildPlaylistList } from '../features/playlists/ChildPlaylistList'
import { CreateChildPlaylistForm } from '../features/playlists/CreateChildPlaylistForm'
import { EditChildPlaylistForm } from '../features/playlists/EditChildPlaylistForm'
import { TrashIcon, ArrowLeftIcon, ChevronRightIcon, ChevronDownIcon, PlusIcon, ArrowPathIcon } from '@heroicons/react/24/outline'
import { Button, Card, CardBody, CardTitle, CardActions, Alert, LoadingSpinner } from '../components/ui'
import type { BasePlaylist, ChildPlaylist } from '../types/playlist'

interface BasePlaylistsPageProps {
  onBack: () => void
}

export function BasePlaylistsPage({ onBack }: BasePlaylistsPageProps) {
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)
  const [expandedPlaylist, setExpandedPlaylist] = useState<string | null>(null)
  const [showCreateChildForm, setShowCreateChildForm] = useState<string | null>(null)
  const [editingChildPlaylist, setEditingChildPlaylist] = useState<ChildPlaylist | null>(null)
  const queryClient = useQueryClient()

  const { data: basePlaylists, isLoading, error } = useQuery({
    queryKey: ['basePlaylists'],
    queryFn: () => apiClient.getUserBasePlaylists(),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => apiClient.deleteBasePlaylist(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["basePlaylists"] });
      setSuccessMessage("Base playlist deleted successfully!");
      setTimeout(() => setSuccessMessage(null), 5000);
    },
    onError: () => {
      setSuccessMessage("Failed to delete base playlist");
      setTimeout(() => setSuccessMessage(null), 5000);
    },
  });

  const deleteChildMutation = useMutation({
    mutationFn: (id: string) => apiClient.deleteChildPlaylist(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["basePlaylists"] });
      setSuccessMessage("Child playlist deleted successfully!");
      setTimeout(() => setSuccessMessage(null), 5000);
    },
    onError: () => {
      setSuccessMessage("Failed to delete child playlist");
      setTimeout(() => setSuccessMessage(null), 5000);
    },
  });

  const syncMutation = useMutation({
    mutationFn: (basePlaylistId: string) =>
      apiClient.syncBasePlaylist(basePlaylistId),
    onSuccess: () => {
      setSuccessMessage("Playlist synced successfully!");
      setTimeout(() => setSuccessMessage(null), 5000);
    },
    onError: (error) => {
      if (error.message.includes("409")) {
        setSuccessMessage("Sync already in progress for this playlist");
      } else {
        setSuccessMessage("Failed to sync playlist");
      }
      setTimeout(() => setSuccessMessage(null), 5000);
    },
  });

  const handleCreateSuccess = () => {
    setShowCreateForm(false);
    queryClient.invalidateQueries({ queryKey: ["basePlaylists"] });
    setSuccessMessage("Base playlist created successfully!");
    setTimeout(() => setSuccessMessage(null), 5000);
  };

  const handleDelete = (playlist: BasePlaylist) => {
    if (confirm(`Are you sure you want to delete "${playlist.name}"?`)) {
      deleteMutation.mutate(playlist.id);
    }
  };

  const handleChildDelete = (childId: string) => {
    if (confirm("Are you sure you want to delete this child playlist?")) {
      deleteChildMutation.mutate(childId);
    }
  };

  const handleChildEdit = (childPlaylist: ChildPlaylist) => {
    setEditingChildPlaylist(childPlaylist);
    // Ensure the base playlist is expanded
    if (expandedPlaylist !== childPlaylist.base_playlist_id) {
      setExpandedPlaylist(childPlaylist.base_playlist_id);
    }
  };

  const togglePlaylistExpansion = (playlistId: string) => {
    setExpandedPlaylist(expandedPlaylist === playlistId ? null : playlistId);
  };

  const handleCreateChild = (basePlaylistId: string) => {
    setShowCreateChildForm(basePlaylistId);
    // Ensure the base playlist is expanded to show the new child playlist
    if (expandedPlaylist !== basePlaylistId) {
      setExpandedPlaylist(basePlaylistId);
    }
  };

  const handleCreateChildSuccess = () => {
    setShowCreateChildForm(null);
    queryClient.invalidateQueries({ queryKey: ["basePlaylists"] });
    setSuccessMessage("Child playlist created successfully!");
    setTimeout(() => setSuccessMessage(null), 5000);
  };

  const handleCreateChildCancel = () => {
    setShowCreateChildForm(null);
  };

  const handleEditChildSuccess = () => {
    setEditingChildPlaylist(null);
    queryClient.invalidateQueries({ queryKey: ["basePlaylists"] });
    setSuccessMessage("Child playlist updated successfully!");
    setTimeout(() => setSuccessMessage(null), 5000);
  };

  const handleEditChildCancel = () => {
    setEditingChildPlaylist(null);
  };

  const handleSync = (basePlaylistId: string) => {
    if (
      confirm(
        "Are you sure you want to sync this playlist? This will distribute new songs to child playlists."
      )
    ) {
      syncMutation.mutate(basePlaylistId);
    }
  };

  if (showCreateForm) {
    return (
      <div className="max-w-md mx-auto">
        <CreateBasePlaylistForm
          onSuccess={handleCreateSuccess}
          onCancel={() => setShowCreateForm(false)}
        />
      </div>
    );
  }

  if (showCreateChildForm) {
    return (
      <div className="max-w-md mx-auto">
        <CreateChildPlaylistForm
          basePlaylistId={showCreateChildForm}
          onSuccess={handleCreateChildSuccess}
          onCancel={handleCreateChildCancel}
        />
      </div>
    );
  }

  if (editingChildPlaylist) {
    return (
      <div className="max-w-md mx-auto">
        <EditChildPlaylistForm
          childPlaylist={editingChildPlaylist}
          onSuccess={handleEditChildSuccess}
          onCancel={handleEditChildCancel}
        />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Base Playlists</h1>
          <p className="text-gray-600 mt-2">
            Manage your source playlists that songs will be distributed from
          </p>
        </div>
        <Button onClick={onBack} variant="outline">
          <ArrowLeftIcon className="h-4 w-4" />
          Back
        </Button>
      </div>

      {successMessage && (
        <Alert type={successMessage.includes("success") ? "success" : "error"}>
          {successMessage}
        </Alert>
      )}

      <div className="flex justify-between items-center">
        <h2 className="text-xl font-semibold">Your Base Playlists</h2>
        {basePlaylists && basePlaylists.length > 0 && (
          <Button onClick={() => setShowCreateForm(true)}>Create New</Button>
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
              <p className="text-gray-600 mb-4">
                You don't have any base playlists yet.
              </p>
              <Button onClick={() => setShowCreateForm(true)}>
                Create Your First
              </Button>
            </div>
          ) : (
            basePlaylists.map((playlist) => {
              const isExpanded = expandedPlaylist === playlist.id;
              return (
                <div key={playlist.id} className="col-span-full">
                  <Card className="bg-base-200 shadow-md border border-base-300">
                    <CardBody>
                      <div className="flex items-center justify-between">
                        <div className="flex-1">
                          <CardTitle className="flex items-center gap-2">
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() =>
                                togglePlaylistExpansion(playlist.id)
                              }
                              className="p-1 h-6 w-6"
                            >
                              {isExpanded ? (
                                <ChevronDownIcon className="h-4 w-4" />
                              ) : (
                                <ChevronRightIcon className="h-4 w-4" />
                              )}
                            </Button>
                            {playlist.name}
                          </CardTitle>
                          <p className="text-sm text-gray-600 mb-1">
                            Spotify ID: {playlist.spotify_playlist_id}
                          </p>
                          <p className="text-sm">
                            Status:{" "}
                            <span
                              className={
                                playlist.is_active
                                  ? "text-green-600"
                                  : "text-gray-500"
                              }
                            >
                              {playlist.is_active ? "Active" : "Inactive"}
                            </span>
                          </p>
                        </div>
                        <CardActions>
                          <div
                            className={
                              playlist.childs?.length === 0
                                ? "tooltip tooltip-bottom"
                                : ""
                            }
                            data-tip={
                              playlist.childs?.length === 0
                                ? "Create some child playlists to enable sync"
                                : undefined
                            }
                          >
                            <Button
                              onClick={() => handleSync(playlist.id)}
                              variant="outline"
                              size="sm"
                              disabled={
                                playlist.childs?.length === 0 ||
                                syncMutation.isPending ||
                                deleteMutation.isPending
                              }
                              className="btn-primary"
                            >
                              {syncMutation.isPending &&
                              syncMutation.variables === playlist.id ? (
                                <span className="loading loading-spinner loading-sm"></span>
                              ) : (
                                <ArrowPathIcon className="h-4 w-4" />
                              )}
                            </Button>
                          </div>
                          <Button
                            onClick={() => handleDelete(playlist)}
                            variant="outline"
                            size="sm"
                            disabled={
                              deleteMutation.isPending || syncMutation.isPending
                            }
                            className="btn-error"
                          >
                            {deleteMutation.isPending ? (
                              "..."
                            ) : (
                              <TrashIcon className="h-4 w-4" />
                            )}
                          </Button>
                        </CardActions>
                      </div>

                      {/* Child Playlists Section */}
                      {isExpanded && (
                        <div className="mt-4 pt-4 border-t border-gray-200">
                          <div className="flex items-center justify-between mb-4">
                            <h4 className="text-lg font-medium">
                              Child Playlists ({(playlist.childs || []).length})
                            </h4>
                            <Button
                              onClick={() => handleCreateChild(playlist.id)}
                              variant="outline"
                              size="sm"
                              className="btn-primary"
                            >
                              <PlusIcon className="h-4 w-4" />
                              Add Child Playlist
                            </Button>
                          </div>
                          <ChildPlaylistList
                            childPlaylists={playlist.childs || []}
                            onEdit={handleChildEdit}
                            onDelete={handleChildDelete}
                            className="max-w-none"
                            isSyncing={
                              syncMutation.isPending &&
                              syncMutation.variables === playlist.id
                            }
                          />
                        </div>
                      )}
                    </CardBody>
                  </Card>
                </div>
              );
            })
          )}
        </div>
      )}
    </div>
  );
}