package bench

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/worker"
)

// ログインしないユーザーのシナリオ
func (s *Scenario) AnonWorker(step *isucandar.BenchmarkStep, p int32) (*worker.Worker, error) {
	w, err := worker.NewWorker(func(ctx context.Context, _ int) {
		s.AnonScenario(ctx, step)
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

// 匿名User
func (s *Scenario) AnonScenario(ctx context.Context, step *isucandar.BenchmarkStep) error {
	report := timeReporter("anonymous")
	defer report()

	playlistULIDs := []string{}
	ag, _ := s.Option.NewAgent(false)
	{
		res, err := GetRecentPlaylistsAction(ctx, ag)
		v := ValidateResponse("最新プレイリスト一覧(非ログイン)",
			step, res, err,
			WithStatusCode(200),
			WithSuccessResponse(func(r ResponseAPIGetRecentPlaylists) error {
				var prevCreatedAt time.Time
				if len(r.Playlists) != 100 {
					return fmt.Errorf("最新プレイリストが100件ではありません %d", len(r.Playlists))
				}
				// ベンチマーカーが把握している最新投稿時刻の10秒前
				last := s.LastPublicPlaylistCreatedAt().Add(-10 * time.Second)
				// 現在の最新
				top := r.Playlists[0].CreatedAt
				if !last.IsZero() && last.After(top) {
					return fmt.Errorf("最新プレイリスト一覧が古すぎます %s", top)
				}
				for _, p := range r.Playlists {
					// 秒で切りそろえてから比較(サブ秒での逆転は許すため)
					createdAt := p.CreatedAt.Truncate(time.Second)
					if prevCreatedAt.IsZero() {
						prevCreatedAt = createdAt
					} else if createdAt.After(prevCreatedAt) {
						return fmt.Errorf("最新プレイリスト一覧の作成日時が降順になっていません %s < %s", p.CreatedAt, prevCreatedAt)
					}
					prevCreatedAt = createdAt
					if p.IsPublic == false {
						return fmt.Errorf("最新プレイリスト一覧に非公開プレイリストが含まれています")
					}
					if p.IsFavorited {
						return fmt.Errorf("非ログイン状態なのに最新プレイリスト一覧にfavが含まれています")
					}
					if rand.Intn(100) <= 10 {
						playlistULIDs = append(playlistULIDs, p.ULID)
					}
				}
				return nil
			}),
		)
		if v.IsEmpty() {
			step.AddScore(ScoreGetRecentPlaylists)
		} else {
			return v
		}
	}
	if rand.Int31n(100) <= s.RateGetPopularPlaylists() { // pupularは確率で取る
		res, err := GetPopularPlaylistsAction(ctx, ag)
		v := ValidateResponse("人気プレイリスト一覧(非ログイン)",
			step, res, err,
			WithStatusCode(200),
			WithSuccessResponse(func(r ResponseAPIGetPopularPlaylists) error {
				if len(r.Playlists) > 100 {
					return fmt.Errorf("人気プレイリスト一覧が100件を超えています %d", len(r.Playlists))
				}
				if len(r.Playlists) == 0 {
					return fmt.Errorf("人気プレイリスト一覧が空です")
				}
				var prevFavs int
				for _, p := range r.Playlists {
					if prevFavs == 0 {
						prevFavs = p.FavoriteCount
					}
					if p.FavoriteCount == 0 {
						return fmt.Errorf("人気プレイリストのfavが0です")
					}
					if prevFavs < p.FavoriteCount {
						return fmt.Errorf("人気プレイリスト一覧のfav数が降順になっていません %d < %d", prevFavs, p.FavoriteCount)
					}
					prevFavs = p.FavoriteCount
					if p.IsPublic == false {
						return fmt.Errorf("人気プレイリスト一覧に非公開プレイリストが含まれています")
					}
					if rand.Intn(100) <= 10 {
						playlistULIDs = append(playlistULIDs, p.ULID)
					}
				}
				return nil
			}),
		)
		if v.IsEmpty() {
			step.AddScore(ScoreGetPopularPlaylists)
		} else {
			return v
		}
	}

	for _, playlistULID := range playlistULIDs {
		// プレイリスト詳細
		res, err := GetPlaylistAction(ctx, playlistULID, ag)
		v := ValidateResponse("プレイリスト詳細取得(非ログイン)",
			step, res, err,
			WithStatusCode(200),
			WithSuccessResponse(func(r ResponseAPIGetPlaylist) error {
				if r.Playlist.ULID != playlistULID {
					return fmt.Errorf("プレイリストのULIDが一致しません %s != %s", r.Playlist.ULID, playlistULID)
				}
				if r.Playlist.IsFavorited {
					return fmt.Errorf("非ログイン状態なのにプレイリストがfavされています")
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

	return nil
}
