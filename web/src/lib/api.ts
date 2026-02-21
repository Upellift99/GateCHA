import axios from 'axios'

const api = axios.create({ baseURL: '/api/admin' })

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('gatecha_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (r) => r,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('gatecha_token')
      window.location.href = '/login'
    }
    return Promise.reject(err)
  }
)

export default api
