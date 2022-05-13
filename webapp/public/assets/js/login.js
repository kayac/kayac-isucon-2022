window.addEventListener('load', () => {
  const form = document.getElementById('login-form')

  const login = async () => {
    const user_account = form['user_account'].value
    const password = form['password'].value

    const response = await fetch('/api/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        user_account,
        password,
      })
    })

    return response
  }

  form.addEventListener('submit', async (event) => {
    event.stopPropagation()
    event.preventDefault()

    const res = await login()
    if (res.status !== 200) {
      // something error
      if (res.status === 401) {
        window.alert('ログイン大失敗 (アカウント情報が違います)')
      } else {
        window.alert('ログイン大失敗 (サーバの調子が悪そう)')
      }
      return
    }

    window.location.href = '/mypage'
  })

})

