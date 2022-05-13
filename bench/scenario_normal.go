package bench

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/worker"
	"github.com/samber/lo"
)

func (s *Scenario) NormalWorker(step *isucandar.BenchmarkStep, p int32) (*worker.Worker, error) {
	w, err := worker.NewWorker(func(ctx context.Context, _ int) {
		s.NormalScenario(ctx, step)
	},
		// 無限回繰り返す
		worker.WithInfinityLoop(),
		worker.WithUnlimitedParallelism(),
	)
	if err != nil {
		return nil, err
	}
	w.SetParallelism(p)
	return w, nil
}

// 普通のUser
func (s *Scenario) NormalScenario(ctx context.Context, step *isucandar.BenchmarkStep) error {
	report := timeReporter("normal")
	defer report()

	user, release := s.ChoiceUser(ctx, s.NormalUsers)
	defer release()
	ag, _ := user.GetAgent(s.Option)
	{
		// ログイン
		res, err := LoginAction(ctx, user, ag)
		v := ValidateResponse("ログイン",
			step, res, err,
			WithStatusCode(200),
			WithSuccessResponse[ResponseAPIBase](),
		)
		if v.IsEmpty() {
			step.AddScore(ScoreLogin)
		} else {
			return v
		}
	}
	if rand.Int31n(100) < s.RateGetPopularPlaylists() { // popularは確率で取る
		res, err := GetPopularPlaylistsAction(ctx, ag)
		v := ValidateResponse("人気プレイリスト一覧",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIGetPopularPlaylists) error {
				if len(r.Playlists) > 100 {
					return fmt.Errorf("人気プレイリスト一覧が100件を超えています %d", len(r.Playlists))
				}
				if len(r.Playlists) == 0 {
					return fmt.Errorf("人気プレイリスト一覧が空です")
				}
				return nil
			}),
		)
		if v.IsEmpty() {
			step.AddScore(ScoreGetPopularPlaylistsLogin)
		} else {
			return v
		}
	}
	var playlistULIDs []string
	{
		// プレイリスト一覧
		res, err := GetPlaylistsAction(ctx, ag)
		v := ValidateResponse("自分のプレイリスト一覧",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIGetPlaylists) error {
				for _, p := range r.CreatedPlaylists {
					playlistULIDs = append(playlistULIDs, p.ULID)
				}
				for _, p := range r.FavoritedPlaylists {
					playlistULIDs = append(playlistULIDs, p.ULID)
				}
				return nil
			}),
		)

		if v.IsEmpty() {
			step.AddScore(ScoreGetPlaylists)
		} else {
			return v
		}
	}
	var favULIDS []string
	// 自分のプレイリスト詳細を1/4ぐらい取る
	playlistULIDs = lo.Samples(playlistULIDs, len(playlistULIDs)/4+1)
	for _, playlistULID := range playlistULIDs {
		res, err := GetPlaylistAction(ctx, playlistULID, ag)
		v := ValidateResponse("プレイリスト詳細(ログイン中)",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIGetPlaylist) error {
				p := r.Playlist
				if rand.Int()%10 == 0 { // 10%の確率でfavを付け外し
					favULIDS = append(favULIDS, p.ULID)
				}
				return nil
			}),
		)
		if v.IsEmpty() {
			step.AddScore(ScoreGetPlaylist)
		} else {
			return v
		}
	}
	for _, id := range favULIDS {
		isFav := rand.Int()%2 == 0
		res, err := FavoritePlaylistAction(ctx, &Playlist{ULID: id, IsFavorited: isFav}, ag)
		v := ValidateResponse("プレイリストにfavする",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
		)
		if v.IsEmpty() {
			step.AddScore(ScoreFavoritePlaylist)
		} else {
			return v
		}
	}
	var playlistULID string
	{
		// 新規プレイリスト作成
		playlistName := DisplayName()
		res, err := AddPlayistAction(ctx, playlistName, ag)
		v := ValidateResponse("新規プレイリスト作成",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIAddPlaylist) error {
				if r.PlaylistULID == "" {
					return fmt.Errorf("作成されたプレイリストのULIDが空です")
				}
				playlistULID = r.PlaylistULID
				return nil
			}),
		)
		if v.IsEmpty() {
			step.AddScore(ScoreAddPlaylist)
		} else {
			return v
		}
	}
	{
		// プレイリスト一覧
		res, err := GetPlaylistsAction(ctx, ag)
		v := ValidateResponse("自分のプレイリスト一覧",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIGetPlaylists) error {
				if len(r.CreatedPlaylists) == 0 {
					return fmt.Errorf("作成済みプレイリストが0件です")
				}
				var found bool
				for _, p := range r.CreatedPlaylists {
					if p.ULID == playlistULID {
						if p.SongCount != 0 {
							return fmt.Errorf("作成済みプレイリストの曲数が想定外です")
						}
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("作成したプレイリストが見つかりません")
				}
				return nil
			}),
		)
		if v.IsEmpty() {
			step.AddScore(ScoreGetPlaylists)
		} else {
			return v
		}
	}
	var playlist Playlist
	{
		// プレイリスト詳細 作成直後なので空
		res, err := GetPlaylistAction(ctx, playlistULID, ag)
		v := ValidateResponse("プレイリスト詳細(ログイン中)",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIGetPlaylist) error {
				if r.Playlist.ULID != playlistULID {
					return fmt.Errorf("ULIDが一致しません")
				}
				playlist = r.Playlist
				return nil
			}),
		)
		if v.IsEmpty() {
			step.AddScore(ScoreGetPlaylist)
		} else {
			return v
		}
	}
	addSongs := lo.Samples(s.Songs, rand.Intn(80))
	{
		// プレイリスト更新、n曲追加 公開にする
		playlist.Songs = append(playlist.Songs, addSongs...)
		playlist.IsPublic = true
		res, err := UpdatePlayistAction(ctx, &playlist, ag)
		v := ValidateResponse("プレイリスト更新 曲追加 公開にする",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIGetPlaylist) error {
				if r.Playlist.ULID != playlistULID {
					return fmt.Errorf("プレイリストのULIDが一致しません")
				}
				if r.Playlist.SongCount != len(addSongs) {
					return fmt.Errorf("%d曲追加したはずですが%d曲になっています", len(addSongs), r.Playlist.SongCount)
				}
				expectd := playlist.Songs.ULIDs()
				for i, ret := range r.Playlist.Songs.ULIDs() {
					if ret != expectd[i] {
						return fmt.Errorf("%d番目の曲 %s が、追加した曲 %s と一致しません", i+1, ret, expectd[i])
					}
				}
				// 公開にしたので最新時刻を設定
				s.SetLastPublicPlaylistCreatedAt(r.Playlist.CreatedAt)
				return nil
			}),
		)
		if v.IsEmpty() {
			step.AddScore(ScoreUpdatePlaylist)
		} else {
			return v
		}
	}
	{
		// プレイリスト一覧を取り直す。曲が入っているはず
		res, err := GetPlaylistsAction(ctx, ag)
		v := ValidateResponse("プレイリスト一覧",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIGetPlaylists) error {
				if len(r.CreatedPlaylists) == 0 {
					return fmt.Errorf("作成済みプレイリストが0件です")
				}
				var found bool
				for _, p := range r.CreatedPlaylists {
					if p.ULID == playlistULID {
						if p.SongCount != len(addSongs) {
							return fmt.Errorf("作成済みプレイリストの曲数が想定外です")
						}
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("作成されたプレイリスト %s が作成済みプレイリストにありません", playlistULID)
				}
				return nil
			}),
		)
		if v.IsEmpty() {
			step.AddScore(ScoreGetPlaylists)
		} else {
			return v
		}
	}
	{
		// 最新プレイリスト一覧を見に行く(投稿したばかりなのであるはず)
		res, err := GetRecentPlaylistsAction(ctx, ag)
		v := ValidateResponse("最新プレイリスト一覧",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIGetRecentPlaylists) error {
				var found bool
				var oldestCreatedAt time.Time
				for _, p := range r.Playlists {
					if p.ULID == playlistULID {
						if p.SongCount != len(addSongs) {
							return fmt.Errorf("プレイリストの %s 曲数(%d)が違います", playlistULID, p.SongCount)
						}
						found = true
						break
					}
					oldestCreatedAt = p.CreatedAt
				}
				// top100に見つからない
				// 既に100件以上投稿されて押し出されている場合は見つからないのでoldestと比較も必要
				if !found && oldestCreatedAt.After(playlist.CreatedAt) {
					return fmt.Errorf("最新プレイリストにプレイリスト %s が見つかりません", playlistULID)
				}
				return nil
			}),
		)
		if v.IsEmpty() {
			step.AddScore(ScoreGetRecentPlaylistsLogin)
		} else {
			return v
		}
	}

	return nil
}
