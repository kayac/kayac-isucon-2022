package main

import "time"

// API essential types

type Playlist struct {
	ULID            string    `json:"ulid"`
	Name            string    `json:"name"`
	UserDisplayName string    `json:"user_display_name"`
	UserAccount     string    `json:"user_account"`
	SongCount       int       `json:"song_count"`
	FavoriteCount   int       `json:"favorite_count"`
	IsFavorited     bool      `json:"is_favorited"`
	IsPublic        bool      `json:"is_public"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type PlaylistDetail struct {
	*Playlist
	Songs []Song `json:"songs"`
}

type Song struct {
	ULID        string `json:"ulid"`
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	Album       string `json:"album"`
	TrackNumber int    `json:"track_number"`
	IsPublic    bool   `json:"is_public"`
}

// API request types

type SignupRequest struct {
	UserAccount string `json:"user_account"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type LoginRequest struct {
	UserAccount string `json:"user_account"`
	Password    string `json:"password"`
}

type AddPlaylistRequest struct {
	Name string `json:"name"`
}

type UpdatePlaylistRequest struct {
	Name      *string  `json:"name"`
	SongULIDs []string `json:"song_ulids,omitempty"`
	IsPublic  bool     `json:"is_public"`
}

type FavoritePlaylistRequest struct {
	IsFavorited bool `json:"is_favorited"`
}

type AdminPlayerBanRequest struct {
	UserAccount string `json:"user_account"`
	IsBan       bool   `json:"is_ban"`
}

// API response types

type BasicResponse struct {
	Result bool    `json:"result"`
	Status int     `json:"status"`
	Error  *string `json:"error,omitempty"`
}

type GetRecentPlaylistsResponse struct {
	BasicResponse
	Playlists []Playlist `json:"playlists"`
}

type GetPlaylistsResponse struct {
	BasicResponse
	CreatedPlaylists   []Playlist `json:"created_playlists"`
	FavoritedPlaylists []Playlist `json:"favorited_playlists"`
}

type AddPlaylistResponse struct {
	BasicResponse
	PlaylistULID string `json:"playlist_ulid"`
}

type SinglePlaylistResponse struct {
	BasicResponse
	Playlist PlaylistDetail `json:"playlist"`
}

type AdminPlayerBanResponse struct {
	BasicResponse
	UserAccount string    `json:"user_account"`
	DisplayName string    `json:"display_name"`
	IsBan       bool      `json:"is_ban"`
	CreatedAt   time.Time `json:"created_at"`
}
