import { useState } from 'react'
import {
  Card,
  CardBody,
  CardTitle,
  Button,
  Input,
  Alert,
} from "../../components/ui";
import { apiClient } from "../../lib/api";
import type { CreateBasePlaylistRequest } from "../../types/playlist";

interface CreateBasePlaylistFormProps {
  onSuccess?: () => void;
  onCancel?: () => void;
}

export function CreateBasePlaylistForm({
  onSuccess,
  onCancel,
}: CreateBasePlaylistFormProps) {
  const [name, setName] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

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
      };

      await apiClient.createBasePlaylist(request);

      // Reset form
      setName("");

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
          Create a new base playlist. A corresponding Spotify playlist will be
          created automatically.
        </p>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label
              htmlFor="playlist-name"
              className="block text-sm font-medium mb-2"
            >
              Playlist Name
            </label>
            <Input
              type="text"
              placeholder="Enter playlist name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              disabled={isLoading}
            />
          </div>

          {error && <Alert type="error">{error}</Alert>}

          <div className="flex gap-2 pt-2">
            <Button type="submit" disabled={isLoading || !name.trim()}>
              {isLoading ? "Creating..." : "Create Playlist"}
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
      </CardBody>
    </Card>
  );
}