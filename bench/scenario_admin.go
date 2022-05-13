package bench

import (
	"context"
	"fmt"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/worker"
	"github.com/samber/lo"
)

func (s *Scenario) AdminWorker(step *isucandar.BenchmarkStep, _ int32) (*worker.Worker, error) {
	w, err := worker.NewWorker(func(ctx context.Context, _ int) {
		timer := time.NewTimer(3 * time.Second)
		select {
		case <-ctx.Done():
		case <-timer.C:
		}
		s.AdminScenario(ctx, step)
	},
		// 無限回繰り返す
		worker.WithInfinityLoop(),
		worker.WithMaxParallelism(1),
	)
	if err != nil {
		return nil, err
	}
	w.SetParallelism(1)
	return w, nil
}

// AdminUserのシナリオ
func (s *Scenario) AdminScenario(ctx context.Context, step *isucandar.BenchmarkStep) error {
	report := timeReporter("admin")
	defer report()

	adminAg, _ := s.AdminUser.GetAgent(s.Option)
	{
		// admin user ログイン
		admin := s.AdminUser
		res, err := LoginAction(ctx, admin, adminAg)
		v := ValidateResponse(
			"adminログイン",
			step, res, err,
			WithStatusCode(200),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	var banPlayList Playlist
	{
		// RecentPlaylistを取得
		res, err := GetRecentPlaylistsAction(ctx, adminAg)
		v := ValidateResponse("最新プレイリスト一覧を取得",
			step, res, err,
			WithStatusCode(200),
			WithSuccessResponse(func(r ResponseAPIGetRecentPlaylists) error {
				if len(r.Playlists) != 100 {
					return fmt.Errorf("最新プレイリストには100曲あるはずですが、%d曲しかありません", len(r.Playlists))
				}
				banPlayList = lo.Sample(r.Playlists)
				return nil
			}),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	{
		// admin ban
		banUser := &User{
			Account: banPlayList.UserAccount,
		}
		res, err := AdminBanAction(ctx, banUser, true, adminAg)
		v := ValidateResponse(
			"admin ban",
			step, res, err,
			WithStatusCode(200),
			WithSuccessResponse(func(r ResponseAdminBan) error {
				if r.UserAccount != banUser.Account {
					return fmt.Errorf("banしたユーザーが正しくありません")
				}
				if !r.IsBan {
					return fmt.Errorf("banが正常に実行されていません")
				}
				return nil
			}),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	// ban操作のあと3秒は猶予がある
	time.Sleep(3 * time.Second)
	{
		// プレイリスト詳細が404になる
		res, err := GetPlaylistAction(ctx, banPlayList.ULID, adminAg)
		v := ValidateResponse(
			"banされたプレイリスト詳細は404",
			step, res, err,
			WithStatusCode(404),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	{
		// admin unban
		banUser := &User{
			Account: banPlayList.UserAccount,
		}
		res, err := AdminBanAction(ctx, banUser, false, adminAg)
		v := ValidateResponse(
			"ban解除",
			step, res, err,
			WithStatusCode(200),
			WithSuccessResponse(func(r ResponseAdminBan) error {
				if r.UserAccount != banUser.Account {
					return fmt.Errorf("ban解除したユーザーが正しくありません")
				}
				if r.IsBan {
					return fmt.Errorf("ban解除が正常に実行されていません")
				}
				return nil
			}),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	// ban操作のあと3秒は猶予がある
	time.Sleep(3 * time.Second)
	{
		// プレイリスト詳細が200になる
		res, err := GetPlaylistAction(ctx, banPlayList.ULID, adminAg)
		v := ValidateResponse(
			"プレイリスト詳細",
			step, res, err,
			WithStatusCode(200),
		)
		if !v.IsEmpty() {
			return v
		}
	}

	return nil
}
