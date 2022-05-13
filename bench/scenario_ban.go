package bench

import (
	"context"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/worker"
)

func (s *Scenario) BannedWorker(step *isucandar.BenchmarkStep, p int32) (*worker.Worker, error) {
	w, err := worker.NewWorker(func(ctx context.Context, _ int) {
		timer := time.NewTimer(3 * time.Second)
		select {
		case <-ctx.Done():
		case <-timer.C:
		}
		s.BannedScenario(ctx, step)
	},
		// 無限回繰り返す
		worker.WithInfinityLoop(),
		worker.WithMaxParallelism(10),
	)
	if err != nil {
		return nil, err
	}
	w.SetParallelism(p)
	return w, nil
}

// Ban済みUserのシナリオ
func (s *Scenario) BannedScenario(ctx context.Context, step *isucandar.BenchmarkStep) error {
	report := timeReporter("banned")
	defer report()

	user, release := s.ChoiceUser(ctx, s.BannedUsers)
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
			WithStatusCode(401),
			WithErrorResponse[ResponseAPIBase](),
		)
		if v.IsEmpty() {
			step.AddScore(ScoreLogin)
		} else {
			return v
		}
	}
	return nil
}
