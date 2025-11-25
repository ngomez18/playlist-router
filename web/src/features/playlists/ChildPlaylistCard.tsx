import type { ChildPlaylist, RangeFilter, SetFilter } from '../../types/playlist'
import { Card, CardBody, CardTitle, CardActions } from '../../components/ui/Card'
import { Button } from '../../components/ui/Button'
import { LoadingSpinner } from "../../components/ui";
import { PencilIcon, TrashIcon } from "@heroicons/react/24/outline";

interface ChildPlaylistCardProps {
  childPlaylist: ChildPlaylist;
  onEdit?: (playlist: ChildPlaylist) => void;
  onDelete?: (id: string) => void;
  isLoading?: boolean;
}

export function ChildPlaylistCard({
  childPlaylist,
  onEdit,
  onDelete,
  isLoading = false,
}: ChildPlaylistCardProps) {
  const createdDate = new Date(childPlaylist.created).toLocaleDateString();
  const hasFilters =
    childPlaylist.filter_rules &&
    Object.keys(childPlaylist.filter_rules).length > 0;

  return (
    <Card
      className={`w-full bg-base-100 border-2 border-base-300 shadow-lg relative ${
        isLoading ? "opacity-60" : ""
      }`}
    >
      {isLoading && (
        <div className="absolute inset-0 bg-base-100/80 backdrop-blur-sm flex items-center justify-center z-10 rounded-lg">
          <div className="text-center">
            <LoadingSpinner size="md" />
            <p className="text-sm text-base-content/70 mt-2">Syncing...</p>
          </div>
        </div>
      )}
      <CardBody>
        <CardTitle className="flex items-center justify-between">
          <span className="truncate">{childPlaylist.name}</span>
          <div className="flex items-center gap-2">
            {/* Status Badge */}
            <span
              className={`badge ${
                childPlaylist.is_active ? "badge-success" : "badge-neutral"
              }`}
            >
              {childPlaylist.is_active ? "Active" : "Inactive"}
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
          <p className="truncate">
            Spotify ID: {childPlaylist.spotify_playlist_id}
          </p>
        </div>

        {/* Filter Rules Summary */}
        {hasFilters && (
          <div className="mt-3">
            <div className="collapse collapse-arrow border-2 border-primary/30 bg-base-200 rounded-lg shadow-md">
              <input type="checkbox" className="peer" />
              <div className="collapse-title text-sm font-medium">
                Audio Filters (
                {Object.keys(childPlaylist.filter_rules || {}).length})
              </div>
              <div className="collapse-content">
                <div className="space-y-2 text-xs">
                  {Object.entries(childPlaylist.filter_rules || {}).map(
                    ([key, value]) => (
                      <div key={key} className="flex justify-between">
                        <span className="capitalize">
                          {key.replace("_", " ")}:
                        </span>
                        <span className="text-base-content/70">
                          {(() => {
                            // Handle boolean values (like explicit filter)
                            if (typeof value === "boolean") {
                              if (key === "explicit") {
                                return value ? "Explicit Only" : "Clean Only";
                              }
                              return value ? "Yes" : "No";
                            }

                            // Handle object values (ranges and sets)
                            if (typeof value === "object" && value !== null) {
                              if ("min" in value || "max" in value) {
                                const range = value as RangeFilter;
                                let minVal = range.min || -1;
                                let maxVal = range.max || -1;

                                // Convert duration from ms to seconds for display
                                if (key === "duration_ms") {
                                  minVal =
                                    minVal > 0
                                      ? Math.round(minVal / 1000)
                                      : minVal;
                                  maxVal =
                                    maxVal > 0
                                      ? Math.round(maxVal / 1000)
                                      : maxVal;
                                }

                                if (minVal < 0) {
                                  return `> ${maxVal}`;
                                }

                                if (maxVal < 0) {
                                  return `< ${minVal}`;
                                }

                                return `${minVal} - ${maxVal}`;
                              }

                              // Handle set filters (include/exclude)
                              if ("include" in value || "exclude" in value) {
                                const set = value as SetFilter;
                                const parts = [];
                                if (set.include && set.include.length > 0) {
                                  parts.push(`+${set.include.join(", ")}`);
                                }
                                if (set.exclude && set.exclude.length > 0) {
                                  parts.push(`-${set.exclude.join(", ")}`);
                                }
                                return parts.join("; ");
                              }

                              return "Custom";
                            }

                            // Fallback for other types
                            return String(value);
                          })()}
                        </span>
                      </div>
                    )
                  )}
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
                className="btn-sm"
              >
                <PencilIcon className="h-4 w-4" />
              </Button>
            )}
            {onDelete && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => onDelete(childPlaylist.id)}
                className="btn-sm btn-error"
              >
                <TrashIcon className="h-4 w-4" />
              </Button>
            )}
          </CardActions>
        )}
      </CardBody>
    </Card>
  );
}