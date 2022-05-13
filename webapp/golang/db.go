package main

import "time"

type UserRow struct {
	Account       string    `db:"account"`
	PasswordHash  string    `db:"password_hash"`
	DisplayName   string    `db:"display_name"`
	IsBan         bool      `db:"is_ban"`
	CreatedAt     time.Time `db:"created_at"`
	LastLoginedAt time.Time `db:"last_logined_at"`
}

type SongRow struct {
	ID          int    `db:"id"`
	ULID        string `db:"ulid"`
	Title       string `db:"title"`
	ArtistID    int    `db:"artist_id"`
	Album       string `db:"album"`
	TrackNumber int    `db:"track_number"`
	IsPublic    bool   `db:"is_public"`
}

type ArtistRow struct {
	ID   int    `db:"id"`
	ULID string `db:"ulid"`
	Name string `db:"name"`
}

type PlaylistRow struct {
	ID          int       `db:"id"`
	ULID        string    `db:"ulid"`
	Name        string    `db:"name"`
	UserAccount string    `db:"user_account"`
	IsPublic    bool      `db:"is_public"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type PlaylistSongRow struct {
	PlaylistID int `db:"playlist_id"`
	SortOrder  int `db:"sort_order"`
	SongID     int `db:"song_id"`
}

type PlaylistFavoriteRow struct {
	ID                  int       `db:"id"`
	PlaylistID          int       `db:"playlist_id"`
	FavoriteUserAccount string    `db:"favorite_user_account"`
	CreatedAt           time.Time `db:"created_at"`
}
