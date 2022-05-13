window.addEventListener('load', () => {

  const getPlaylist = async () => {
    const playlistUlid = document.getElementById('playlist-id').getAttribute('value')

    const response = await fetch('/api/playlist/' + playlistUlid)
    const json = await response.json()
    if (!json.result) {
      alert('プレイリスト取得失敗 (他人の非公開プレイリストかも?)')
    }

    const playlist = json.playlist
    const createdAt = new Date(Date.parse(playlist.created_at)).toLocaleString()
    const updatedAt = new Date(Date.parse(playlist.updated_at)).toLocaleString()
    const favstat = playlist.is_favorited ? 'fav' : ''
    const visibility = playlist.is_public ? '公開' : '非公開'

    const metadata = document.getElementById('playlist-metadata')

    metadata.innerHTML = `
    
    <h3>${playlist.name}
    <div class="love-button">
      <a class="${favstat} action-button" onclick="favHandler('${playlist.ulid}', ${playlist.is_favorited})"><svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="currentColor" class="bi bi-heart-fill" viewBox="0 0 16 16">
      <path fill-rule="evenodd" d="M8 1.314C12.438-3.248 23.534 4.735 8 15-7.534 4.736 3.562-3.248 8 1.314z"/>
    </svg></a>
    </div></h3>
    <ul>
      <li>作成者: ${playlist.user_display_name}</li>
      <li>曲数: ${playlist.song_count}曲</li>
      <li>集めたラブ: ${playlist.favorite_count}個</li>
      <li>ID: ${playlist.ulid}</li>
      <li>作成日時: ${createdAt}</li>
      <li>更新日時: ${updatedAt}</li>
      <li>公開ステータス: ${visibility}</li>
    </ul>
    `

    const tbody = document.getElementById('songs-table-body')

    const rowHTML = playlist.songs.map((song, index) => {
      return `
      <tr>
        <td>${index+1}</td>
        <td>${song.title}</td>
        <td>${song.artist}</td>
        <td>${song.album}</td>
        <td>${song.track_number}</td>
      </tr>
      `
    })

    tbody.innerHTML = rowHTML.join('\n')

    return response
  }

  const main = document.getElementById('main-app')
  main.addEventListener('refreshRequired', () => {
    getPlaylist()
  })

  getPlaylist()
})
