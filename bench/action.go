package bench

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/isucon/isucandar/agent"
)

var globalPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

func GetInitializeAction(ctx context.Context, ag *agent.Agent) (*http.Response, error) {
	// リクエストを生成
	b, reset, err := newRequestBody(struct{}{})
	if err != nil {
		return nil, err
	}
	defer reset()
	req, err := ag.POST("/initialize", b)
	if err != nil {
		return nil, err
	}

	// リクエストを実行
	return ag.Do(ctx, req)
}

func GetRootAction(ctx context.Context, ag *agent.Agent) (*http.Response, error) {
	// リクエストを生成
	req, err := ag.GET("/")
	if err != nil {
		return nil, err
	}

	// リクエストを実行
	return ag.Do(ctx, req)
}

func newRequestBody(obj any) (*bytes.Buffer, func(), error) {
	b := globalPool.Get().(*bytes.Buffer)
	reset := func() {
		b.Reset()
		globalPool.Put(b)
	}
	if err := json.NewEncoder(b).Encode(obj); err != nil {
		reset()
		return nil, nil, err
	}
	return b, reset, nil
}

func SignupAction(ctx context.Context, ag *agent.Agent) (*User, *http.Response, error) {
	// リクエストを生成
	account, password, name := GenerateUserAccount(), RandomString(32), DisplayName()
	b, reset, err := newRequestBody(struct {
		UserAccount string `json:"user_account"`
		Password    string `json:"password"`
		DisplayName string `json:"display_name"`
	}{
		UserAccount: account,
		Password:    password,
		DisplayName: name,
	})
	if err != nil {
		return nil, nil, err
	}
	defer reset()

	req, err := ag.POST("/api/signup", b)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	// リクエストを実行
	if res, err := ag.Do(ctx, req); err != nil {
		return nil, res, err
	} else {
		return &User{
			Account:     account,
			Password:    password,
			DisplayName: name,
		}, res, nil
	}
}

func LoginAction(ctx context.Context, user *User, ag *agent.Agent) (*http.Response, error) {
	report := timeReporter("login action")
	defer report()
	// リクエストを生成
	b, reset, err := newRequestBody(struct {
		UserAccount string `json:"user_account"`
		Password    string `json:"password"`
	}{
		UserAccount: user.Account,
		Password:    user.Password,
	})
	if err != nil {
		return nil, err
	}
	defer reset()

	req, err := ag.POST("/api/login", b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	// リクエストを実行
	return ag.Do(ctx, req)
}

func LogoutAction(ctx context.Context, user *User, ag *agent.Agent) (*http.Response, error) {
	// リクエストを生成
	b, reset, err := newRequestBody(struct{}{})
	if err != nil {
		return nil, err
	}
	defer reset()

	req, err := ag.POST("/api/logout", b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	// リクエストを実行
	return ag.Do(ctx, req)
}

func GetPlaylistsAction(ctx context.Context, ag *agent.Agent) (*http.Response, error) {
	req, err := ag.GET("/api/playlists")
	if err != nil {
		return nil, err
	}
	// リクエストを実行
	return ag.Do(ctx, req)
}

func GetRecentPlaylistsAction(ctx context.Context, ag *agent.Agent) (*http.Response, error) {
	req, err := ag.GET("/api/recent_playlists")
	if err != nil {
		return nil, err
	}
	// リクエストを実行
	return ag.Do(ctx, req)
}

func GetPopularPlaylistsAction(ctx context.Context, ag *agent.Agent) (*http.Response, error) {
	req, err := ag.GET("/api/popular_playlists")
	if err != nil {
		return nil, err
	}
	// リクエストを実行
	return ag.Do(ctx, req)
}

func GetPlaylistAction(ctx context.Context, id string, ag *agent.Agent) (*http.Response, error) {
	req, err := ag.GET("/api/playlist/" + id)
	if err != nil {
		return nil, err
	}
	// リクエストを実行
	return ag.Do(ctx, req)
}

func AddPlayistAction(ctx context.Context, name string, ag *agent.Agent) (*http.Response, error) {
	// リクエストを生成
	b, reset, err := newRequestBody(struct {
		Name string `json:"name"`
	}{
		Name: name,
	})
	if err != nil {
		return nil, err
	}
	defer reset()

	req, err := ag.POST("/api/playlist/add", b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	// リクエストを実行
	return ag.Do(ctx, req)
}

func UpdatePlayistAction(ctx context.Context, p *Playlist, ag *agent.Agent) (*http.Response, error) {
	// リクエストを生成
	b, reset, err := newRequestBody(struct {
		Name      string   `json:"name"`
		SongULIDs []string `json:"song_ulids"`
		IsPublic  bool     `json:"is_public"`
	}{
		Name:      p.Name,
		SongULIDs: p.Songs.ULIDs(),
		IsPublic:  p.IsPublic,
	})
	if err != nil {
		return nil, err
	}
	defer reset()

	req, err := ag.POST("/api/playlist/"+p.ULID+"/update", b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	// リクエストを実行
	return ag.Do(ctx, req)
}

func FavoritePlaylistAction(ctx context.Context, p *Playlist, ag *agent.Agent) (*http.Response, error) {
	// リクエストを生成
	b, reset, err := newRequestBody(struct {
		IsFavorited bool `json:"is_favorited"`
	}{
		IsFavorited: p.IsFavorited,
	})
	if err != nil {
		return nil, err
	}
	defer reset()

	req, err := ag.POST("/api/playlist/"+p.ULID+"/favorite", b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	// リクエストを実行
	return ag.Do(ctx, req)
}

func DeletePlaylistAction(ctx context.Context, p *Playlist, ag *agent.Agent) (*http.Response, error) {
	req, err := ag.POST("/api/playlist/"+p.ULID+"/delete", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	// リクエストを実行
	return ag.Do(ctx, req)
}

func AdminBanAction(ctx context.Context, user *User, isBan bool, ag *agent.Agent) (*http.Response, error) {
	b, reset, err := newRequestBody(struct {
		IsBan       bool   `json:"is_ban"`
		UserAccount string `json:"user_account"`
	}{
		IsBan:       isBan,
		UserAccount: user.Account,
	})
	if err != nil {
		return nil, err
	}
	defer reset()

	req, err := ag.POST("/api/admin/user/ban", b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	// リクエストを実行
	return ag.Do(ctx, req)
}
