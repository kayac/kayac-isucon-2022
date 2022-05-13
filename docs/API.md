## API

### 基本的なレスポンス型

- 成功時
```json
{
  "result": true,
  "status": 200
}
```

- 失敗時
```json
{
  "result": false,
  "status": 401,
  "error": "authentication required."
}
```

- 成功失敗に関わらず、必ず `"result"` キーと `"status"` キーが存在する
- `"result"` がfalseの場合は、`"error"` キーが存在し、理由文字列が入っている
- `"status"` は対応するHTTPステータスコードが入っている
- 30xでリダイレクトするときはレスポンスボディは存在しない
- APIによっては、上記の基本的レスポンスに加えてAPIの結果を格納するキーが存在する

### # POST `/api/signup`

アカウントを作成してセッションIDを返す
セッションIDはCookieに入れる

#### Request

key | value | note
--- | --- | ---
user_account | string | 一意なユーザーアカウント文字列
password | string |
display_name | string | プレイリスト作者欄等に表示される名前

user_account

- 既存ユーザーと重複NG
- 4文字以上36文字以内
- 半角小文字、数字、アンダーバー、ハイフンのみ `[a-z0-9_\-]`

password

- 8文字以上64文字以内
- 半角英数字、アンダーバー、ハイフンのみ `[a-zA-Z0-9_\-]`

display_name

- 2文字以上24文字以内

```json
{
  "user_account": "isucon",
  "password": "isuconpass",
  "display_name": "イスコン"
}
```

#### Response

key | value | note
--- | --- | ---

```json
{}
```

HTTP Header

```
Set-Cookie: session_id=isuconsession
```

### # POST `/api/login`

ログインしてセッションIDを返す

以下の場合はログインに失敗する(HTTP status 401)
- ユーザーが存在しない
- ユーザーがBANされている
- パスワードが間違っている

#### Request

key | value | note
--- | --- | ---
user_account | string |
password | string |

user_account

- 4文字以上191文字以内
- 半角小文字、数字、アンダーバー、ハイフンのみ `[a-z0-9_\-]`

password

- 8文字以上64文字以内
- 半角英数字、アンダーバー、ハイフンのみ `[a-zA-Z0-9_\-]`

```json
{
  "user_account": "isucon",
  "password": "isuconpass"
}
```

#### 通常ユーザーのResponse

key | value | note
--- | --- | ---

```json
{}
```

HTTP Header

```
Set-Cookie: session_id=isuconsession
```

#### BANユーザーの場合のResponse

```json
{
  "result": false,
  "status": 200,
  "error": "BAN済みユーザーです"
}
```

### # POST `/api/logout`

ログインセッションを破棄する

#### Request

key | value | note
--- | --- | ---

HTTP Header

session_id

- 空でも無効なセッションでもエラーを返さずOKとする

```
Cookie: session_id=isuconsession
```

#### Response

key | value | note
--- | --- | ---

```json
{}
```

### # GET `/api/recent_playlists`

全体のプレイリストを作成時刻が最新のものから100件返す

- 公開playlistしか含まれない
- 認証不要
- BANされているユーザーが作成したplaylistは含まれない
- 認証している場合
  - 自分が作成したplaylist、favoriteしたplaylistは is_favorited に正しい情報が入る
  - それ以外のplaylistでは is_favorited は常に false になる

#### Request

key | value | note
--- | --- | ---

#### Response

key | value | note
--- | --- | ---
playlists | playlist_summary[] | プレイリストの概要の配列


```json
{
  "playlists": [
    {
      "ulid": "801G018N01064WKJE8000000000",
      "name": "イスコンのプレイリスト",
      "user_display_name": "イスコン",
      "song_count": 10,
      "favorite_count": 3,
      "is_favorited": false,
      "is_public": true
    },
    {
      "ulid": "01G0B6EHG8T1YXJ9SRH02DHDRD",
      "name": "ツクエのプレイリスト",
      "user_display_name": "ツクエ",
      "song_count": 10,
      "favorite_count": 3,
      "is_favorited": true,
      "is_public": true
    },
  ],
}
```

### # GET `/api/popular_playlists`

全体のプレイリストをFavoriteが多い順に100件返す

- 公開playlistしか含まれない
- 認証不要
- BANされているユーザーが作成したplaylistは含まれない
- 認証している場合
  - 自分が作成したplaylist、favoriteしたplaylistは is_favorited に正しい情報が入る
  - それ以外のplaylistでは is_favorited は常に false になる

#### Request

key | value | note
--- | --- | ---

#### Response

key | value | note
--- | --- | ---
playlists | playlist_summary[] | プレイリストの概要の配列


```json
{
  "playlists": [
    {
      "ulid": "801G018N01064WKJE8000000000",
      "name": "イスコンのプレイリスト",
      "user_display_name": "イスコン",
      "song_count": 10,
      "favorite_count": 3,
      "is_favorited": false,
      "is_public": true
    },
    {
      "ulid": "01G0B6EHG8T1YXJ9SRH02DHDRD",
      "name": "ツクエのプレイリスト",
      "user_display_name": "ツクエ",
      "song_count": 10,
      "favorite_count": 3,
      "is_favorited": true,
      "is_public": true
    },
  ],
}
```

### # GET `/api/playlists`

ログイン中のユーザーがトップページに表示させるプレイリストを返す

- 自分が作成したプレイリスト一覧 作成日降順 最新100件まで
- favしたプレイリスト一覧 fav日降順 最新100件まで
  - 自分以外が作成した、非公開プレイリストは含まれない
  - 自分が作成した非公開プレイリストは含まれる
  - 作成したユーザーがbanされている場合は含まれない

この一覧にはそれぞれの曲の一覧は含まれない

- 認証必須

#### Request

key | value | note
--- | --- | ---

HTTP Header

```
Cookie: session_id=isuconsession
```

#### Response

key | value | note
--- | --- | ---
created_playlists | playlist_summary[] | 自分が作成したプレイリストの概要の配列
favorited_playlists | playlist_summary[] | おきにいりプレイリストの概要の配列

```json
{
  "created_playlists": [
    {
      "ulid": "801G018N01064WKJE8000000000",
      "name": "イスコンのプレイリスト",
      "user_display_name": "イスコン",
      "song_count": 10,
      "favorite_count": 3,
      "is_favorited": false,
      "is_public": true
    },
  ],
  "favorited_playlists": [
    {
      "ulid": "01G0B6EHG8T1YXJ9SRH02DHDRD",
      "name": "ツクエのプレイリスト",
      "user_display_name": "ツクエ",
      "song_count": 10,
      "favorite_count": 3,
      "is_favorited": true,
      "is_public": true
    },
  ],
}
```

### # POST `/api/playlist/add`

新しく空のプレイリストを作成する。is_public は常に false で作成される。

- 認証必須

#### Request

##### JSON Bodyとして渡す

- 全て必須パラメータ

key | value | note
--- | --- | ---
name | string | 作成するプレイリスト名<br>

name

- 2文字以上191文字以内

```json
{
  "name": "イスコンのプレイリスト"
}
```

HTTP Header

```
Cookie: session_id=isuconsession
```

#### Response

key | value | note
--- | --- | ---
playlist_ulid | string | 作成したプレイリストのulid

```json
{
  "playlist_ulid":  "801G018N01064WKJE8000000000",
}
```

### # GET `/api/playlist/{:playlist_ulid}`

指定したプレイリストの詳細を返す
プレイリストの中の曲の一覧を含む
曲の一覧はプレイリストを更新する際に指定した順で返る

- 認証なしでも叩けるが、is_favoritedは常にfalseになる
- プレイリストを作成したユーザーがBANされている場合、HTTP status 404

#### Request

##### URL parameterとして渡す

key | value | note
--- | --- | ---
playlist_ulid | string |

playlist_ulid

- `[A-Z0-9]` ULIDとして有効なで構成されていること
- 存在しないplaylist_ulidなら400エラー
- 対象playlistがpublicではない かつ ログインセッションが有効でないなら404エラー
- 対象playlistがpublicではない かつ ログインセッションが有効 かつ playlistの作成者が自分でなければ404エラー

```
"playlist_ulid": "801G018N01064WKJE8000000000"
```

HTTP Header

session_id: 任意

```
Cookie: session_id=isuconsession
```

#### Response

key | value | note
--- | --- | ---
playlist | playlist_detail | プレイリストの詳細

```json
{
  "playlist": {
    "ulid": "801G018N01064WKJE8000000000",
    "name": "イスコンのプレイリスト",
    "user_display_name": "イスコン",
    "song_count": 10,
    "songs": [
      {
        "ulid": "01G0180MS86400000000000000",
        "title": "椅子 on the floor",
        "artist": "ISU",
        "album": "ISU THE BEST",
        "track_number": 1,
        "is_public": true,
      },
    ],
    "favorite_count": 3,
    "is_favorited": false,
    "is_public": true
  }
}
```

### # POST `/api/playlist/{:playlist_ulid}/update`

プレイリストの内容を更新する
プレイリストの作成者しか編集できない

- 認証必須

#### Request

##### URL parameterとして渡す

- 必須パラメータ

key | value | note
--- | --- | ---
playlist_ulid | string |

playlist_ulid

- `[A-Z0-9]` ULIDとして有効なで構成されていること
- 存在しないplaylist_ulidなら404エラー
- 対象playlistの作成者が自分でなければ404エラー

```
"playlist_ulid": "801G018N01064WKJE8000000000"
```

##### JSON Bodyで渡す

すべて必須(欠けている場合は400エラーとする)

key | value | note
--- | --- | ---
name | string | プレイリスト名
song_ulids | int[] | song_ulidの配列
is_public | boolean | 公開ステータス

name

- すでに同じ名前のプレイリストを持っている場合は400エラー
- 2文字以上24文字以内

songs

- songsが81曲以上の場合は400エラー
- songsの中に重複する楽曲があれば400エラー

```json
{
  "name": "イスコンプレイリストその2",
  "song_ulids": [
    "01G04NHJ0JNVW66VZ0FEQ05S2X",
    "01G04NHJ0JQYJSJETTHDFM7BWE",
    "01G04NHJ0JSAZ6B18QM411JCXV"
  ],
  "is_public": false
}
```

HTTP Header

```
Cookie: session_id=isuconsession
```

#### Response

key | value | note
--- | --- | ---
playlist | playlist_detail | 更新後のplaylist

```json
{
  "playlist": {
    "ulid": "801G018N01064WKJE8000000000",
    "name": "イスコンのプレイリスト",
    "user_display_name": "イスコン",
    "song_count": 10,
    "songs": [
      {
        "ulid": "01G0180MS86400000000000000",
        "title": "椅子 on the floor",
        "artist": "ISU",
        "album": "ISU THE BEST",
        "track_number": 1,
        "is_public": true,
      }
    ],
    "favorite_count": 3,
    "is_favorited": false,
    "is_public": true
  }
}
```

### # POST `/api/playlist/{:playlist_ulid}/delete`

プレイリストを削除する  
プレイリストの作成者しか削除できない

- 認証必須

#### Request

##### URL parameterとして渡す

- 必須パラメータ

key | value | note
--- | --- | ---
playlist_ulid | string |

playlist_ulid

- `[A-Z0-9]` ULIDとして有効なで構成されていること
- 存在しないplaylist_ulidなら404エラー
- 対象playlistの作成者が自分でなければ400エラー

```
"playlist_ulid": "801G018N01064WKJE8000000000"
```

HTTP Header

session_id

- 必須ではない

```
Cookie: session_id=isuconsession
```

#### Response

key | value | note
--- | --- | ---

```json
{}
```

### # POST `/api/playlist/{:playlist_ulid}/favorite`

プレイリストのお気に入り登録状態を更新する  

- 認証必須
- 自分以外が作成したプレイリストの場合、以下の時は失敗する(HTTP status 404)
  - プレイリストがprivate
  - プレイリストを作成したユーザーがbanされている

#### Request

##### URL parameterとして渡す

key | value | note
--- | --- | ---
playlist_ulid | string |

playlist_ulid

- `[A-Z0-9]` ULIDとして有効なで構成されていること
- 存在しないplaylist_ulidなら404エラー
- 対象playlistがpublicではない かつ ログインセッションが有効でないなら404エラー
- 対象playlistがpublicではない かつ ログインセッションが有効 かつ playlistの作成者が自分でなければ404エラー
- fav状態が変わらないリクエストはエラーにせず200を返す

```
"playlist_ulid": "801G018N01064WKJE8000000000"
```

##### JSON Bodyとして渡す

- 全て必須パラメータ

key | value | note
--- | --- | ---
is_favorited | boolean | 更新後のお気に入り登録状態

favorite

- 更新前と更新後が変わらない場合も200 OKを返す

```json
{
  "is_favorited": true
}
```

HTTP Header

```
Cookie: session_id=isuconsession
```

#### Response

key | value | note
--- | --- | ---
playlist | playlist_detail | 更新後のplaylist

```json
{
  "playlist": {
    "ulid": "801G018N01064WKJE8000000000",
    "name": "イスコンのプレイリスト",
    "user_display_name": "イスコン",
    "song_count": 10,
    "songs": [
      {
        "ulid": "01G0180MS86400000000000000",
        "title": "椅子 on the floor",
        "artist": "ISU",
        "album": "ISU THE BEST",
        "track_number": 1,
        "is_public": true,
      }
    ],
    "favorite_count": 3,
    "is_favorited": false,
    "is_public": true
  }
}
```

### # POST `/api/admin/user/ban`

ユーザーのBAN状況を更新する

- 管理者ユーザーの認証必須
- BANされたユーザーは以下の状態になる
  - ログインに失敗する
  - 有効なログインセッションを持っていても、ログアウト以外の全てのAPIリクエストが失敗する
- BANされたユーザーが作成したプレイリストは、他のユーザーに対してのAPIレスポンスに含まれなくなる

このAPIの実行結果は3秒以内に他のAPIレスポンスに反映されている必要がある

#### Request

##### JSON Bodyとして渡す

- 全て必須パラメータ

key | value | note
--- | --- | ---
user_account | string | 対象ユーザー
is_ban | boolean | 指定ユーザーをBANに指定するか

user_ulid

- `[A-Z0-9]` ULIDとして有効なで構成されていること
- 存在しないuser_ulidなら400エラー

```json
{
  "user_account": "isucon",
  "is_ban": true
}
```

HTTP Header

```
Cookie: session_id=rootusersession
```

#### Response

key | value | note
--- | --- | ---
(なし) | user | 更新後のuser情報

```json
{
  "user_account": "isucon",
  "display_name": "イスコン",
  "is_ban": true
}
```

#### 管理者ユーザーでない場合のエラーレスポンス

```json
{
  "result": false,
  "status": 403,
  "error": "not admin user"
}
```

## 型定義

date は ISO8601 フォーマットの文字列とする

### # user

key | value | note
--- | --- | ---
user_account | string |
display_name | string | 表示名、プレイリスト作者などで使われる
is_ban | boolean | BANされているか
created_at | date | ユーザーを作成した日時

```json
{
  "user_account": "isucon",
  "display_name": "イスコン",
  "is_ban": false,
  "created_at": "2012-04-23T18:25:43.511Z",
}
```

### # song

key | value | note
--- | --- | ---
ulid | string | 曲固有の識別子
title | string | 曲名
artist | string | アーティスト名 id参照
album | string | アルバム名
track_number | int | album内での曲順 1 based
is_public | boolean | 公開中かどうか

```json
{
  "ulid": "01G0180MS86400000000000000",
  "title": "椅子 on the floor",
  "artist": "ISU",
  "album": "ISU THE BEST",
  "track_number": 1,
  "is_public": true,
}
```

### # playlist_summary

多数のプレイリストの一覧表示に利用する情報

key | value | note
--- | --- | ---
ulid | string | プレイリストの固有識別子 ULID
name | string |
user_display_name | string | 作成者のdisplay name
user_account | string | 作成者のaccount
song_count | int | プレイリスト内の曲数
favorite_count | int | プレイリストがお気に入りされた回数
is_favorited | boolean | 自分がプレイリストをお気に入り済みか
is_public | boolean | | 公開中かどうか
created_at | date | プレイリストを作成した日時
updated_at | date | プレイリストを最終更新した日時

```json
{
  "ulid": "801G018N01064WKJE8000000000",
  "name": "イスコンのプレイリスト",
  "user_display_name": "イスコン",
  "song_count": 10,
  "favorite_count": 3,
  "is_favorited": false,
  "is_public": true,
  "created_at": "2012-04-23T18:25:43.511Z",
  "updated_at":"2012-04-23T18:25:43.511Z"
}
```

### # playlist_detail

個々のプレイリストの内容

key | value | note
--- | --- | ---
ulid | string | プレイリストの固有識別子 ULID
name | string |
user_display_name | string | 作成者のdisplay name
song_count | int | プレイリスト内の曲数
songs | song[] | プレイリスト内の曲一覧
favorite_count | int | プレイリストがお気に入りされた回数
is_favorited | boolean | 自分がプレイリストをお気に入り済みか
is_public | boolean | 公開中かどうか
created_at | date | プレイリストを作成した日時
updated_at | date | プレイリストを最終更新した日時

```json
{
  "ulid": "801G018N01064WKJE8000000000",
  "name": "イスコンのプレイリスト",
  "user_display_name": "イスコン",
  "song_count": 10,
  "songs": [
    "((songの配列))"
  ],
  "favorite_count": 3,
  "is_favorited": false,
  "is_public": true,
  "created_at": "2012-04-23T18:25:43.511Z",
  "updated_at":"2012-04-23T18:25:43.511Z"
}
```
