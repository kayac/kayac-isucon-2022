window.addEventListener('load', () => {
  const form = document.getElementById('signup-form')

  const signup = async () => {
    const user_account = form['user_account'].value
    const password = form['password'].value
    const password2 = form['password2'].value
    const display_name = form['display_name'].value

    if (password !== password2) {
      window.alert('パスワードが一致していません!')
      return
    }

    const response = await fetch('/api/signup', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        user_account,
        password,
        display_name,
      })
    })

    return response
  }

  form.addEventListener('submit', async (event) => {
    event.stopPropagation()
    event.preventDefault()

    const res = await signup()
    if (res.status !== 200) {
      // something error
      return
    }

    window.location.href = '/mypage'
  })

})

