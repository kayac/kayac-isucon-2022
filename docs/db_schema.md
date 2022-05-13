# DB schema

- スキーマ表現上、 boolean は TINYINT(2) で定義します
- numericなPrimary KeyはBIGINTにします(実用上、32bitでは足りないので!)
- utf8mb4 を採用するので、Index Lengthを考慮してVARCHARのサイズは191文字にします

### user

name | type | opts | note
--- | --- | --- | ---
account | varchar(191) | PRIMARY KEY | ユーザーが指定できるユーザーアカウント
display_name | varchar(191) | |
password_hash | varchar(191) | |
is_ban | boolean | | BANされている（無効）アカウントかどうか
created_at | timestamp | | ユーザーを作成した日時
last_logined_at | timestamp | | ユーザーが最終ログイン日時

### song

name | type | opts | note
--- | --- | --- | ---
id | bigint | PRIMARY KEY, AUTO_INCREMENT |
ulid | varchar(191) | | ユーザーから見えるsongのID ULID
title | varchar(191) | | 曲名
artist_id | bigint | | artist tableのID
album | varchar(191) | | アルバム名
track_number | int | | アルバム内の曲順
is_public | boolean | | 公開中かどうか

### artist

name | type | opts | note
--- | --- | --- | ---
id | bigint | PRIMARY KEY, AUTO_INCREMENT |
ulid | varchar(191) | | ユーザーから見えるartistのID ULID
name | varchar(191) | | アーティスト名

### playlist

name | type | opts | note
--- | --- | --- | ---
id | bigint | PRIMARY KEY, AUTO_INCREMENT |
ulid | varchar(191) | | ユーザーから見えるplaylistのID ULID
name | varchar(191) | | プレイリスト名
user_acount | varchar(191) | | プレイリストを作成したユーザー
url_string | varchar(191) | | プレイリストURL用の識別子
is_public | boolean | | 公開中かどうか
created_at | timestamp | | プレイリストを作成した日時
updated_at | timestamp | | プレイリストを最終更新した日時

### playlist_song

name | type | opts | note
--- | --- | --- | ---
playlist_id | bigint | PRIMARY KEY |
sort_order | int | PRIMARY KEY |
song_id | bigint | | 楽曲ID

### playlist_favorite

name | type | opts | note
--- | --- | --- | ---
id | bigint | AUTO_INCREMENT |
playlist_id | bigint | | 対象のプレイリストのID
favorite_user_account | string | | プレイリストをふぁぼしたユーザー
created_at | timestamp | | favした日時
