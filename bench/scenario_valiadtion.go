package bench

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/isucon/isucandar"
	"github.com/samber/lo"
)

// 整合性検証シナリオ
// 自分で作ったplaylistを直後に削除したりするので、並列で実行するとfavした他人のplaylistが削除されて壊れる可能性がある
// 負荷テスト中には実行してはいけない
func (s *Scenario) ValidationScenario(ctx context.Context, step *isucandar.BenchmarkStep) error {
	report := timeReporter("validation")
	defer report()

	ContestantLogger.Println("整合性チェックを開始します")
	defer ContestantLogger.Printf("整合性チェックを終了します")

	ag, _ := s.Option.NewAgent(false)
	{
		// GET /
		res, err := GetRootAction(ctx, ag)
		v := ValidateResponse("トップページ", step, res, err, WithStatusCode(200, 304))
		if !v.IsEmpty() {
			return v
		}
	}

	var user *User
	{
		// ユーザー作成
		u, res, err := SignupAction(ctx, ag)
		user = u
		v := ValidateResponse("新規ユーザー登録",
			step, res, err,
			WithStatusCode(200),
			WithSuccessResponse[ResponseAPIBase](),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	{
		// ログイン
		res, err := LoginAction(ctx, user, ag)
		v := ValidateResponse("ログイン",
			step, res, err,
			WithStatusCode(200),
			WithSuccessResponse[ResponseAPIBase](),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	var topPlaylist Playlist
	{
		// 人気プレイリスト一覧
		res, err := GetPopularPlaylistsAction(ctx, ag)
		v := ValidateResponse("人気プレイリスト一覧",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIGetRecentPlaylists) error {
				if len(r.Playlists) > 100 {
					return fmt.Errorf("人気プレイリスト一覧が100より多くあります %d", len(r.Playlists))
				}
				if len(r.Playlists) < 50 {
					return fmt.Errorf("人気プレイリスト一覧が少なすぎます %d", len(r.Playlists))
				}
				topPlaylist = r.Playlists[0]
				var prevFavs int
				for _, p := range r.Playlists {
					if prevFavs == 0 {
						prevFavs = p.FavoriteCount
					}
					if prevFavs < p.FavoriteCount {
						return fmt.Errorf("人気プレイリスト一覧のfav数が降順になっていません %d < %d", prevFavs, p.FavoriteCount)
					}
					prevFavs = p.FavoriteCount
					if p.IsPublic == false {
						return fmt.Errorf("人気プレイリスト一覧に非公開プレイリストが含まれています")
					}
				}
				return nil
			}),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	{
		// favをつける
		topPlaylist.IsFavorited = true
		res, err := FavoritePlaylistAction(ctx, &topPlaylist, ag)
		v := ValidateResponse("人気のプレイリストにfavする",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIGetPlaylist) error {
				if !r.Playlist.IsFavorited {
					return fmt.Errorf("プレイリストのfavがついていません")
				}
				if r.Playlist.FavoriteCount == 0 {
					return fmt.Errorf("プレイリストのfav数が0です")
				}
				return nil
			}),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	{
		// 人気プレイリスト一覧
		res, err := GetPopularPlaylistsAction(ctx, ag)
		v := ValidateResponse("人気プレイリスト一覧 favがついている",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIGetRecentPlaylists) error {
				if len(r.Playlists) > 100 {
					return fmt.Errorf("人気プレイリスト一覧が100より多くあります %d", len(r.Playlists))
				}
				if len(r.Playlists) < 50 {
					return fmt.Errorf("人気プレイリスト一覧が少なすぎます %d", len(r.Playlists))
				}
				var found bool
				for _, p := range r.Playlists {
					if p.FavoriteCount == 0 {
						return fmt.Errorf("人気プレイリストのfav数が0です")
					}
					if topPlaylist.ULID == p.ULID {
						if !p.IsFavorited {
							return fmt.Errorf("favしたプレイリストにfavがついていません")
						}
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("favしたプレイリストが人気プレイリスト一覧にありません")
				}
				return nil
			}),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	{
		// プレイリスト一覧 作成済み0, fav済み1
		res, err := GetPlaylistsAction(ctx, ag)
		v := ValidateResponse("ログイン後自分のプレイリスト一覧を取得",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIGetPlaylists) error {
				if len(r.CreatedPlaylists) != 0 {
					return fmt.Errorf("作成済みプレイリストは0件のはずですが、%d件あります", len(r.CreatedPlaylists))
				}
				if len(r.FavoritedPlaylists) != 1 {
					return fmt.Errorf("fav済みプレイリストは1件のはずですが、%d件あります", len(r.FavoritedPlaylists))
				}
				return nil
			}),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	var playlistULID, playlistName string
	{
		// 新規プレイリスト作成
		playlistName = DisplayName()
		res, err := AddPlayistAction(ctx, playlistName, ag)
		v := ValidateResponse("新規プレイリスト作成",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIAddPlaylist) error {
				if r.PlaylistULID == "" {
					return fmt.Errorf("プレイリストのULIDが空です")
				}
				playlistULID = r.PlaylistULID
				return nil
			}),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	{
		// プレイリスト一覧 作成済み1, fav済み1
		res, err := GetPlaylistsAction(ctx, ag)
		v := ValidateResponse("自分のプレイリスト一覧に作成済みプレイリストが含まれている",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIGetPlaylists) error {
				if len(r.CreatedPlaylists) != 1 {
					return fmt.Errorf("作成済みプレイリストは1件のはずですが、%d件あります", len(r.CreatedPlaylists))
				}
				if r.CreatedPlaylists[0].ULID != playlistULID {
					return fmt.Errorf("作成されたプレイリストのULIDが一致しません")
				}
				if len(r.FavoritedPlaylists) != 1 {
					return fmt.Errorf("fav済みプレイリストは1件のはずですが、%d件あります", len(r.FavoritedPlaylists))
				}
				return nil
			}),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	var playlist Playlist
	{
		// プレイリスト詳細 作成直後なので空
		res, err := GetPlaylistAction(ctx, playlistULID, ag)
		v := ValidateResponse("作成したプレイリスト詳細",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIGetPlaylist) error {
				if r.Playlist.ULID != playlistULID {
					return fmt.Errorf("プレイリストのULIDが一致しません")
				}
				if r.Playlist.SongCount != 0 {
					return fmt.Errorf("プレイリストの曲数が0ではありません")
				}
				if r.Playlist.UserDisplayName != user.DisplayName {
					return fmt.Errorf(
						"プレイリストのdisplay_name %s が自分自身のdisplay_name %s と一致しません",
						r.Playlist.UserDisplayName,
						user.DisplayName,
					)
				}
				if r.Playlist.Name != playlistName {
					return fmt.Errorf("プレイリストの名前 %s が作成したプレイリストの名前 %s と一致しません", r.Playlist.Name, playlistName)
				}
				playlist = r.Playlist
				return nil
			}),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	addSongs := lo.Samples(s.Songs, rand.Intn(4))
	{
		// プレイリスト更新、n曲追加 公開にする
		playlist.Songs = append(playlist.Songs, addSongs...)
		playlist.IsPublic = true
		res, err := UpdatePlayistAction(ctx, &playlist, ag)
		v := ValidateResponse("プレイリスト更新 曲を追加して公開にする",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIGetPlaylist) error {
				if r.Playlist.ULID != playlistULID {
					return fmt.Errorf("曲を追加したプレイリストのULIDとレスポンスが一致しません")
				}
				if r.Playlist.SongCount != len(addSongs) {
					return fmt.Errorf("プレイリストの曲数 %d が追加した曲数 %d と一致しません", r.Playlist.SongCount, len(addSongs))
				}
				expectd := playlist.Songs.ULIDs()
				for i, ret := range r.Playlist.Songs.ULIDs() {
					if ret != expectd[i] {
						return fmt.Errorf("%d番目に追加した曲のUILD %s がレスポンス %s と一致しません", i, expectd[i], ret)
					}
				}
				return nil
			}),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	{
		// プレイリスト詳細 n曲含む
		res, err := GetPlaylistAction(ctx, playlistULID, ag)
		v := ValidateResponse("プレイリスト詳細に追加した曲が含まれている",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIGetPlaylist) error {
				if r.Playlist.ULID != playlistULID {
					return fmt.Errorf("リクエストしたプレイリストのULIDとレスポンスが一致しません")
				}
				if !r.Playlist.IsPublic {
					return fmt.Errorf("プレイリストが公開されていません")
				}
				if r.Playlist.SongCount != len(addSongs) {
					return fmt.Errorf("プレイリストの曲数 %d が追加した曲数 %d と一致しません", r.Playlist.SongCount, len(addSongs))
				}
				expectd := playlist.Songs.ULIDs()
				for i, ret := range r.Playlist.Songs.ULIDs() {
					if ret != expectd[i] {
						return fmt.Errorf("%d番目に追加した曲のUILD %s がレスポンス %s と一致しません", i, expectd[i], ret)
					}
				}
				return nil
			}),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	{
		// favをつける
		playlist.IsFavorited = true
		res, err := FavoritePlaylistAction(ctx, &playlist, ag)
		v := ValidateResponse("プレイリストにfavする",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIGetPlaylist) error {
				if !r.Playlist.IsFavorited {
					return fmt.Errorf("プレイリストのfavがついていません")
				}
				if r.Playlist.FavoriteCount == 0 {
					return fmt.Errorf("プレイリストのfav数が0です")
				}
				return nil
			}),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	{
		// 最新プレイリスト一覧 今投稿したプレイリストが含まれている & favがついている
		// 自分がfavしていないプレイリストは is_favorited が false
		res, err := GetRecentPlaylistsAction(ctx, ag)
		v := ValidateResponse("最新プレイリスト一覧に公開した自分のプレイリストが含まれている",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIGetRecentPlaylists) error {
				var found bool
				var prevCreatedAt time.Time
				if len(r.Playlists) != 100 {
					return fmt.Errorf("playlists must be 100")
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
						// この時点では2件しかfavしていないはず
						if p.ULID != playlistULID && p.ULID != topPlaylist.ULID {
							return fmt.Errorf("favしていないはずのプレイリストがfavされています")
						}
					}
					if p.ULID == playlistULID {
						found = true
						if p.SongCount != len(addSongs) {
							return fmt.Errorf("プレイリスト %s には %d曲含まれているはずですが%d曲です", p.ULID, len(addSongs), p.SongCount)
						}
						if p.FavoriteCount == 0 {
							return fmt.Errorf("プレイリスト %s はfavされているはずですがfav数が0です", p.ULID)
						}
						if p.IsFavorited != true {
							return fmt.Errorf("プレイリスト %s はfavされているはずですがfavされていません", p.ULID)
						}
					}
				}
				if !found {
					return fmt.Errorf("作成したプレイリスト %s が最新プレイリスト一覧に含まれていません", playlistULID)
				}
				return nil
			}),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	{
		// プレイリスト一覧 作成済み1, fav済み2
		res, err := GetPlaylistsAction(ctx, ag)
		v := ValidateResponse("プレイリスト詳細に作成済みとfav済みがある",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIGetPlaylists) error {
				if len(r.CreatedPlaylists) != 1 {
					return fmt.Errorf("作成済みプレイリストは1つであるはずですが %d 件です", len(r.CreatedPlaylists))
				}
				if r.CreatedPlaylists[0].ULID != playlistULID {
					return fmt.Errorf("作成済みプレイリストのULIDが一致しません %s != %s", r.CreatedPlaylists[0].ULID, playlistULID)
				}
				if len(r.FavoritedPlaylists) != 2 {
					return fmt.Errorf("fav済みプレイリストは2つであるはずですが %d 件です", len(r.FavoritedPlaylists))
				}
				var found bool
				for _, p := range r.FavoritedPlaylists {
					if p.ULID == playlistULID {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("favしたプレイリストがありません")
				}
				return nil
			}),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	{
		// プレイリスト詳細 fav済み
		res, err := GetPlaylistAction(ctx, playlistULID, ag)
		v := ValidateResponse("プレイリスト詳細でfav済みになっている",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIGetPlaylist) error {
				if r.Playlist.ULID != playlistULID {
					return fmt.Errorf("プレイリストのULIDが一致しません %s != %s", r.Playlist.ULID, playlistULID)
				}
				if !r.Playlist.IsFavorited {
					return fmt.Errorf("プレイリスト %s がfavされているはずですがfavされていません", r.Playlist.ULID)
				}
				if r.Playlist.FavoriteCount == 0 {
					return fmt.Errorf("プレイリスト %s はfavされているはずですが0件です", r.Playlist.ULID)
				}
				return nil
			}),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	{
		// favをはずす
		playlist.IsFavorited = false
		res, err := FavoritePlaylistAction(ctx, &playlist, ag)
		v := ValidateResponse("プレイリストのfavをはずす",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIGetPlaylist) error {
				if r.Playlist.ULID != playlistULID {
					return fmt.Errorf("プレイリストのULIDが一致しません %s != %s", r.Playlist.ULID, playlistULID)
				}
				if r.Playlist.IsFavorited {
					return fmt.Errorf("プレイリスト %s はfavされていないはずですがfavされています", r.Playlist.ULID)
				}
				return nil
			}),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	{
		// プレイリスト詳細 favなし
		res, err := GetPlaylistAction(ctx, playlistULID, ag)
		v := ValidateResponse("プレイリスト詳細でfavなしになっている",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIGetPlaylist) error {
				if r.Playlist.ULID != playlistULID {
					return fmt.Errorf("プレイリストのULIDが一致しません %s != %s", r.Playlist.ULID, playlistULID)
				}
				if r.Playlist.IsFavorited {
					return fmt.Errorf("プレイリスト %s はfavされていないはずですがfavされています", r.Playlist.ULID)
				}
				return nil
			}),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	{
		// プレイリスト一覧 作成済みにはあるがfav済みにはない
		res, err := GetPlaylistsAction(ctx, ag)
		v := ValidateResponse("プレイリスト一覧 作成済みがあるがfav済みはない",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIGetPlaylists) error {
				var found bool
				for _, p := range r.CreatedPlaylists {
					if p.ULID == playlist.ULID {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("プレイリスト %s が作成済みプレイリスト一覧にありません", playlist.ULID)
				}
				for _, p := range r.FavoritedPlaylists {
					if p.ULID == playlist.ULID {
						return fmt.Errorf("プレイリスト %s がfav済みプレイリスト一覧にないはずなのにあります", playlist.ULID)
					}
				}
				return nil
			}),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	{
		// 作成したプレイリスト削除
		res, err := DeletePlaylistAction(ctx, &playlist, ag)
		v := ValidateResponse("プレイリスト削除",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse[ResponseAPIBase](),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	{
		// プレイリスト詳細 404
		res, err := GetPlaylistAction(ctx, playlistULID, ag)
		v := ValidateResponse("削除されたプレイリストは404になる",
			step, res, err,
			WithStatusCode(404),
			WithCacheControlPrivate(),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	{
		// プレイリスト一覧 作成済みとfav済みに削除したものが含まれていない
		res, err := GetPlaylistsAction(ctx, ag)
		v := ValidateResponse("プレイリスト一覧に削除したプレイリストが含まれていない",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIGetPlaylists) error {
				for _, p := range r.CreatedPlaylists {
					if p.ULID == playlist.ULID {
						return fmt.Errorf("削除したプレイリスト %s が作成済みプレイリスト一覧にあります", playlist.ULID)
					}
				}
				for _, p := range r.FavoritedPlaylists {
					if p.ULID == playlist.ULID {
						return fmt.Errorf("削除したプレイリスト %s がfav済みプレイリスト一覧にあります", playlist.ULID)
					}
				}
				return nil
			}),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	var othersPlaylist Playlist
	{
		// 最新プレイリスト一覧 今削除したプレイリストが含まれていない
		res, err := GetRecentPlaylistsAction(ctx, ag)
		v := ValidateResponse("最新プレイリスト一覧に削除済みプレイリストが含まれていない",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIGetRecentPlaylists) error {
				for _, p := range r.Playlists {
					if p.ULID == playlistULID {
						return fmt.Errorf("削除したプレイリスト %s が最新プレイリスト一覧にあります", playlist.ULID)
					}
					if p.UserDisplayName != user.DisplayName {
						p := p
						othersPlaylist = p // 他人のをfavするために使う
					}
				}
				return nil
			}),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	{
		// 他人のにfavをつける
		othersPlaylist.IsFavorited = true
		res, err := FavoritePlaylistAction(ctx, &othersPlaylist, ag)
		v := ValidateResponse("自分以外のプレイリストをfavできる",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIGetPlaylist) error {
				if r.Playlist.ULID != othersPlaylist.ULID {
					return fmt.Errorf("favしたプレイリスト %s がレスポンスと一致しません", othersPlaylist.ULID)
				}
				if !r.Playlist.IsFavorited {
					return fmt.Errorf("favしたプレイリスト %s がfav済みになっていません", othersPlaylist.ULID)
				}
				if r.Playlist.FavoriteCount == 0 {
					return fmt.Errorf("favしたプレイリスト %s のfav数が0です", othersPlaylist.ULID)
				}
				return nil
			}),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	{
		// ログアウト
		res, err := LogoutAction(ctx, user, ag)
		v := ValidateResponse("ログアウトできる",
			step, res, err,
			WithStatusCode(200),
			WithSuccessResponse[ResponseAPIBase](),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	// ログアウト後1秒待ってみる
	time.Sleep(time.Second)
	{
		// プレイリスト一覧 ログアウト後には見えない
		res, err := GetPlaylistsAction(ctx, ag)
		v := ValidateResponse("非ログイン状態では自分のプレイリストは見えない",
			step, res, err,
			WithStatusCode(401),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	{
		// プレイリスト詳細 ログアウトしているのでfavはついていない
		res, err := GetPlaylistAction(ctx, othersPlaylist.ULID, ag)
		v := ValidateResponse("非ログイン状態でプレイリスト詳細を取得した場合 favは常にfalse",
			step, res, err,
			WithStatusCode(200, 304),
			WithSuccessResponse(func(r ResponseAPIGetPlaylist) error {
				if r.Playlist.ULID != othersPlaylist.ULID {
					return fmt.Errorf("リクエストしたプレイリスト %s がレスポンスと一致しません", othersPlaylist.ULID)
				}
				if r.Playlist.IsFavorited {
					return fmt.Errorf("非ログイン状態でリクエストしたプレイリスト %s がfav済みになっています", othersPlaylist.ULID)
				}
				return nil
			}),
		)
		if !v.IsEmpty() {
			return v
		}
	}

	// これからbanされるuser
	var banUser *User
	{
		// ユーザー作成
		u, res, err := SignupAction(ctx, ag)
		banUser = u
		defer res.Body.Close()
		v := ValidateResponse("ユーザー新規登録",
			step, res, err,
			WithStatusCode(200),
			WithSuccessResponse[ResponseAPIBase](),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	banUserAg, _ := banUser.GetAgent(s.Option)
	{
		res, err := LoginAction(ctx, banUser, banUserAg)
		v := ValidateResponse("ユーザーログイン",
			step, res, err,
			WithStatusCode(200),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	time.Sleep(100 * time.Millisecond)
	var banUserPlaylist Playlist
	{
		// banされるuserの新規プレイリスト作成
		playlistName = DisplayName()
		res, err := AddPlayistAction(ctx, playlistName, banUserAg)
		v := ValidateResponse("新規プレイリスト作成",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIAddPlaylist) error {
				banUserPlaylist = Playlist{
					Name:     playlistName,
					ULID:     r.PlaylistULID,
					Songs:    lo.Samples(s.Songs, rand.Intn(40)),
					IsPublic: true,
				}
				return nil
			}),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	{
		res, err := UpdatePlayistAction(ctx, &banUserPlaylist, banUserAg)
		v := ValidateResponse("プレイリストを更新",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIGetPlaylist) error {
				return nil
			}),
		)
		if !v.IsEmpty() {
			return v
		}
	}

	adminAg, _ := s.AdminUser.GetAgent(s.Option)
	{
		admin := s.AdminUser
		res, err := LoginAction(ctx, admin, adminAg)
		v := ValidateResponse("adminがログイン",
			step, res, err,
			WithStatusCode(200),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	{
		res, err := GetPlaylistAction(ctx, banUserPlaylist.ULID, adminAg)
		v := ValidateResponse("プレイリスト詳細がある",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	{
		res, err := AdminBanAction(ctx, banUser, true, adminAg)
		v := ValidateResponse("adminがユーザーをban",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAdminBan) error {
				if r.UserAccount != banUser.Account {
					return fmt.Errorf("banしたユーザー %s がレスポンス %s と一致しません", banUser.Account, r.UserAccount)
				}
				if !r.IsBan {
					return fmt.Errorf("ユーザー %s がbanされていません", banUser.Account)
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
		res, err := GetPlaylistsAction(ctx, banUserAg)
		v := ValidateResponse("banされたユーザーは自分のプレイリストを取得できない",
			step, res, err,
			WithStatusCode(401),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	{
		res, err := GetPlaylistAction(ctx, banUserPlaylist.ULID, adminAg)
		v := ValidateResponse("banされたユーザーのプレイリスト詳細が404になる",
			step, res, err,
			WithStatusCode(404),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	{
		res, err := GetRecentPlaylistsAction(ctx, adminAg)
		v := ValidateResponse("最新プレイリスト一覧にbanされたユーザーのプレイリストはない",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAPIGetRecentPlaylists) error {
				var found bool
				if len(r.Playlists) != 100 {
					return fmt.Errorf("最近作成されたプレイリストは100件あるはずですが%d件です", len(r.Playlists))
				}
				for _, p := range r.Playlists {
					if p.ULID == banUserPlaylist.ULID {
						found = true
						break
					}
				}
				if found {
					return fmt.Errorf("最近作成したプレイリストにbanされたユーザーのプレイリストが含まれています")
				}
				return nil
			}),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	{
		// admin unban
		res, err := AdminBanAction(ctx, banUser, false, adminAg)
		v := ValidateResponse("adminがbanを解除",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
			WithSuccessResponse(func(r ResponseAdminBan) error {
				if r.UserAccount != banUser.Account {
					return fmt.Errorf("ban解除したユーザー %s がレスポンス %s と一致しません", banUser.Account, r.UserAccount)
				}
				if r.IsBan {
					return fmt.Errorf("ban解除したユーザー %s がまだbanされています", banUser.Account)
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
		res, err := GetPlaylistAction(ctx, banUserPlaylist.ULID, adminAg)
		v := ValidateResponse("ban解除されたユーザーのプレイリスト詳細",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	{
		res, err := LoginAction(ctx, banUser, banUserAg)
		v := ValidateResponse("ban解除されたユーザーがログイン",
			step, res, err,
			WithStatusCode(200),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	{
		res, err := GetPlaylistsAction(ctx, banUserAg)
		v := ValidateResponse("ban解除されたユーザーが自分のプレイリストを取得",
			step, res, err,
			WithStatusCode(200),
			WithCacheControlPrivate(),
		)
		if !v.IsEmpty() {
			return v
		}
	}
	return nil
}
