import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { apiClient } from '../../lib/api'
import { Button, Input, Alert } from '../../components/ui'
import type { CreateChildPlaylistRequest, MetadataFilters, RangeFilter, SetFilter } from '../../types/playlist'

interface CreateChildPlaylistFormProps {
  basePlaylistId: string
  onSuccess: () => void
  onCancel: () => void
}

export function CreateChildPlaylistForm({ 
  basePlaylistId, 
  onSuccess, 
  onCancel 
}: CreateChildPlaylistFormProps) {
  const [formData, setFormData] = useState<CreateChildPlaylistRequest>({
    name: '',
    description: '',
    filter_rules: {},
  })

  const [expandedCategories, setExpandedCategories] = useState<Record<string, boolean>>({
    'track': true,
    'artist': true,
    'search': true
  })

  const createMutation = useMutation({
    mutationFn: (data: CreateChildPlaylistRequest) => 
      apiClient.createChildPlaylist(basePlaylistId, data),
    onSuccess: () => {
      onSuccess()
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!formData.name.trim()) return
    
    // Clean up filter rules - remove empty ranges and sets, but preserve boolean values
    const cleanedFilters: MetadataFilters = {}
    if (formData.filter_rules) {
      Object.entries(formData.filter_rules).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          if (typeof value === 'boolean') {
            // Boolean filters like 'explicit' should be preserved
            (cleanedFilters as Record<string, boolean>)[key] = value
          } else if (value && typeof value === 'object') {
            if ('min' in value || 'max' in value) {
              const range = value as RangeFilter
              if (range.min !== undefined || range.max !== undefined) {
                (cleanedFilters as Record<string, RangeFilter>)[key] = range
              }
            } else if ('include' in value || 'exclude' in value) {
              const set = value as SetFilter
              if ((set.include && set.include.length > 0) || (set.exclude && set.exclude.length > 0)) {
                (cleanedFilters as Record<string, SetFilter>)[key] = set
              }
            }
          }
        }
      })
    }
    
    // Require at least one filter
    if (Object.keys(cleanedFilters).length === 0) {
      alert('Please set at least one filter for your child playlist.')
      return
    }
    
    createMutation.mutate({
      ...formData,
      description: formData.description?.trim() || undefined,
      filter_rules: cleanedFilters,
    })
  }

  const handleInputChange = (field: keyof CreateChildPlaylistRequest) => 
    (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      setFormData(prev => ({
        ...prev,
        [field]: e.target.value
      }))
    }

  const handleFilterChange = (filterKey: keyof MetadataFilters, type: 'min' | 'max', value: string) => {
    if (value === '') {
      setFormData(prev => ({
        ...prev,
        filter_rules: {
          ...prev.filter_rules,
          [filterKey]: {
            ...(prev.filter_rules?.[filterKey] as RangeFilter || {}),
            [type]: undefined
          }
        }
      }))
      return
    }

    const numValue = parseFloat(value)
    
    // No conversion needed for metadata filters - all use their natural ranges
    const apiValue = numValue
    
    setFormData(prev => ({
      ...prev,
      filter_rules: {
        ...prev.filter_rules,
        [filterKey]: {
          ...(prev.filter_rules?.[filterKey] as RangeFilter || {}),
          [type]: apiValue
        }
      }
    }))
  }

  const getFilterValue = (filterKey: keyof MetadataFilters, type: 'min' | 'max'): string => {
    const filter = formData.filter_rules?.[filterKey] as RangeFilter
    const value = filter?.[type]
    if (value === undefined) return ''
    
    // No conversion needed for metadata filters
    return value.toString()
  }


  const toggleCategory = (category: string) => {
    setExpandedCategories(prev => ({
      ...prev,
      [category]: !prev[category]
    }))
  }

  const handleBooleanFilterChange = (filterKey: keyof MetadataFilters, value: string) => {
    const boolValue = value === '' ? undefined : value === 'true'
    
    setFormData(prev => ({
      ...prev,
      filter_rules: {
        ...prev.filter_rules,
        [filterKey]: boolValue
      }
    }))
  }

  const getBooleanFilterValue = (filterKey: keyof MetadataFilters): string => {
    const value = formData.filter_rules?.[filterKey]
    if (value === undefined) return ''
    return value ? 'true' : 'false'
  }

  return (
    <div className="card w-full max-w-2xl bg-base-200 shadow-xl border border-base-300">
      <div className="card-body">
        <h2 className="card-title">Create Child Playlist</h2>
        <p className="text-sm text-base-content/70 mb-4">
          Create a filtered playlist that will receive songs from your base playlist.
        </p>

        {createMutation.error && (
          <Alert type="error" className="mb-4">
            {createMutation.error.message || 'Failed to create child playlist'}
          </Alert>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="form-control">
            <span className="label-text mb-2 block">Name</span>
            <Input
              type="text"
              placeholder="Enter playlist name"
              value={formData.name}
              onChange={handleInputChange('name')}
              required
              className="input-bordered"
              disabled={createMutation.isPending}
            />
          </div>

          <div className="form-control">
            <span className="label-text mb-2 block">Description</span>
            <Input
              type="text"
              placeholder="Optional description for your playlist"
              value={formData.description || ''}
              onChange={handleInputChange('description')}
              className="input-bordered"
              disabled={createMutation.isPending}
            />
          </div>

          {/* Metadata Filters Section */}
          <div className="form-control">
            <span className="label-text mb-3 block">Metadata Filters</span>
            <div className="space-y-4">
              
              {/* Track Information Category */}
              <div className="collapse collapse-arrow bg-base-100 border border-base-300">
                <input 
                  type="checkbox" 
                  checked={expandedCategories.track}
                  onChange={() => toggleCategory('track')}
                />
                <div className="collapse-title text-sm font-semibold">
                  Track Information
                </div>
                <div className="collapse-content space-y-4">
                  
                  {/* Duration */}
                  <div className="form-control">
                    <span className="label-text text-xs mb-1 block">Duration (seconds)</span>
                    <div className="grid grid-cols-2 gap-2">
                      <input
                        type="number"
                        placeholder="Min (e.g., 180)"
                        min="0"
                        max="1800"
                        value={getFilterValue('duration_ms', 'min') ? Math.round(parseInt(getFilterValue('duration_ms', 'min')) / 1000).toString() : ''}
                        onChange={(e) => handleFilterChange('duration_ms', 'min', e.target.value ? (parseInt(e.target.value) * 1000).toString() : '')}
                        className="input input-bordered input-sm"
                        disabled={createMutation.isPending}
                      />
                      <input
                        type="number"
                        placeholder="Max (e.g., 300)"
                        min="0"
                        max="1800"
                        value={getFilterValue('duration_ms', 'max') ? Math.round(parseInt(getFilterValue('duration_ms', 'max')) / 1000).toString() : ''}
                        onChange={(e) => handleFilterChange('duration_ms', 'max', e.target.value ? (parseInt(e.target.value) * 1000).toString() : '')}
                        className="input input-bordered input-sm"
                        disabled={createMutation.isPending}
                      />
                    </div>
                  </div>

                  {/* Popularity */}
                  <div className="form-control">
                    <span className="label-text text-xs mb-1 block">Popularity (0 - 100)</span>
                    <div className="grid grid-cols-2 gap-2">
                      <input
                        type="number"
                        placeholder="Min (e.g., 50)"
                        min="0"
                        max="100"
                        value={getFilterValue('popularity', 'min')}
                        onChange={(e) => handleFilterChange('popularity', 'min', e.target.value)}
                        className="input input-bordered input-sm"
                        disabled={createMutation.isPending}
                      />
                      <input
                        type="number"
                        placeholder="Max (e.g., 100)"
                        min="0"
                        max="100"
                        value={getFilterValue('popularity', 'max')}
                        onChange={(e) => handleFilterChange('popularity', 'max', e.target.value)}
                        className="input input-bordered input-sm"
                        disabled={createMutation.isPending}
                      />
                    </div>
                  </div>

                  {/* Explicit Content */}
                  <div className="form-control">
                    <span className="label-text text-xs mb-1 block">Content Rating</span>
                    <select
                      value={getBooleanFilterValue('explicit')}
                      onChange={(e) => handleBooleanFilterChange('explicit', e.target.value)}
                      className="select select-bordered select-sm"
                      disabled={createMutation.isPending}
                    >
                      <option value="">Any (Clean + Explicit)</option>
                      <option value="false">Clean Only</option>
                      <option value="true">Explicit Only</option>
                    </select>
                  </div>
                </div>
              </div>

              {/* Artist & Album Category */}
              <div className="collapse collapse-arrow bg-base-100 border border-base-300">
                <input 
                  type="checkbox" 
                  checked={expandedCategories.artist}
                  onChange={() => toggleCategory('artist')}
                />
                <div className="collapse-title text-sm font-semibold">
                  Artist & Album Information
                </div>
                <div className="collapse-content space-y-4">
                  
                  {/* Genres */}
                  <div className="form-control opacity-50">
                    <span className="label-text text-xs mb-1 block">Genres (Coming Soon)</span>
                    <div className="grid grid-cols-1 gap-2">
                      <Input
                        type="text"
                        placeholder="Include genres (e.g., rock, pop, jazz) - comma separated"
                        value=""
                        onChange={() => {}}
                        className="input-bordered input-sm"
                        disabled={true}
                      />
                      <Input
                        type="text"
                        placeholder="Exclude genres - comma separated"
                        value=""
                        onChange={() => {}}
                        className="input-bordered input-sm"
                        disabled={true}
                      />
                    </div>
                  </div>

                  {/* Release Year */}
                  <div className="form-control">
                    <span className="label-text text-xs mb-1 block">Release Year</span>
                    <div className="grid grid-cols-2 gap-2">
                      <input
                        type="number"
                        placeholder="Min (e.g., 2000)"
                        min="1900"
                        max="2030"
                        value={getFilterValue('release_year', 'min')}
                        onChange={(e) => handleFilterChange('release_year', 'min', e.target.value)}
                        className="input input-bordered input-sm"
                        disabled={createMutation.isPending}
                      />
                      <input
                        type="number"
                        placeholder="Max (e.g., 2024)"
                        min="1900"
                        max="2030"
                        value={getFilterValue('release_year', 'max')}
                        onChange={(e) => handleFilterChange('release_year', 'max', e.target.value)}
                        className="input input-bordered input-sm"
                        disabled={createMutation.isPending}
                      />
                    </div>
                  </div>

                  {/* Artist Popularity */}
                  <div className="form-control">
                    <span className="label-text text-xs mb-1 block">Artist Popularity (0 - 100)</span>
                    <div className="grid grid-cols-2 gap-2">
                      <input
                        type="number"
                        placeholder="Min (e.g., 30)"
                        min="0"
                        max="100"
                        value={getFilterValue('artist_popularity', 'min')}
                        onChange={(e) => handleFilterChange('artist_popularity', 'min', e.target.value)}
                        className="input input-bordered input-sm"
                        disabled={createMutation.isPending}
                      />
                      <input
                        type="number"
                        placeholder="Max (e.g., 100)"
                        min="0"
                        max="100"
                        value={getFilterValue('artist_popularity', 'max')}
                        onChange={(e) => handleFilterChange('artist_popularity', 'max', e.target.value)}
                        className="input input-bordered input-sm"
                        disabled={createMutation.isPending}
                      />
                    </div>
                  </div>
                </div>
              </div>

              {/* Search-based Filters Category */}
              <div className="collapse collapse-arrow bg-base-100 border border-base-300">
                <input 
                  type="checkbox" 
                  checked={expandedCategories.search}
                  onChange={() => toggleCategory('search')}
                />
                <div className="collapse-title text-sm font-semibold">
                  Search-based Filters
                </div>
                <div className="collapse-content space-y-4">
                  
                  {/* Track Keywords */}
                  <div className="form-control opacity-50">
                    <span className="label-text text-xs mb-1 block">Track Name Keywords (Coming Soon)</span>
                    <div className="grid grid-cols-1 gap-2">
                      <Input
                        type="text"
                        placeholder="Include keywords (e.g., love, dance, party) - comma separated"
                        value=""
                        onChange={() => {}}
                        className="input-bordered input-sm"
                        disabled={true}
                      />
                      <Input
                        type="text"
                        placeholder="Exclude keywords - comma separated"
                        value=""
                        onChange={() => {}}
                        className="input-bordered input-sm"
                        disabled={true}
                      />
                    </div>
                  </div>

                  {/* Artist Keywords */}
                  <div className="form-control opacity-50">
                    <span className="label-text text-xs mb-1 block">Artist Name Keywords (Coming Soon)</span>
                    <div className="grid grid-cols-1 gap-2">
                      <Input
                        type="text"
                        placeholder="Include keywords (e.g., band, DJ, feat) - comma separated"
                        value=""
                        onChange={() => {}}
                        className="input-bordered input-sm"
                        disabled={true}
                      />
                      <Input
                        type="text"
                        placeholder="Exclude keywords - comma separated"
                        value=""
                        onChange={() => {}}
                        className="input-bordered input-sm"
                        disabled={true}
                      />
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div className="alert alert-warning">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" className="stroke-current shrink-0 w-6 h-6">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z"></path>
            </svg>
            <div className="text-sm">
              <div className="font-medium">Required:</div>
              <div>You must set at least one filter above. Songs from your base playlist will automatically be sorted into this child playlist based on the audio characteristics you define.</div>
            </div>
          </div>

          <div className="card-actions justify-end gap-2">
            <Button
              type="button"
              variant="ghost"
              onClick={onCancel}
              disabled={createMutation.isPending}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={!formData.name.trim() || createMutation.isPending}
              className="btn-primary"
            >
              {createMutation.isPending ? (
                <>
                  <span className="loading loading-spinner loading-sm"></span>
                  Creating...
                </>
              ) : (
                'Create Playlist'
              )}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}