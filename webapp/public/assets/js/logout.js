window.addEventListener('load', () => {
  const form = document.getElementById('logout-form')
  
  const logout = async () => {
    const response = await fetch('/api/logout', {
      method: 'POST',
    })
  
    return response
  }
  
  form.addEventListener('submit', async (event) => {
    event.stopPropagation()
    event.preventDefault()
  
    const res = await logout()
    if (res.status !== 200) {
      // something error
      return
    }
  
    window.location.href = '/'
  })

})
