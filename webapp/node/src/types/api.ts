// API essential types
export type Playlist = {
  ulid: string
  name: string
  user_display_name: string
  user_account: string
  song_count: number
  favorite_count: number
  is_favorited: boolean
  is_public: boolean
  created_at: Date
  updated_at: Date
}

export type PlaylistDetail = Playlist & {
  songs: Song[]
}

export type Song = {
  ulid: string
  title: string
  artist: string
  album: string
  track_number: number
  is_public: boolean
}

// API request types
export type SignupRequest = {
  user_account: string
  password: string
  display_name: string
}

export type LoginRequest = {
  user_account: string
  password: string
}

export type AddPlaylistRequest = {
  name: string
}

export type UpdatePlaylistRequest = {
  name?: string
  song_ulids?: string[]
  is_public?: boolean
}

export type FavoritePlaylistRequest = {
  is_favorited: boolean
}

export type AdminPlayerBanRequest = {
  user_account: string
  is_ban: boolean
}

// API response types
export type BasicResponse = {
  result: boolean
  status: number
  error?: string
}

export type GetRecentPlaylistsResponse = BasicResponse & {
  playlists: Playlist[]
}

export type GetPlaylistsResponse = BasicResponse & {
  created_playlists: Playlist[]
  favorited_playlists: Playlist[]
}

export type AddPlaylistResponse = BasicResponse & {
  playlist_ulid: string
}

export type SinglePlaylistResponse = BasicResponse & {
  playlist: PlaylistDetail
}

export type AdminPlayerBanResponse = BasicResponse & {
  user_account: string
  display_name: string
  is_ban: boolean
  created_at: Date
}
