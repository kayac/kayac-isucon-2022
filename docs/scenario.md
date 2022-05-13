# ベンチシナリオ覚書

ポイント

- APIを網羅する
- POST直後のGETで結果確認
- 重複不可を守っていることをチェック

```
(もしあれば)静的ファイルチェック

[ユーザー作成、ログイン]
/api/signup (error, user_idに利用不可能な文字が含まれている)
/api/signup (error, passwordに利用不可能な文字が含まれている)
/api/signup (error, user_idが長すぎる)
/api/signup (error, passwordが長すぎる)
/api/signup (error, display_nameが長すぎる)
/api/signup (OK)
/api/login (error, パラメータ不足)
/api/login (error, 存在しないユーザー)
/api/login (error, password間違い)
/api/login (OK)
/api/logout (OK)
/api/playlists (error, ログインセッション無し)
/api/logout (error, ログインセッションなし) #TODO: cookieのセッション破棄させる処理しかないのであれば不要？
/api/login (OK)
/api/playlists (OK)

[ 非ログイン時チェック ]
/api/logout (ok)
/api/playlists (error, ログインセッション無し)
/api/playlist/add (error, ログインセッション無し)
/api/playlist/{:xxx} (OK, リストの表示はログイン不要)
/api/playlist/{:xxx}/update (error, ログインセッション無し)

[ プレイリストの更新 ]
/api/login (OK)
/api/playlist/add (error, プレイリスト名が長すぎる)
/api/playlist/add (OK)
/api/playlists (OK, 新規プレイリストが追加されていることを確認)
/api/playlist/{:xxx} (OK)
/api/playlist/{:xxx}/update (OK, 初回の曲追加)
/api/playlist/{:xxx} (OK, 内容があることを確認)
/api/playlist/{:xxx}/update (OK, 内容の変更なし)
/api/playlist/{:xxx}/update (error, 曲数が81曲以上)
/api/playlist/{:xxx}/update (OK, 曲重複)
/api/playlist/{:xxx}/update (OK, songsから曲削除)
/api/playlist/{:xxx} (OK, 消した曲が入っていないことを確認)
/api/playlist/{:xxx}/update (OK, 消した曲もう一度追加)
/api/playlist/{:xxx} (OK, 入れた曲が入っていることを確認)
/api/playlist/{:xxx}/update (OK, 曲の順序入れ替え)
/api/playlist/{:xxx} (OK, 曲順が正しいことを確認)

/api/playlist/{:xxx}/update (error, 他のユーザーのプレイリスト)

[ プレイリストの削除 ]
/api/playlist/{:xxx}/delete (OK)
/api/playlist/{:xxx} (error, 存在しないプレイリスト)
/api/playlists (OK, 消したプレイリストが存在しないことを確認)

/api/playlist/{:xxx}/delete (error, 他のユーザーのプレイリスト)

[ プレイリストのお気に入り]

WIP
```

たくさん回してリクエスト数カウントしてスコアになるかなあ
WIP
