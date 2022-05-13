## どの画面でどのAPIを叩いているか表!


API                               | top | login | signup | mypage | playlist | playlist_edit | 備考
----------------------------------|-----|-------|--------|--------|----------|---------------|------
POST /api/signup                  |     |       |   v    |        |          |               |  
POST /api/login                   |     |   v   |        |        |          |               |  
POST /api/logout                  |  x  |   x   |   x    |   x    |     x    |      x        |  共通ヘッダでログイン状態の時
GET  /api/recent_playlists        |  v  |       |        |        |          |               |  
GET  /api/popular_playlists       |  v  |       |        |        |          |               |  
GET  /api/playlists               |     |       |        |   v    |          |               |  
POST /api/playlist/add            |     |       |        |   v    |          |               |  
GET  /api/playlist/:ulid          |     |       |        |        |     v    |      v        |  
POST /api/playlist/:ulid/update   |     |       |        |        |          |      v        |  
POST /api/playlist/:ulid/delete   |     |       |        |   v    |          |               |  
POST /api/playlist/:ulid/favorite |  v  |       |        |   v    |     v    |               |  
POST /api/admin/user/ban          |     |       |        |        |          |               |  画面なし
