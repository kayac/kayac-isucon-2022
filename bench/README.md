# bench

## How to run

前提

- repo root にいる状態
- docker-compose で起動している状態
  - nginx が port 80
  - mysql が port 13306

```console
$ gh release download dummydata/20220421_0056-prod # もしくはreleaseからisucon_listen80_dump_prod.tar.gzをダウンロード
$ tar xvf isucon_listen80_dump_prod.tar.gz
$ mysql -uroot -proot --host 127.0.0.1 --port 13306 < isucon_listen80_dump.sql
$ cd bench
$ make
$ ./bench -target-url http://localhost  # nginxのportを変えている場合はportを合わせる
```
