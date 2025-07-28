export interface BasePlaylist {
  id: string
  user_id: string
  name: string
  spotify_playlist_id: string
  is_active: boolean
  created: string
  updated: string
}

export interface CreateBasePlaylistRequest {
  name: string
  spotify_playlist_id?: string
}