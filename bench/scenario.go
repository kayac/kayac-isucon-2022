package bench

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucandar/score"
	"github.com/isucon/isucandar/worker"
)

var (
	Debug     = false
	MaxErrors = 30
)

const (
	ErrFailedLoadJSON failure.StringCode = "load-json"
	ErrCannotNewAgent failure.StringCode = "agent"
	ErrInvalidRequest failure.StringCode = "request"
)

// シナリオで発生するスコアのタグ
const (
	ScoreGETRoot             score.ScoreTag = "GET /"
	ScoreSignup              score.ScoreTag = "POST /api/signup"
	ScoreLogin               score.ScoreTag = "POST /api/login"
	ScoreLogout              score.ScoreTag = "POST /api/logout"
	ScoreGetPlaylist         score.ScoreTag = "GET /api/playlist/{}"
	ScoreGetPlaylists        score.ScoreTag = "GET /api/playlists"
	ScoreGetRecentPlaylists  score.ScoreTag = "GET /api/recent_playlists"
	ScoreGetPopularPlaylists score.ScoreTag = "GET /api/popular_playlists"
	ScoreAddPlaylist         score.ScoreTag = "POST /api/playlist/{}/add"
	ScoreUpdatePlaylist      score.ScoreTag = "POST /api/playlist/{}/update"
	ScoreFavoritePlaylist    score.ScoreTag = "POST /api/playlist/favorite"
	ScoreAdminBan            score.ScoreTag = "POST /api/admin/user/ban"

	ScoreGetRecentPlaylistsLogin  score.ScoreTag = "GET /api/recent_playlists (login)"
	ScoreGetPopularPlaylistsLogin score.ScoreTag = "GET /api/popular_playlists (login)"
)

// オプションと全データを持つシナリオ構造体
type Scenario struct {
	mu sync.RWMutex

	Option Option

	NormalUsers mapset.Set
	HeavyUsers  mapset.Set
	BannedUsers mapset.Set
	Songs       Songs
	AdminUser   *User

	lastPlaylistCreatedAt   time.Time
	rateGetPopularPlaylists int32

	Errors failure.Errors
}

func (s *Scenario) SetRateGetPopularPlaylists(rate int32) {
	AdminLogger.Printf("set rate get popular playlists to %d", rate)
	atomic.StoreInt32(&s.rateGetPopularPlaylists, rate)
}

func (s *Scenario) RateGetPopularPlaylists() int32 {
	return atomic.LoadInt32(&s.rateGetPopularPlaylists)
}

func (s *Scenario) SetLastPublicPlaylistCreatedAt(t time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastPlaylistCreatedAt = t
}

func (s *Scenario) LastPublicPlaylistCreatedAt() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastPlaylistCreatedAt
}

// isucandar.PrepeareScenario を満たすメソッド
// isucandar.Benchmark の Prepare ステップで実行される
func (s *Scenario) Prepare(ctx context.Context, step *isucandar.BenchmarkStep) error {
	s.SetRateGetPopularPlaylists(10)

	// Userのロード
	us, err := LoadFromJSONFile[User](filepath.Join(s.Option.DataDir, "users.json"))
	if err != nil {
		return err
	}
	AdminLogger.Printf("%d users loaded", len(us))
	s.AdminUser = &User{Account: "adminuser", Password: "adminpass"}
	s.NormalUsers = mapset.NewSet()
	s.HeavyUsers = mapset.NewSet()
	s.BannedUsers = mapset.NewSet()
	for _, u := range us {
		if u.IsBan {
			s.BannedUsers.Add(u)
			// banされてたらheavyにいれない
			continue
		}
		if u.IsHeavy {
			s.HeavyUsers.Add(u)
		} else {
			s.NormalUsers.Add(u)
		}
	}
	AdminLogger.Printf(
		"normal:%d heavy:%d banned:%d",
		s.NormalUsers.Cardinality(),
		s.HeavyUsers.Cardinality(),
		s.BannedUsers.Cardinality(),
	)

	if s.Songs, err = LoadFromJSONFile[Song](filepath.Join(s.Option.DataDir, "songs.json")); err != nil {
		return err
	}
	AdminLogger.Printf("%d songs loaded", len(s.Songs))

	// Prepareは60秒以内に完了
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	// GET /initialize 用ユーザーエージェントの生成
	ag, err := s.Option.NewAgent(true)
	if err != nil {
		return failure.NewError(ErrCannotNewAgent, err)
	}

	if s.Option.SkipPrepare {
		return nil
	}

	debug := Debug
	defer func() {
		Debug = debug
	}()
	Debug = true // prepareは常にデバッグログを出す

	// POST /initialize へ初期化リクエスト実行
	res, err := GetInitializeAction(ctx, ag)
	if v := ValidateResponse("初期化", step, res, err, WithStatusCode(200)); !v.IsEmpty() {
		return fmt.Errorf("初期化リクエストに失敗しました %v", v)
	}

	// 検証シナリオを1回まわす
	if err := s.ValidationScenario(ctx, step); err != nil {
		return fmt.Errorf("整合性チェックに失敗しました")
	}

	ContestantLogger.Printf("整合性チェックに成功しました")
	return nil
}

// isucandar.PrepeareScenario を満たすメソッド
// isucandar.Benchmark の Load ステップで実行される
func (s *Scenario) Load(ctx context.Context, step *isucandar.BenchmarkStep) error {
	if s.Option.PrepareOnly {
		return nil
	}
	ContestantLogger.Println("負荷テストを開始します")
	defer ContestantLogger.Println("負荷テストを終了します")
	wg := &sync.WaitGroup{}

	// 通常シナリオ
	normalCase, err := s.NormalWorker(step, 1)
	if err != nil {
		return err
	}
	// favolite を追加するケースのシナリオ
	favCase, err := s.FavoriteWorker(step, 1)
	if err != nil {
		return err
	}
	// banされたユーザーのシナリオ
	bannedCase, err := s.BannedWorker(step, 1)
	if err != nil {
		return err
	}
	// 匿名シナリオ
	anonCase, err := s.AnonWorker(step, 1)
	if err != nil {
		return err
	}

	workers := []*worker.Worker{
		normalCase,
		favCase,
		bannedCase,
		anonCase,
	}
	for _, w := range workers {
		wg.Add(1)
		worker := w
		go func() {
			defer wg.Done()
			worker.Process(ctx)
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.loadAdjustor(ctx, step, normalCase, favCase, anonCase)
	}()
	wg.Wait()
	return nil
}

func (s *Scenario) loadAdjustor(ctx context.Context, step *isucandar.BenchmarkStep, workers ...*worker.Worker) {
	tk := time.NewTicker(time.Second)
	var prevErrors int64
	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
		}
		errors := step.Result().Errors.Count()
		total := errors["load"]
		if total >= int64(MaxErrors) {
			ContestantLogger.Printf("負荷テストを打ち切ります (エラー数:%d)", total)
			AdminLogger.Printf("%#v", errors)
			step.Result().Score.Close()
			step.Cancel()
			return
		}
		addParallels := int32(1)
		if diff := total - prevErrors; diff > 0 {
			ContestantLogger.Printf("エラーが%d件増えました(現在%d件)", diff, total)
		} else {
			ContestantLogger.Println("ユーザーが増えます")

			// popularの取得確率も増やしていく
			rate := s.RateGetPopularPlaylists()
			if rate < 100 {
				// 2回増えると2倍になるペースで増加
				// 10 -> 14 -> 20 -> 28 -> 40 ...
				newRate := int32(float32(rate) * 1.41422)
				if newRate >= 100 {
					newRate = 100
				}
				s.SetRateGetPopularPlaylists(newRate)
			}

			addParallels = 1
		}
		for _, w := range workers {
			w.AddParallelism(addParallels)
		}
		prevErrors = total
	}
}

func (s *Scenario) ChoiceUser(ctx context.Context, pool mapset.Set) (*User, func()) {
	for {
		select {
		case <-ctx.Done():
			return nil, func() {}
		default:
		}
		if u := pool.Pop(); u == nil {
			time.Sleep(time.Second)
			continue
		} else {
			user := u.(*User)
			return user, func() {
				ag, _ := user.GetAgent(s.Option)
				ag.HttpClient.CloseIdleConnections()
				pool.Add(u)
			}
		}
	}
}

var nullFunc = func() {}

func timeReporter(name string) func() {
	if !Debug {
		return nullFunc
	}
	start := time.Now()
	return func() {
		AdminLogger.Printf("Scenario:%s elapsed:%s", name, time.Since(start))
	}
}
