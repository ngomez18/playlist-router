import { useState, useEffect } from 'react'
import {
  Card,
  CardBody,
  CardTitle,
  Button,
  Input,
  Alert,
  Select,
  LoadingSpinner,
} from "../../components/ui";
import { apiClient } from "../../lib/api";
import type { CreateBasePlaylistRequest } from "../../types/playlist";
import type { SpotifyPlaylist } from "../../types/spotify";

interface CreateBasePlaylistFormProps {
  onSuccess?: () => void;
  onCancel?: () => void;
}

export function CreateBasePlaylistForm({
  onSuccess,
  onCancel,
}: CreateBasePlaylistFormProps) {
  const [name, setName] = useState("");
  const [selectedSpotifyPlaylist, setSelectedSpotifyPlaylist] = useState("");
  const [spotifyPlaylists, setSpotifyPlaylists] = useState<SpotifyPlaylist[]>(
    []
  );
  const [isLoading, setIsLoading] = useState(false);
  const [isLoadingPlaylists, setIsLoadingPlaylists] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadSpotifyPlaylists = async () => {
      try {
        setIsLoadingPlaylists(true);
        const playlists = await apiClient.getSpotifyPlaylists();
        setSpotifyPlaylists(playlists);
      } catch (err) {
        console.error("Failed to load Spotify playlists:", err);
        setError("Failed to load your Spotify playlists. Please try again.");
      } finally {
        setIsLoadingPlaylists(false);
      }
    };

    loadSpotifyPlaylists();
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!name.trim()) {
      setError("Playlist name is required");
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const request: CreateBasePlaylistRequest = {
        name: name.trim(),
        spotify_playlist_id: selectedSpotifyPlaylist || undefined,
      };

      await apiClient.createBasePlaylist(request);

      // Reset form
      setName("");
      setSelectedSpotifyPlaylist("");

      // Call success callback
      onSuccess?.();
    } catch (err) {
      console.error("Failed to create base playlist:", err);
      setError("Failed to create playlist. Please try again.");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Card>
      <CardBody>
        <CardTitle>Create Base Playlist</CardTitle>
        <p className="text-sm text-gray-600 mb-4">
          Create a new base playlist by either selecting an existing Spotify
          playlist or creating a new one.
        </p>

        {isLoadingPlaylists ? (
          <div className="flex items-center justify-center py-8">
            <LoadingSpinner />
            <span className="ml-2 text-sm text-gray-600">
              Loading your Spotify playlists...
            </span>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <Select
                label="Choose Spotify Playlist (Optional)"
                value={selectedSpotifyPlaylist}
                onChange={(e) => {
                  setSelectedSpotifyPlaylist(e.target.value);
                  // If a playlist is selected, auto-fill the name
                  if (e.target.value) {
                    const selectedPlaylist = spotifyPlaylists.find(
                      (p) => p.id === e.target.value
                    );
                    if (selectedPlaylist) {
                      setName(selectedPlaylist.name);
                    }
                  }
                }}
                disabled={isLoading}
              >
                <option value="">Create new playlist</option>
                {spotifyPlaylists.map((playlist) => (
                  <option key={playlist.id} value={playlist.id}>
                    {playlist.name} ({playlist.tracks} tracks)
                  </option>
                ))}
              </Select>
            </div>

            <div>
              <label
                htmlFor="playlist-name"
                className="block text-sm font-medium mb-2"
              >
                Base Playlist Name
              </label>
              <Input
                type="text"
                placeholder="Enter playlist name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
                disabled={isLoading}
              />
              {selectedSpotifyPlaylist ? (
                <p className="text-xs text-gray-500 mt-1">
                  This will be used as your base playlist name. The Spotify
                  playlist will remain unchanged.
                </p>
              ) : (
                <p className="text-xs text-gray-500 mt-1">
                  A new Spotify playlist with this name will be created for you.
                </p>
              )}
            </div>

            {error && <Alert type="error">{error}</Alert>}

            <div className="flex gap-2 pt-2">
              <Button type="submit" disabled={isLoading || !name.trim()}>
                {isLoading ? "Creating..." : "Create Base Playlist"}
              </Button>
              {onCancel && (
                <Button
                  type="button"
                  variant="secondary"
                  onClick={onCancel}
                  disabled={isLoading}
                >
                  Cancel
                </Button>
              )}
            </div>
          </form>
        )}
      </CardBody>
    </Card>
  );
}