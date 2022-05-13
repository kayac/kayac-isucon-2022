
window.addEventListener('load', () => {
  const playlistUlid = document.getElementById('playlist-id').getAttribute('value')

  const editingSongsState = []
  const upSong = (index) => {
    const tmp = editingSongsState[index-1]
    editingSongsState[index-1] = editingSongsState[index]
    editingSongsState[index] = tmp
    renderSongList()
  }

  const downSong = (index) => {
    const tmp = editingSongsState[index+1]
    editingSongsState[index+1] = editingSongsState[index]
    editingSongsState[index] = tmp
    renderSongList()
  }

  const deleteSong = (index) => {
    editingSongsState.splice(index, 1)
    renderSongList()
  }

  const candidateSongs = []
  const fetchSongCandidate = async () => {
    const res = await fetch('/assets/songs.json')
    const songs = await res.json()

    candidateSongs.splice(0)
    candidateSongs.push(...songs)
    candidateSongs.splice(1000)

    const select = document.getElementById('candidate-songs')
    const opts = candidateSongs.map(song => {
      return `<option value="${song.ulid}">${song.title} - ${song.artist_id} (${song.album}の${song.track_number}曲目)</option>`
    })
    select.innerHTML = opts.join('\n')
  }

  const addSongHandler = () => {
    // 81曲以上は追加できない
    if (editingSongsState.length >= 80) {
      window.alert('プレイリストに追加できるのは80曲までです。')
      return
    }

    const select = document.getElementById('candidate-songs')
    const songUlid = select.value
    const candidate = candidateSongs.filter(x => x.ulid === songUlid)
    if (!candidate.length) {
      // 未選択
      return
    }

    // 既にsong listにいるやつは追加できない
    const checker = editingSongsState.filter(x => x.ulid === songUlid)
    if (checker.length) {
      window.alert('既にプレイリストにある曲です! 追加できません。')
      return
    }

    const song = candidate[0]
    editingSongsState.unshift(song)
    renderSongList()
  }


  const renderSongList = () => {
    const tbody = document.getElementById('songs-table-body')
    const rowHTML = editingSongsState.map((song, index) => {
      return `
      <tr>
        <td>${index+1}</td>
        <td>${song.title}</td>
        <td>${song.artist}</td>
        <td>${song.album}</td>
        <td>${song.track_number}</td>
        <td>
          <button id="up-${index}">↑</button>
          <button id="down-${index}">↓</button>
          <button id="delete-${index}">削除</button>
        </td>
      </tr>
      `
    })

    tbody.innerHTML = rowHTML.join('\n')
    editingSongsState.forEach((_, index) => {
      const upButton = document.getElementById(`up-${index}`)
      const downButton = document.getElementById(`down-${index}`)
      const delButton = document.getElementById(`delete-${index}`)
      upButton.addEventListener('click', () => { upSong(index) })
      downButton.addEventListener('click', () => { downSong(index) })
      delButton.addEventListener('click', () => { deleteSong(index) })
    })
  }

  const getPlaylist = async () => {

    const response = await fetch('/api/playlist/' + playlistUlid)
    const json = await response.json()
    if (!json.result) {
      alert('プレイリスト取得失敗 (他人の非公開プレイリストかも?)')
    }

    const playlist = json.playlist

    const namebox = document.getElementById('playlist-name')
    namebox.value = playlist.name

    const radioPublic = document.getElementById('public')
    const radioPrivate = document.getElementById('private')

    if (playlist.is_public) {
      radioPublic.checked = true
    } else {
      radioPrivate.checked = true
    }

    // clear previous state
    editingSongsState.splice(0)
    editingSongsState.push(...playlist.songs)

    renderSongList()

    return response
  }

  const editPlaylist = async (postData) => {
    const response = await fetch(`/api/playlist/${playlistUlid}/update`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(postData)
    })

    return response
  }

  const saveHandler = async () => {

    // current values
    const newName = document.getElementById('playlist-name').value
    const newVisibility = document.getElementById('playlist-edit-form').elements['is_public'].value === 'true'

    const postData = {
      name: newName,
      is_public: newVisibility,
      song_ulids: editingSongsState.map(x => x.ulid),
    }

    const res = await editPlaylist(postData)

    const event = new Event('refreshRequired')
    const elem = document.getElementById('main-app')
    elem.dispatchEvent(event)
  }
  const saveButton = document.getElementById('save')
  saveButton.addEventListener('click', saveHandler)
  const save2Button = document.getElementById('save2')
  save2Button.addEventListener('click', saveHandler)

  const addButton = document.getElementById('add-song')
  addButton.addEventListener('click', addSongHandler)

  const main = document.getElementById('main-app')
  main.addEventListener('refreshRequired', () => {
    getPlaylist()
  })

  fetchSongCandidate()
  getPlaylist()
})
