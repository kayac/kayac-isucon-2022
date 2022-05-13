package bench

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"encoding/json"

	"github.com/isucon/isucandar/agent"
	"github.com/samber/lo"
)

type Model interface {
	User | Song
}

func LoadFromJSONFile[T Model](jsonFile string) ([]*T, error) {
	// 引数に渡されたファイルを開く
	file, err := os.Open(jsonFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	objects := make([]*T, 0, 10000) // 大きく確保しておく
	// JSON 形式としてデコード
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&objects); err != nil {
		if err != io.EOF {
			return nil, fmt.Errorf("failed to decode json: %w", err)
		}
	}
	return objects, nil
}

type Users []*User

func (us Users) Choice() *User {
	return lo.Sample(us)
}

type User struct {
	mu sync.RWMutex

	Account     string `json:"account"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	IsBan       bool   `json:"is_ban"`
	IsHeavy     bool   `json:"is_heavy"`

	Agent *agent.Agent
}

func (u *User) GetAgent(o Option) (*agent.Agent, error) {
	u.mu.RLock()
	a := u.Agent
	u.mu.RUnlock()
	if a != nil {
		return a, nil
	}

	u.mu.Lock()
	defer u.mu.Unlock()

	a, err := o.NewAgent(false)
	if err != nil {
		return nil, err
	}
	u.Agent = a
	return a, nil
}

type Song struct {
	ULID       string `json:"ulid"`
	Title      string `json:"title"`
	ArtistName string `json:"artist_name"`
	ArtistID   int64  `json:"artist_id"`
}

type Songs []*Song

func (ss Songs) ULIDs() []string {
	ids := make([]string, 0, len(ss))
	for _, s := range ss {
		ids = append(ids, s.ULID)
	}
	return ids
}

func GenerateUserAccount() string {
	return fmt.Sprintf("user-%s", newULID())
}

type Playlist struct {
	ULID            string    `json:"ulid"`
	Name            string    `json:"name"`
	UserDisplayName string    `json:"user_display_name"`
	UserAccount     string    `json:"user_account"`
	SongCount       int       `json:"song_count"`
	Songs           Songs     `json:"-"`
	FavoriteCount   int       `json:"favorite_count"`
	IsFavorited     bool      `json:"is_favorited"`
	IsPublic        bool      `json:"is_public"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
