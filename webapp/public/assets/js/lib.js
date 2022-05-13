const favHandler = async (id, currentValue) => {
  const favorite = async (id, newState) => {
    const response = await fetch(`/api/playlist/${id}/favorite`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        is_favorited: newState,
      })
    })
    const json = await response.json()
    if (!json.result) {
      alert('ラブ失敗 (ログインしてますか?)')
    }

    return response
  }

  const newState = !currentValue

  await favorite(id, newState)

  const event = new Event('refreshRequired')
  const elem = document.getElementById('main-app')
  elem.dispatchEvent(event)
}

const deleteHandler = async (id) => {
  const deletePlaylist = async (id) => {
    const response = await fetch(`/api/playlist/${id}/delete`, {
      method: 'POST',
    })
    const json = await response.json()
    if (!json.result) {
      alert('delete失敗 (ログインしてますか? 自分のですか?)')
    }
    return response
  }

  if (window.confirm('このプレイリストを削除していいですか?')) {
    await deletePlaylist(id)

    const event = new Event('refreshRequired')
    const elem = document.getElementById('main-app')
    elem.dispatchEvent(event)
 }
}

function playlistToHTML(playlist, owner) {
  const createdAt = new Date(Date.parse(playlist.created_at)).toLocaleString()
  const updatedAt = new Date(Date.parse(playlist.updated_at)).toLocaleString()
  const favstat = playlist.is_favorited ? 'fav' : ''
  const visibility = owner ? (
    playlist.is_public ? `<span title="公開プレイリストです"><svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" fill="currentColor" class="bi bi-eye" viewBox="0 0 16 16">
    <path d="M16 8s-3-5.5-8-5.5S0 8 0 8s3 5.5 8 5.5S16 8 16 8zM1.173 8a13.133 13.133 0 0 1 1.66-2.043C4.12 4.668 5.88 3.5 8 3.5c2.12 0 3.879 1.168 5.168 2.457A13.133 13.133 0 0 1 14.828 8c-.058.087-.122.183-.195.288-.335.48-.83 1.12-1.465 1.755C11.879 11.332 10.119 12.5 8 12.5c-2.12 0-3.879-1.168-5.168-2.457A13.134 13.134 0 0 1 1.172 8z"/>
    <path d="M8 5.5a2.5 2.5 0 1 0 0 5 2.5 2.5 0 0 0 0-5zM4.5 8a3.5 3.5 0 1 1 7 0 3.5 3.5 0 0 1-7 0z"/>
  </svg></span>` : `<span title="非公開プレイリストです"><svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" fill="currentColor" class="bi bi-eye-slash" viewBox="0 0 16 16">
    <path d="M13.359 11.238C15.06 9.72 16 8 16 8s-3-5.5-8-5.5a7.028 7.028 0 0 0-2.79.588l.77.771A5.944 5.944 0 0 1 8 3.5c2.12 0 3.879 1.168 5.168 2.457A13.134 13.134 0 0 1 14.828 8c-.058.087-.122.183-.195.288-.335.48-.83 1.12-1.465 1.755-.165.165-.337.328-.517.486l.708.709z"/>
    <path d="M11.297 9.176a3.5 3.5 0 0 0-4.474-4.474l.823.823a2.5 2.5 0 0 1 2.829 2.829l.822.822zm-2.943 1.299.822.822a3.5 3.5 0 0 1-4.474-4.474l.823.823a2.5 2.5 0 0 0 2.829 2.829z"/>
    <path d="M3.35 5.47c-.18.16-.353.322-.518.487A13.134 13.134 0 0 0 1.172 8l.195.288c.335.48.83 1.12 1.465 1.755C4.121 11.332 5.881 12.5 8 12.5c.716 0 1.39-.133 2.02-.36l.77.772A7.029 7.029 0 0 1 8 13.5C3 13.5 0 8 0 8s.939-1.721 2.641-3.238l.708.709zm10.296 8.884-12-12 .708-.708 12 12-.708.708z"/>
  </svg></span>`
  ) : ''
  const editButton = owner ? `<a class="action-button" href="/playlist/${playlist.ulid}/edit">
    <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" fill="currentColor" class="bi bi-pencil-square" viewBox="0 0 16 16">
      <path d="M15.502 1.94a.5.5 0 0 1 0 .706L14.459 3.69l-2-2L13.502.646a.5.5 0 0 1 .707 0l1.293 1.293zm-1.75 2.456-2-2L4.939 9.21a.5.5 0 0 0-.121.196l-.805 2.414a.25.25 0 0 0 .316.316l2.414-.805a.5.5 0 0 0 .196-.12l6.813-6.814z"/>
      <path fill-rule="evenodd" d="M1 13.5A1.5 1.5 0 0 0 2.5 15h11a1.5 1.5 0 0 0 1.5-1.5v-6a.5.5 0 0 0-1 0v6a.5.5 0 0 1-.5.5h-11a.5.5 0 0 1-.5-.5v-11a.5.5 0 0 1 .5-.5H9a.5.5 0 0 0 0-1H2.5A1.5 1.5 0 0 0 1 2.5v11z"/>
    </svg></a>` : ''

  const deleteButton = owner ? `<a class="action-button" onclick="deleteHandler('${playlist.ulid}')">
  <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" fill="currentColor" class="bi bi-trash" viewBox="0 0 16 16">
  <path d="M5.5 5.5A.5.5 0 0 1 6 6v6a.5.5 0 0 1-1 0V6a.5.5 0 0 1 .5-.5zm2.5 0a.5.5 0 0 1 .5.5v6a.5.5 0 0 1-1 0V6a.5.5 0 0 1 .5-.5zm3 .5a.5.5 0 0 0-1 0v6a.5.5 0 0 0 1 0V6z"/>
  <path fill-rule="evenodd" d="M14.5 3a1 1 0 0 1-1 1H13v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V4h-.5a1 1 0 0 1-1-1V2a1 1 0 0 1 1-1H6a1 1 0 0 1 1-1h2a1 1 0 0 1 1 1h3.5a1 1 0 0 1 1 1v1zM4.118 4 4 4.059V13a1 1 0 0 0 1 1h6a1 1 0 0 0 1-1V4.059L11.882 4H4.118zM2.5 3V2h11v1h-11z"/>
</svg></a>` : ''

  return `
  <div class="playlist-card" id="id-${playlist.ulid}">
    <div class="playlist-card-heading">
      <div class="love-button">
        ${visibility}
        ${editButton}
        ${deleteButton}
        <a class="${favstat} action-button" onclick="favHandler('${playlist.ulid}', ${playlist.is_favorited})"><svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" fill="currentColor" class="bi bi-heart-fill" viewBox="0 0 16 16">
        <path fill-rule="evenodd" d="M8 1.314C12.438-3.248 23.534 4.735 8 15-7.534 4.736 3.562-3.248 8 1.314z"/>
      </svg></a>
      </div>
      <h3><a href="/playlist/${playlist.ulid}">${playlist.name}</a></h3>
    </div>
    <div class="playlist-card-body">
      <div class="playlist-art"></div>
      <div class="created-by">
        <label>作成者: </label> ${playlist.user_display_name}
      </div>
      <div class="song-count">
        <label>曲数: </label> ${playlist.song_count}曲
      </div>
      <div class="fav-count">
        <label>集めたラブ: </label> ${playlist.favorite_count}コ
      </div>
      <div class="date">
        作成: ${createdAt} | 最終更新: ${updatedAt}
      </div>  
    </div>
  </div>`
}