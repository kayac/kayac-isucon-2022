import express, { Request, Response, NextFunction } from 'express'
import partials from 'express-partials'
import session from 'express-session'
import bcrypt from 'bcrypt'
import util from 'util'
import { ulid } from 'ulid'
import mysql, { RowDataPacket, QueryError } from 'mysql2/promise'
const mysqlSession = require('express-mysql-session')(session)
const anonUserAccount = '__'

import {
  UserRow, SongRow, ArtistRow,
  PlaylistRow, PlaylistSongRow, PlaylistFavoriteRow,
} from './types/db'

import {
  Playlist, PlaylistDetail, Song,
  SignupRequest, LoginRequest, AdminPlayerBanRequest,
  AddPlaylistRequest, UpdatePlaylistRequest, FavoritePlaylistRequest,
  BasicResponse, SinglePlaylistResponse, AdminPlayerBanResponse,
  GetRecentPlaylistsResponse, GetPlaylistsResponse, AddPlaylistResponse,
} from './types/api'

declare module 'express-session' {
  interface SessionData {
    user_account: string
  }

  interface Store {
    getAsync(str: string): Promise<SessionData | null | undefined>
  }
}

const sessionCookieName = 'listen80_session'
const publicPath = './public'
const dbConfig = {
  host: process.env['ISUCON_DB_HOST'] ?? '127.0.0.1',
  port: Number(process.env['ISUCON_DB_PORT'] ?? 3306),
  user: process.env['ISUCON_DB_USER'] ?? 'isucon',
  password: process.env['ISUCON_DB_PASSWORD'] ?? 'isucon',
  database: process.env['ISUCON_DB_NAME'] ?? 'isucon_listen80',
}

const pool = mysql.createPool(dbConfig)
const sessionStore: session.Store = new mysqlSession({}, pool)

const app = express()
app.use('/assets', express.static(publicPath + '/assets'))
app.use(express.json())
app.use(session({
  name: sessionCookieName,
  secret: 'powawa',
  store: sessionStore,
  resave: false,
  saveUninitialized: false,
}))
app.use(partials())
app.use((_req: Request, res: Response, next: NextFunction) => {
  res.set('Cache-Control', 'private')
  next()
})

app.set('view engine', 'ejs')
app.set('views', './src/views')
app.set('etag', false)

// error wrapper
function error(req: Request, res: Response, code: number, message: string) {
  console.log(`${req.method} ${req.path} ${code} error: ${message}`)
  const body: BasicResponse = {
    result: false,
    status: code,
    error: message,
  }

  if (code === 401) {
    req.session.destroy(() => {
      res.clearCookie(sessionCookieName)
      res.status(code).json(body)
    })
    return
  }

  res.status(code).json(body)
}

async function validateSession(req: Request): Promise<{ valid: boolean, user?: UserRow }> {
  if (!req.session || !req.session.user_account) {
    return {
      valid: false,
    }
  }

  // session storeの確認
  sessionStore.getAsync = util.promisify(sessionStore.get)
  const session = await sessionStore.getAsync(req.session.id)
  if (!session || session.user_account !== req.session.user_account) {
    return {
      valid: false,
    }
  }

  // BAN statusの確認
  const [[user]] = await pool.query<UserRow[]>(
    'SELECT * FROM user where `account` = ?',
    [req.session.user_account],
  )
  if (!user || user.is_ban) {
    return {
      valid: false,
    }
  }

  return {
    valid: true,
    user,
  }
}

function generatePasswordHash(password: string): string {
  const round = 4
  return bcrypt.hashSync(password, round)
}

function comparePasswordHash(newPassword: string, passwordHash: string): boolean {
  return bcrypt.compareSync(newPassword, passwordHash)
}

// 認証必須ページ
const authRequiredPages = [
  { path: '/mypage', view: 'mypage' },
  { path: '/playlist/:ulid/edit', view: 'playlist_edit' },
]
authRequiredPages.forEach(page => {
  app.get(page.path, async (req: Request, res: Response) => {
    // check login state
    const { valid, user } = await validateSession(req)
    if (!valid || !user) {
      res.redirect('/')
      return
    }

    res.render(page.view + '.ejs', {
      loggedIn: true,
      params: req.params,
      displayName: user.display_name,
      userAccount: user.account,
    })
  })
})

// 認証不要ページ(ログインしている場合はヘッダを変える)
const authOptionalPages = [
  { path: '/', view: 'top' },
  { path: '/playlist/:ulid', view: 'playlist' },
]
authOptionalPages.forEach(page => {
  app.get(page.path, async (req: Request, res: Response) => {
    const { valid, user } = await validateSession(req)
    if (user && user.is_ban) {
      return error(req, res, 401, 'failed to fetch user (no such user)')
    }

    res.render(page.view + '.ejs', {
      loggedIn: valid,
      params: req.params,
      displayName: user ? user.display_name : '',
      userAccount: user ? user.account : '',
    })
  })
})

// 認証関連ページ
const authPages = [
  { path: '/signup', view: 'signup' },
  { path: '/login', view: 'login' },
]
authPages.forEach(page => {
  app.get(page.path, async (req: Request, res: Response) => {
    res.render(page.view + '.ejs', {
      loggedIn: false,
    })
  })
})

// DBにアクセスして結果を引いてくる関数
async function getPlaylistByUlid(db: mysql.Connection, playlistUlid: string): Promise<PlaylistRow | undefined> {
  const [[row]] = await db.query<PlaylistRow[]>(
    'SELECT * FROM playlist WHERE `ulid` = ?',
    [playlistUlid],
  )
  return row
}

async function getPlaylistById(db: mysql.Connection, playlistId: number): Promise<PlaylistRow | undefined> {
  const [[row]] = await db.query<PlaylistRow[]>(
    'SELECT * FROM playlist WHERE `id` = ?',
    [playlistId],
  )
  return row
}

async function getSongByUlid(db: mysql.Connection, songUlid: string): Promise<SongRow | undefined> {
  const [[result]] = await db.query<SongRow[]>(
    'SELECT * FROM song WHERE `ulid` = ?',
    [songUlid],
  )
  return result
}

async function isFavoritedBy(db: mysql.Connection, userAccount: string, playlistId: number): Promise<boolean> {
  const [[row]] = await db.query<RowDataPacket[]>(
    'SELECT COUNT(*) AS cnt FROM playlist_favorite where favorite_user_account = ? AND playlist_id = ?',
    [userAccount, playlistId],
  )
  return row.cnt > 0
}

async function getFavoritesCountByPlaylistId(db: mysql.Connection, playlistId: number): Promise<number> {
  const [[row]] = await db.query<RowDataPacket[]>(
    'SELECT COUNT(*) AS cnt FROM playlist_favorite where playlist_id = ?',
    [playlistId],
  )
  return row.cnt
}

async function getSongsCountByPlaylistId(db: mysql.Connection, playlistId: number): Promise<number> {
  const [[row]] = await db.query<RowDataPacket[]>(
    'SELECT COUNT(*) AS cnt FROM playlist_song where playlist_id = ?',
    [playlistId],
  )
  return row.cnt
}

async function getRecentPlaylistSummaries(db: mysql.Connection, userAccount: string): Promise<Playlist[]> {
  const [allPlaylists] = await db.query<PlaylistRow[]>(
    'SELECT * FROM playlist where is_public = ? ORDER BY created_at DESC',
    [true],
  )
  if (!allPlaylists.length) return []

  const playlists: Playlist[] = []
  for (const playlist of allPlaylists) {
    const user = await getUserByAccount(db, playlist.user_account)
    if (!user || user.is_ban) {
      // banされていたら除外する
      continue
    }

    const songCount = await getSongsCountByPlaylistId(db, playlist.id)
    const favoriteCount = await getFavoritesCountByPlaylistId(db, playlist.id)

    let isFavorited: boolean = false
    if (userAccount != anonUserAccount) {
      // 認証済みの場合はfavを取得
      isFavorited = await isFavoritedBy(db, userAccount, playlist.id)
    }

    playlists.push({
      ulid: playlist.ulid,
      name: playlist.name,
      user_display_name: user.display_name,
      user_account: user.account,
      song_count: songCount,
      favorite_count: favoriteCount,
      is_favorited: isFavorited,
      is_public: !!playlist.is_public,
      created_at: playlist.created_at,
      updated_at: playlist.updated_at
    })
    if (playlists.length >= 100) {
      break
    }
  }
  return playlists
}

async function getPopularPlaylistSummaries(db: mysql.Connection, userAccount: string): Promise<Playlist[]> {
  const [popular] = await db.query<PlaylistFavoriteRow[]>(
    `SELECT playlist_id, count(*) AS favorite_count FROM playlist_favorite GROUP BY playlist_id ORDER BY count(*) DESC`,
  )
  if (!popular.length) return []

  const playlists: Playlist[] = []
  for (const row of popular) {
    const playlist = await getPlaylistById(db, row.playlist_id)
    // 非公開のものは除外する
    if (!playlist || !playlist.is_public) continue

    const user = await getUserByAccount(db, playlist.user_account)
    if (!user || user.is_ban) {
      // banされていたら除外する
      continue
    }

    const songCount = await getSongsCountByPlaylistId(db, playlist.id)
    const favoriteCount = await getFavoritesCountByPlaylistId(db, playlist.id)

    let isFavorited: boolean = false
    if (userAccount != anonUserAccount) {
      // 認証済みの場合はfavを取得
      isFavorited = await isFavoritedBy(db, userAccount, playlist.id)
    }

    playlists.push({
      ulid: playlist.ulid,
      name: playlist.name,
      user_display_name: user.display_name,
      user_account: user.account,
      song_count: songCount,
      favorite_count: favoriteCount,
      is_favorited: isFavorited,
      is_public: !!playlist.is_public,
      created_at: playlist.created_at,
      updated_at: playlist.updated_at
    })
    if (playlists.length >= 100) {
      break
    }
  }
  return playlists
}

async function getCreatedPlaylistSummariesByUserAccount(db: mysql.Connection, userAccount: string): Promise<Playlist[]> {
  const [playlists] = await db.query<PlaylistRow[]>(
    'SELECT * FROM playlist where user_account = ? ORDER BY created_at DESC LIMIT 100',
    [userAccount],
  )
  if (!playlists.length) return []

  const user = await getUserByAccount(db, userAccount)
  if (!user || user.is_ban) return []

  return await Promise.all(playlists.map(async (row: PlaylistRow) => {
    const songCount = await getSongsCountByPlaylistId(db, row.id)
    const favoriteCount = await getFavoritesCountByPlaylistId(db, row.id)
    const isFavorited = await isFavoritedBy(db, userAccount, row.id)

    return {
      ulid: row.ulid,
      name: row.name,
      user_display_name: user.display_name,
      user_account: user.account,
      song_count: songCount,
      favorite_count: favoriteCount,
      is_favorited: isFavorited,
      is_public: !!row.is_public,
      created_at: row.created_at,
      updated_at: row.updated_at
    }
  }))
}

async function getFavoritedPlaylistSummariesByUserAccount(db: mysql.Connection, userAccount: string): Promise<Playlist[]> {
  const [playlistFavorites] = await db.query<PlaylistFavoriteRow[]>(
    'SELECT * FROM playlist_favorite where favorite_user_account = ? ORDER BY created_at DESC',
    [userAccount],
  )
  const playlists: Playlist[] = []
  for (const fav of playlistFavorites) {
    const playlist = await getPlaylistById(db, fav.playlist_id)
    // 非公開は除外する
    if (!playlist || !playlist.is_public) continue

    const user = await getUserByAccount(db, playlist.user_account)
    // 作成したユーザーがbanされていたら除外する
    if (!user || user.is_ban) continue

    const songCount = await getSongsCountByPlaylistId(db, playlist.id)
    const favoriteCount = await getFavoritesCountByPlaylistId(db, playlist.id)
    const isFavorited = await isFavoritedBy(db, userAccount, playlist.id)

    playlists.push({
      ulid: playlist.ulid,
      name: playlist.name,
      user_display_name: user.display_name,
      user_account: user.account,
      song_count: songCount,
      favorite_count: favoriteCount,
      is_favorited: isFavorited,
      is_public: !!playlist.is_public,
      created_at: playlist.created_at,
      updated_at: playlist.updated_at
    })
    if (playlists.length >= 100) break
  }
  return playlists
}

async function getPlaylistDetailByPlaylistUlid(db: mysql.Connection, playlistUlid: string, viewerUserAccount: (string | undefined)): Promise<PlaylistDetail | undefined> {
  const playlist = await getPlaylistByUlid(db, playlistUlid)
  if (!playlist) return

  const user = await getUserByAccount(db, playlist.user_account)
  if (!user || user.is_ban) return

  const favoriteCount = await getFavoritesCountByPlaylistId(db, playlist.id)
  let isFavorited: boolean = false
  if (viewerUserAccount) {
    isFavorited = await isFavoritedBy(db, viewerUserAccount, playlist.id)
  }

  const resPlaylistSongs = await db.query<PlaylistSongRow[]>(
    'SELECT * FROM playlist_song WHERE playlist_id = ?',
    [playlist.id],
  )
  const [playlistSongRows] = resPlaylistSongs

  const songs: Song[] = await Promise.all(playlistSongRows.map(async (row: PlaylistSongRow): Promise<Song> => {
    const [[song]] = await db.query<SongRow[]>(
      'SELECT * FROM song WHERE id = ?',
      [row.song_id],
    )

    const [[artist]] = await db.query<ArtistRow[]>(
      'SELECT * FROM artist WHERE id = ?',
      [song.artist_id],
    )

    return {
      ulid: song.ulid,
      title: song.title,
      artist: artist.name,
      album: song.album,
      track_number: song.track_number,
      is_public: !!song.is_public,
    }
  }))

  return {
    ulid: playlist.ulid,
    name: playlist.name,
    user_display_name: user.display_name,
    user_account: user.account,
    song_count: songs.length,
    favorite_count: favoriteCount,
    is_favorited: isFavorited,
    is_public: !!playlist.is_public,
    songs: songs,
    created_at: playlist.created_at,
    updated_at: playlist.updated_at
  }
}

async function getPlaylistFavoritesByPlaylistIdAndUserAccount(db: mysql.Connection, playlistId: number, favoriteUserAccount: (string | undefined)): Promise<PlaylistFavoriteRow | undefined> {
  const [result] = await db.query<PlaylistFavoriteRow[]>(
    'SELECT * FROM playlist_favorite WHERE `playlist_id` = ? AND `favorite_user_account` = ?',
    [playlistId, favoriteUserAccount],
  )
  if (!result.length) return

  return result[0]
}

async function getUserByAccount(db: mysql.Connection, account: string): Promise<UserRow | undefined> {
  const [result] = await db.query<UserRow[]>(
    'SELECT * FROM user WHERE `account` = ?',
    [account],
  )
  if (!result.length) return

  return result[0]
}

async function insertPlaylistSong(db: mysql.Connection, arg: { playlistId: number, sortOrder: number, songId: number }) {
  await db.query(
    'INSERT INTO playlist_song (`playlist_id`, `sort_order`, `song_id`) VALUES (?, ?, ?)',
    [arg.playlistId, arg.sortOrder, arg.songId],
  )
}

async function insertPlaylistFavorite(db: mysql.Connection, arg: { playlistId: number, favoriteUserAccount: string, createdAt: Date }) {
  await db.query(
    'INSERT INTO playlist_favorite (`playlist_id`, `favorite_user_account`, `created_at`) VALUES (?, ?, ?)',
    [arg.playlistId, arg.favoriteUserAccount, arg.createdAt],
  )
}



// POST /api/signup
app.post('/api/signup', async (req: Request, res: Response) => {
  const { user_account, password, display_name } = req.body as SignupRequest

  // validation
  if (!user_account || user_account.length < 4 || 191 < user_account.length || user_account.match(/[^a-zA-Z0-9\-_]/) !== null) {
    return error(req, res, 400, 'bad user_account')
  }
  if (!password || password.length < 8 || 64 < password.length || password.match(/[^a-zA-Z0-9\-_]/) != null) {
    return error(req, res, 400, 'bad password')
  }
  if (!display_name || display_name.length < 2 || 24 < display_name.length) {
    return error(req, res, 400, 'bad display_name')
  }

  // password hashを作る
  const passwordHash = generatePasswordHash(password)

  // default value
  const is_ban = false
  const signupTimestamp = new Date

  const db = await pool.getConnection()
  try {
    const displayName = display_name ? display_name : user_account
    await db.query(
      'INSERT INTO user (`account`, `display_name`, `password_hash`, `is_ban`, `created_at`, `last_logined_at`) VALUES (?, ?, ?, ?, ?, ?)',
      [user_account, displayName, passwordHash, is_ban, signupTimestamp, signupTimestamp],
    )

    req.session.regenerate((_err) => {
      req.session.user_account = user_account
      const body: BasicResponse = {
        result: true,
        status: 200,
      }
      res.status(body.status).json(body)
    })

  } catch (err) {
    if ((err as QueryError).code === 'ER_DUP_ENTRY') {
      return error(req, res, 400, 'account already exist')
    }

    console.log(err)
    error(req, res, 500, 'failed to signup')
  } finally {
    db.release()
  }
})

// POST /api/login
app.post('/api/login', async (req: Request, res: Response) => {
  const { user_account, password } = req.body as LoginRequest

  // validation
  if (!user_account || user_account.length < 4 || 191 < user_account.length || user_account.match(/[^a-zA-Z0-9\-_]/) !== null) {
    return error(req, res, 400, 'bad user_account')
  }
  if (!password || password.length < 8 || 64 < password.length || password.match(/[^a-zA-Z0-9\-_]/) != null) {
    return error(req, res, 400, 'bad password')
  }

  // password check
  const db = await pool.getConnection()
  try {
    const user = await getUserByAccount(db, user_account)
    if (!user || user.is_ban) {
      // ユーザがいないかbanされている
      return error(req, res, 401, 'failed to login (no such user)')
    }

    if (!comparePasswordHash(password, user.password_hash)) {
      // wrong password
      return error(req, res, 401, 'failed to login (wrong password)')
    }

    // 最終ログイン日時を更新
    await db.query('UPDATE user SET last_logined_at = ? WHERE account = ?', [new Date, user.account])

    req.session.regenerate((_err) => {
      req.session.user_account = user.account
      const body: BasicResponse = {
        result: true,
        status: 200,
      }
      res.status(body.status).json(body)
    })

  } catch (err) {
    console.log(err)
    error(req, res, 500, 'failed to login (server error)')
  } finally {
    db.release()
  }
})

// POST /api/logout
app.post('/api/logout', async (req: Request, res: Response) => {
  req.session.destroy(() => {
    res.clearCookie(sessionCookieName)
    const body: BasicResponse = {
      result: true,
      status: 200,
    }
    res.status(body.status).json(body)
  })
})

// GET /api/recent_playlists
app.get('/api/recent_playlists', async (req: Request, res: Response) => {
  const user_account = req.session.user_account ?? anonUserAccount

  const db = await pool.getConnection()
  try {
    const playlists = await getRecentPlaylistSummaries(db, user_account)

    const body: GetRecentPlaylistsResponse = {
      result: true,
      status: 200,
      playlists: playlists,
    }
    res.status(body.status).json(body)
  } catch (err) {
    console.log(err)
    error(req, res, 500, 'internal server error')
  } finally {
    db.release()
  }
})

// GET /api/popular_playlists
app.get('/api/popular_playlists', async (req: Request, res: Response) => {
  const user_account = req.session.user_account ?? anonUserAccount

  const db = await pool.getConnection()
  try {
    // トランザクションを使わないとfav数の順番が狂うことがある
    await db.beginTransaction()
    const playlists = await getPopularPlaylistSummaries(db, user_account)

    const body: GetRecentPlaylistsResponse = {
      result: true,
      status: 200,
      playlists: playlists,
    }
    res.status(body.status).json(body)
    await db.commit()
  } catch (err) {
    await db.rollback()
    console.log(err)
    error(req, res, 500, 'internal server error')
  } finally {
    db.release()
  }
})

// GET /api/playlists
app.get('/api/playlists', async (req: Request, res: Response) => {
  const { valid } = await validateSession(req)
  if (!valid) {
    return error(req, res, 401, 'login required')
  }
  const user_account = req.session.user_account ?? anonUserAccount

  const db = await pool.getConnection()
  try {
    const createdPlaylists = await getCreatedPlaylistSummariesByUserAccount(db, user_account)
    const favoritedPlaylists = await getFavoritedPlaylistSummariesByUserAccount(db, user_account)

    const body: GetPlaylistsResponse = {
      result: true,
      status: 200,
      created_playlists: createdPlaylists,
      favorited_playlists: favoritedPlaylists,
    }
    res.status(body.status).json(body)
  } catch (err) {
    console.log(err)
    error(req, res, 500, 'internal server error')
  } finally {
    db.release()
  }
})

// GET /api/playlist/{:playlistUlid}
app.get('/api/playlist/:playlistUlid', async (req: Request, res: Response) => {
  // ログインは不要
  const userAccount: string = req.session.user_account || anonUserAccount
  const playlistUlid: string = req.params.playlistUlid

  // validation
  if (!playlistUlid || playlistUlid.match(/[^a-zA-Z0-9]/) !== null) {
    return error(req, res, 400, 'bad playlist ulid')
  }

  const db = await pool.getConnection()
  try {
    const playlist = await getPlaylistByUlid(db, playlistUlid)
    if (!playlist) {
      return error(req, res, 404, 'playlist not found')
    }

    // 作成者が自分ではない、privateなプレイリストは見れない
    if (playlist.user_account != userAccount && !playlist.is_public) {
      return error(req, res, 404, 'playlist not found')
    }

    const playlistDetails = await getPlaylistDetailByPlaylistUlid(db, playlist.ulid, userAccount)
    if (!playlistDetails) {
      return error(req, res, 404, 'playlist not found')
    }

    const body: SinglePlaylistResponse = {
      result: true,
      status: 200,
      playlist: playlistDetails,
    }
    res.status(body.status).json(body)
  } catch (err) {
    console.log(err)
    error(req, res, 500, 'internal server error')
  } finally {
    db.release()
  }
})

// POST /api/playlist/add
app.post('/api/playlist/add', async (req: Request, res: Response) => {
  const { valid } = await validateSession(req)
  if (!valid) {
    return error(req, res, 401, 'login required')
  }

  const { name } = req.body as AddPlaylistRequest
  // validation
  if (!name || name.length < 2 || 191 < name.length) {
    return error(req, res, 400, 'invalid name')
  }

  const userAccount = req.session.user_account ?? anonUserAccount
  const createdTimestamp = new Date
  const playlist_ulid = ulid(createdTimestamp.getTime())

  const db = await pool.getConnection()
  try {
    await db.query(
      'INSERT INTO playlist (`ulid`, `name`, `user_account`, `is_public`, `created_at`, `updated_at`) VALUES (?, ?, ?, ?, ?, ?)',
      [playlist_ulid, name, userAccount, false, createdTimestamp, createdTimestamp], // 作成時は非公開
    )

    const body: AddPlaylistResponse = {
      result: true,
      status: 200,
      playlist_ulid: playlist_ulid,
    }
    res.status(body.status).json(body)
  } catch (err) {
    console.log(err)
    error(req, res, 500, 'internal server error')
  } finally {
    db.release()
  }
})

// POST /api/playlist/update
app.post('/api/playlist/:playlistUlid/update', async (req: Request, res: Response) => {
  const { valid } = await validateSession(req)
  if (!valid) {
    return error(req, res, 401, 'login required')
  }
  const userAccount = req.session.user_account

  const db = await pool.getConnection()
  try {
    const playlistUlid: string = req.params.playlistUlid
    const playlist = await getPlaylistByUlid(db, playlistUlid)
    if (!playlist) {
      return error(req, res, 404, 'playlist not found')
    }
    if (playlist.user_account != userAccount) {
      // 権限エラーだが、URI上のパラメータが不正なので404を返す
      return error(req, res, 404, 'playlist not found')
    }

    const { name, song_ulids, is_public } = req.body as UpdatePlaylistRequest
    // validation
    if (!playlistUlid || playlistUlid.match(/[^a-zA-Z0-9]/) !== null) {
      return error(req, res, 404, 'bad playlist ulid')
    }
    // 3つの必須パラメータをチェック
    if (!name || !song_ulids || is_public === undefined) {
      return error(req, res, 400, 'name, song_ulids and is_public is required')
    }
    // nameは2文字以上191文字以内
    if (name.length < 2 || 191 < name.length) {
      return error(req, res, 400, 'invalid name')
    }
    // 曲数は最大80曲
    if (80 < song_ulids.length) {
      return error(req, res, 400, 'invalid song_ulids')
    }
    // 曲は重複してはいけない
    const songUlidsSet = new Set(song_ulids)
    if (songUlidsSet.size != song_ulids.length) {
      return error(req, res, 400, 'invalid song_ulids')
    }

    const updatedTimestamp = new Date

    await db.beginTransaction()

    // name, is_publicの更新
    await db.query(
      'UPDATE playlist SET name = ?, is_public = ?, `updated_at` = ? WHERE `ulid` = ?',
      [name, is_public, updatedTimestamp, playlist.ulid],
    )

    // songsを削除→新しいものを入れる
    await db.query(
      'DELETE FROM playlist_song WHERE playlist_id = ?',
      [playlist.id],
    )

    for (const [index, songUlid] of song_ulids.entries()) {
      const song = await getSongByUlid(db, songUlid)
      if (!song) {
        await db.rollback()
        return error(req, res, 400, `song not found. ulid: ${songUlid}`)
      }

      // songSortOrderは 0 based、保存するsort_orderは 1 based なので+1
      await insertPlaylistSong(db, {
        playlistId: playlist.id,
        sortOrder: index + 1,
        songId: song.id,
      })
    }

    await db.commit()

    const playlistDetails = await getPlaylistDetailByPlaylistUlid(db, playlist.ulid, userAccount)
    if (!playlistDetails) {
      return error(req, res, 500, 'error occurred: getPlaylistDetailByPlaylistUlid')
    }

    const body: SinglePlaylistResponse = {
      result: true,
      status: 200,
      playlist: playlistDetails,
    }
    res.status(body.status).json(body)
  } catch (err) {
    await db.rollback()
    console.log(err)
    error(req, res, 500, 'internal server error')
  } finally {
    db.release()
  }
})

// POST /api/playlist/delete
app.post('/api/playlist/:playlistUlid/delete', async (req: Request, res: Response) => {
  const { valid } = await validateSession(req)
  if (!valid) {
    return error(req, res, 401, 'login required')
  }
  const playlistUlid: string = req.params.playlistUlid
  // validation
  if (!playlistUlid || playlistUlid.match(/[^a-zA-Z0-9]/) !== null) {
    return error(req, res, 404, 'bad playlist ulid')
  }

  const db = await pool.getConnection()
  try {
    const playlist = await getPlaylistByUlid(db, playlistUlid)
    if (!playlist) {
      return error(req, res, 400, 'playlist not found')
    }

    if (playlist.user_account !== req.session.user_account) {
      return error(req, res, 400, "do not delete other users playlist")
    }

    await db.query('DELETE FROM playlist WHERE `ulid` = ?', [playlist.ulid])
    await db.query('DELETE FROM playlist_song WHERE playlist_id = ?', [playlist.id])
    await db.query('DELETE FROM playlist_favorite WHERE playlist_id = ?', [playlist.id])

    const body: BasicResponse = {
      result: true,
      status: 200,
    }
    res.status(200).json(body)
  } catch (err) {
    console.log(err)
    error(req, res, 500, 'internal server error')
  } finally {
    db.release()
  }
})

// POST /api/playlist/:ulid/favorite
app.post('/api/playlist/:playlistUlid/favorite', async (req: Request, res: Response) => {
  const { valid, user } = await validateSession(req)
  if (!valid || !user || !req.session.user_account) {
    return error(req, res, 401, 'login required')
  }
  const playlistUlid: string = req.params.playlistUlid
  const { is_favorited } = req.body as FavoritePlaylistRequest
  if (!playlistUlid || playlistUlid.match(/[^a-zA-Z0-9]/) !== null) {
    return error(req, res, 404, 'bad playlist ulid')
  }

  const db = await pool.getConnection()
  try {
    const playlist = await getPlaylistByUlid(db, playlistUlid)
    if (!playlist) {
      return error(req, res, 404, 'playlist not found')
    }
    // 操作対象のプレイリストが他のユーザーの場合、banされているかプレイリストがprivateならばnot found
    if (playlist.user_account !== user.account) {
      if (!user || user.is_ban || !playlist.is_public) {
        return error(req, res, 404, 'playlist not found')
      }
    }

    if (is_favorited) {
      // insert
      const createdTimestamp = new Date
      const playlistFavorite = await getPlaylistFavoritesByPlaylistIdAndUserAccount(db, playlist.id, req.session.user_account)
      if (!playlistFavorite) {
        await insertPlaylistFavorite(db, {
          playlistId: playlist.id,
          favoriteUserAccount: req.session.user_account,
          createdAt: createdTimestamp,
        })
      }
    } else {
      // delete
      await db.query(
        'DELETE FROM playlist_favorite WHERE `playlist_id` = ? AND `favorite_user_account` = ?',
        [playlist.id, req.session.user_account],
      )
    }

    const playlistDetail = await getPlaylistDetailByPlaylistUlid(db, playlist.ulid, req.session.user_account)
    if (!playlistDetail) {
      return error(req, res, 404, 'failed to fetch playlist detail')
    }

    const body: SinglePlaylistResponse = {
      result: true,
      status: 200,
      playlist: playlistDetail,
    }
    res.status(body.status).json(body)
  } catch (err) {
    console.log(err)
    error(req, res, 500, 'internal server error')
  } finally {
    db.release()
  }
})

// POST /api/admin/user/ban
app.post('/api/admin/user/ban', async (req: Request, res: Response) => {
  const { valid, user } = await validateSession(req)
  if (!valid || !user) {
    return error(req, res, 401, 'login required')
  }

  // 管理者userであることを確認,でなければ403
  if (!isAdminUser(user.account)) {
    return error(req, res, 403, 'not admin user')
  }

  const { user_account, is_ban } = req.body as AdminPlayerBanRequest

  const db = await pool.getConnection()
  try {
    await db.query('UPDATE user SET `is_ban` = ?  WHERE `account` = ?', [is_ban, user_account])
    const user = await getUserByAccount(db, user_account)
    if (!user) {
      return error(req, res, 400, 'user not found')
    }

    const body: AdminPlayerBanResponse = {
      result: true,
      status: 200,
      user_account: user.account,
      display_name: user.display_name,
      is_ban: !!user.is_ban,
      created_at: user.created_at,
    }
    res.status(body.status).json(body)
  } catch (err) {
    console.log(err)
    error(req, res, 500, 'internal server error')
  } finally {
    db.release()
  }
})

function isAdminUser(account: string): boolean {
  // ひとまず一人決め打ち、後に条件増やすかも
  if (account === "adminuser") {
    return true
  }
  return false
}

// 競技に必要なAPI
// DBの初期化処理
const lastCreatedAt: string = '2022-05-13 09:00:00.000'

app.post('/initialize', async (req: Request, res: Response) => {
  const db = await pool.getConnection()
  try {
    await db.query(
      'DELETE FROM user WHERE ? < created_at',
      [lastCreatedAt]
    )
    await db.query(
      'DELETE FROM playlist WHERE ? < created_at OR user_account NOT IN (SELECT account FROM user)',
      [lastCreatedAt]
    )
    await db.query(
      'DELETE FROM playlist_song WHERE playlist_id NOT IN (SELECT id FROM playlist)',
    )
    await db.query(
      'DELETE FROM playlist_favorite WHERE playlist_id NOT IN (SELECT id FROM playlist) OR ? < created_at',
      [lastCreatedAt]
    )
    const body: BasicResponse = {
      result: true,
      status: 200,
    }
    res.status(body.status).json(body)
  } catch {
    error(req, res, 500, 'internal server error')
  }
})

const port = parseInt(process.env['SERVER_APP_PORT'] ?? '3000', 10)
console.log('starting listen80 server on :' + port + ' ...')
app.listen(port)
