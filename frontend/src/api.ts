import axios from 'axios'
import type {
  Category, Item, Job, ListParams, ListResponse, Meta, Settings, Stats,
} from './types'

export const api = axios.create({ baseURL: '/api' })

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('nv_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('nv_token')
      if (window.location.pathname !== '/login') {
        window.location.href = '/login'
      }
    }
    return Promise.reject(err)
  },
)

export const authApi = {
  status: () => api.get<{ configured: boolean; needs_setup: boolean }>('/auth/status').then(r => r.data),
  login: (password: string) =>
    api.post<{ token: string; expires: string; user: string; setup: boolean }>('/auth/login', { password }).then(r => r.data),
}

export const itemApi = {
  list: (params: ListParams) => api.get<ListResponse>('/items', { params }).then(r => r.data),
  get: (id: number) => api.get<Item>(`/items/${id}`).then(r => r.data),
  create: (data: Partial<Item>) => api.post<Item>('/items', data).then(r => r.data),
  update: (id: number, data: Partial<Item>) => api.put<Item>(`/items/${id}`, data).then(r => r.data),
  remove: (id: number) => api.delete(`/items/${id}`).then(r => r.data),
  upload: (file: File, type: string, title?: string) => {
    const form = new FormData()
    form.append('file', file)
    form.append('type', type)
    if (title) form.append('title', title)
    return api.post<Item>('/items/upload', form).then(r => r.data)
  },
  favorite: (id: number, favorite: boolean) => api.post(`/items/${id}/favorite`, { favorite }).then(r => r.data),
  fetchMeta: (url: string) => api.post<Meta>('/items/fetch-meta', { url }).then(r => r.data),
  download: (id: number) => api.post(`/items/${id}/download`).then(r => r.data),
  aiTag: (id: number) => api.post(`/items/${id}/ai-tag`).then(r => r.data),
  fileUrl: (id: number) => `/api/items/${id}/file${mediaToken()}`,
  thumbUrl: (id: number) => `/api/items/${id}/thumbnail${mediaToken()}`,
}

function mediaToken() {
  const t = localStorage.getItem('nv_token')
  return t ? `?token=${encodeURIComponent(t)}` : ''
}

export const tagApi = {
  list: () => api.get<{ tags: string[] }>('/tags').then(r => r.data.tags),
}

export const categoryApi = {
  list: () => api.get<{ categories: Category[] }>('/categories').then(r => r.data.categories),
  create: (data: { name: string; parent_id?: number | null; sort?: number }) =>
    api.post<{ id: number }>('/categories', data).then(r => r.data),
  update: (id: number, data: { name: string; parent_id?: number | null; sort?: number }) =>
    api.put(`/categories/${id}`, data).then(r => r.data),
  remove: (id: number) => api.delete(`/categories/${id}`).then(r => r.data),
}

export const jobApi = {
  list: () => api.get<{ jobs: Job[] }>('/download/jobs').then(r => r.data.jobs),
  cancel: (id: number) => api.post(`/download/jobs/${id}/cancel`).then(r => r.data),
}

export const settingsApi = {
  get: () => api.get<Settings>('/settings').then(r => r.data),
  update: (data: { ai?: { base_url?: string; api_key?: string; model?: string }; password?: string }) =>
    api.put('/settings', data).then(r => r.data),
  stats: () => api.get<Stats>('/system/stats').then(r => r.data),
}
