package bench

import (
	"context"
	"math/rand"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/worker"
)

func (s *Scenario) FavoriteWorker(step *isucandar.BenchmarkStep, p int32) (*worker.Worker, error) {
	w, err := worker.NewWorker(func(ctx context.Context, _ int) {
		s.FavoriteScenario(ctx, step)
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

// 新着Playlistにfav/add爆撃をするシナリオ
func (s *Scenario) FavoriteScenario(ctx context.Context, step *isucandar.BenchmarkStep) error {
	report := timeReporter("favorite")
	defer report()

	// fav爆をするのはヘビーユーザー
	user, release := s.ChoiceUser(ctx, s.HeavyUsers)
	defer release()
	if user == nil {
		return nil
	}
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
	var ULIDS []string
	{
		// 最新プレイリスト一覧
		res, err := GetRecentPlaylistsAction(ctx, ag)
		v := ValidateResponse("最新プレイリスト一覧",
			step, res, err,
			WithStatusCode(200),
			WithSuccessResponse(func(r ResponseAPIGetRecentPlaylists) error {
				for _, p := range r.Playlists {
					if !p.IsFavorited && rand.Int()%4 == 0 { // 25%の確率で
						ULIDS = append(ULIDS, p.ULID)
					}
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
	if rand.Int31n(100) < s.RateGetPopularPlaylists() { // popularは確率で取る
		// 人気プレイリスト一覧
		res, err := GetPopularPlaylistsAction(ctx, ag)
		v := ValidateResponse("人気プレイリスト一覧",
			step, res, err,
			WithStatusCode(200),
			WithSuccessResponse(func(r ResponseAPIGetRecentPlaylists) error {
				for _, p := range r.Playlists {
					if !p.IsFavorited && rand.Int()%4 == 0 { // 25%の確率で
						ULIDS = append(ULIDS, p.ULID)
					}
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
	for _, id := range ULIDS {
		res, err := FavoritePlaylistAction(ctx, &Playlist{ULID: id, IsFavorited: true}, ag)
		v := ValidateResponse(
			"プレイリストをfavする",
			step, res, err,
			WithStatusCode(200, 404), // banされたときは404になるのでどっちでも許容
		)
		v.Add(step)
		if v.IsEmpty() {
			step.AddScore(ScoreFavoritePlaylist)
		} else {
			return v
		}
	}
	return nil
}
