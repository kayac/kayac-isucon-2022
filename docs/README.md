# 社内 ISUCON レギュレーション & 当日マニュアル

## スケジュール

- 9:30 競技説明開始
- 10:00 競技開始
- 17:00 競技終了
- 18:00 結果発表・講評

## サーバ事項

参加者は主催者が用意したAmazon Web Services（以下AWS）のEC2インスタンスを1台利用する。

参加登録時に申請したGitHubアカウントに登録されている公開鍵 (`github.com/{username}.keys` で配布されているもの) を用いて、`isucon` ユーザーで ssh 接続が可能。

## ソフトウェア事項

コンテストにあたり、参加者は与えられたソフトウェア、もしくは自分で競技時間内に実装したソフトウェアを用いる。

高速化対象のソフトウェアとして、主催者からTypeScriptとGoによるWebアプリケーションが与えられる。独自で実装したものを用いてもよい。

競技における高速化対象のアプリケーションとして与えられたアプリケーションから、以下の機能は変更しないこと。

  * アクセス先のURI（ポート、およびHTTPリクエストパス）
    * ただしサーバー側で生成する部分（ID）は文字種（[0-9] や [0-9a-zA-Z_] など）を変えない範囲で自由に生成しても良い
  * APIレスポンス(JSON)の構造
  * JavaScript/CSSファイルの内容
  * 画像および動画等のメディアファイルの内容
  * パスワードハッシュアルゴリズム、およびそのラウンド数

各サーバにおけるソフトウェアの入れ替え、設定の変更、アプリケーションコードの変更および入れ替えなどは一切禁止しない。起動したインスタンス以外の外部リソースを利用する行為（他のインスタンスに処理を委譲するなど）は禁止する。
ただしモニタリングやテスト、開発などにおいては、PCや外部のサーバーを利用しても構わない。

許可される事項には、例として以下のような作業が含まれる。

  * DBスキーマの変更やインデックスの作成・削除
  * キャッシュ機構の追加、jobqueue機構の追加による遅延書き込み
  * 利用するミドルウェアの変更
  * 他の言語による再実装

ただし以下の事項に留意すること。

  * コンテスト進行用のメンテナンスコマンドが正常に動作するよう互換性を保つこと
    * EC2上で動作している ssm-agent を停止しないこと
  * サーバ再起動後にすべてのアプリケーションコードが正常動作する状態を維持すること
  * ベンチマーク実行時にアプリケーションに書き込まれたデータは再起動後にも取得できること

## 禁止事項

以下の事項は特別に禁止する。

  * 他のチームへの妨害と主催者がみなす全ての行為

## 採点

採点は採点条件（後述）をクリアした参加者の間で、性能値（後述）の高さを競うものとする。

採点条件として、以下の各チェックの検査を通過するものとする。

  * 負荷走行中、ログイン中のユーザーがPOSTした内容は、POSTへのHTTPレスポンスが返却された後に同一ユーザーがリクエストしたレスポンス内容に即座に反映されていること
    * 例外がある場合は [API仕様書](API.md) に記述されている
  * ログイン中のユーザーに対する API(URLのパス `/api/*`)のレスポンスについては、HTTPヘッダで `Cache-Control: private` を付与すること
  * APIレスポンスのJSON構造が変化していないこと
  * 各APIリクエストは15秒以内にレスポンスを返却する必要がある
    * 例外として、ベンチマーカーが開始時に1回だけ送信する初期化リクエスト `POST /initialize` については30秒まで許容する
  * ブラウザから対象アプリケーションにアクセスした結果、ページ上の表示および各種動作が正常であること

性能値として、以下の指標を用いる。

  * 性能値を計測するための計測ツールの実行時間は1分間とする
    * ベンチマーカーが開始時に行う「整合性チェック」が完了するまでの時間は含まない
  * 計測時間内のHTTPリクエスト成功数をベースとする
    * 「整合性チェック」のリクエスト数はスコアとして算入しない
  * リクエストの種類毎に配点を変更する
    * ログイン中のユーザーによるプレイリストの更新 (`POST /api/playlist/{*}/update`) は10点
    * ログイン中のユーザーによる最新プレイリスト一覧取得(`GET /api/recent_playlists`)、人気のプレイリスト一覧取得(`GET /api/popular_playlists`) は10点
    * それ以外のリクエストは全て1点
  * エラーの数により減点する
    * リクエストタイムアウトが発生した場合、レスポンスがアプリケーションとして期待される挙動を示さなかった場合にはエラーとする
    * エラーが1回発生するごとに-10点
    * 負荷走行中にエラーが30回以上発生した場合は、その時点で負荷走行を打ち切ることがある

## アプリケーションについて

このアプリケーションは **Listen80** というプレイリスト共有サービスである。

![](listen80.png)

次の機能が提供されている。

- ログインしない状態で、トップページが閲覧できる
  - 人気のプレイリスト100件と最新プレイリスト100件が閲覧できる
  - プレイリストの詳細が閲覧できる
- サインアップすることで、新規ユーザーを登録できる
- ログインすると、マイページが表示される
  - 自分が作成したプレイリスト最新最大100件と、自分がラブ(♡)を付与したプレイリスト最新最大100件が閲覧できる
- マイページからはプレイリストの作成と編集ができる
  - プレイリストは非公開(自分以外には閲覧できない)状態で作成される
  - 編集して曲を追加したり、公開にしたりできる
- 管理APIがあり、ユーザーをBANできる
  - BANされたユーザーはAPIが利用できなくなる
  - BANされたユーザーが作成したプレイリストは他のユーザーからは見えなくなる
  - 管理APIではユーザーのBANを解除できる

詳細な API 仕様については [API仕様書](API.md) に記述されている。

仕様と初期実装に齟齬がある場合、ベンチマーカーがエラーを検出しない限りはどちらに準拠しても構わない。

テスト用のユーザー

- 一般ユーザー アカウント名 `isucon` パスワード `isuconpass`
- 管理者 アカウント名 `adminuser` パスワード `adminuser`

### アプリケーション動作環境

競技用EC2インスタンスでは、docker compose によってアプリケーションとミドルウェア(nginx, MySQL)が起動している。

`/home/isucon/webapp` で次のコマンドを実行することで、停止、コンテナのビルド、起動などが行える。

```console
$ docker-compose down     # 停止
$ docker-compose up --build    # コンテナビルドをしてから起動(foreground)
$ docker-compose up --build -d # コンテナビルドをしてから起動(daemon)
$ docker-compose logs -f       # コンテナがstdout/stderrに出力したログを閲覧
```

#### アプリケーション

アプリケーションの実装は `webapp/{node,golang}` 以下に配置されている。

初期状態ではnode(TypeScript)実装が起動している。Go実装に切り替えたい場合は `webapp/docker-compose.yml` の `build:` を変更して `docker-compose` でビルドを行うこと。

```yaml
  app:
    cpus: 1
    mem_limit: 1g
    # Go実装の場合は golang/ node実装の場合は node/
    build: node/
    environment:
      ISUCON_DB_HOST: mysql
      ISUCON_DB_PORT: 3306
      ISUCON_DB_USER: isucon
      ISUCON_DB_PASSWORD: isucon
      ISUCON_DB_NAME: isucon_listen80
    links:
      - mysql
    volumes:
      - ./public:/home/isucon/webapp/public
      - gopkg:/usr/local/go/pkg
    init: true
    restart: always
```


#### nginx

docker-compose で起動している nginx は起動時にEC2上の `webapp/nginx/conf.d/default.conf` を読み込む。このファイルを編集することでコンテナをビルドしなくても設定が変更できる。

```yaml
  nginx:
    image: nginx:1.20
    volumes:
      - ./nginx/conf.d:/etc/nginx/conf.d
      - ./public:/public
    ports:
      - "80:80"
    links:
      - app
    restart: always
```

#### MySQL

docker-compose で起動している MySQL は、起動時にEC2上の `webapp/mysql/my.cnf` を読み込む。このファイルを編集することでコンテナをビルドしなくても設定が変更できる。

```yaml
  mysql:
    cpus: 1
    mem_limit: 1g
    image: mysql/mysql-server:8.0.28
    environment:
      - "MYSQL_ROOT_HOST=%"
      - "MYSQL_ROOT_PASSWORD=root"
    volumes:
      - ../sql:/docker-entrypoint-initdb.d
      - mysql:/var/lib/mysql
      - ./mysql/my.cnf:/etc/my.cnf
      - ./mysql/logs:/var/log/mysql
    ports:
      - 13306:3306
    restart: always
```

EC2の TCP 13306 ポートを開いているため、EC2ホスト側から以下のコマンドで接続できる。

```console
$ mysql -uroot -proot --host 127.0.0.1 --port 13306 isucon_listen80
```

### 初期状態へのデータリセット方法

`/home/isucon/sql` 以下に初期化用のSQLファイルがあるので、リセットしたい場合は次のようにしてimportできる。

```console
$ cd /home/isucon
$ mysql -uroot -proot --host 127.0.0.1 --port 13306 isucon_listen80 < sql/90_isucon_listen80_dump.sql
```

```
sql
├── 00_database_user.sql         # ユーザー定義
├── 50_listen80_schema.sql       # アプリケーション用のスキーマ定義
└── 90_isucon_listen80_dump.sql  # 初期データ
```
