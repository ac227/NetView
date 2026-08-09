export interface Item {
  id: number
  type: 'image' | 'video'
  title: string
  description: string
  source_url: string
  local_path: string
  thumbnail_path: string
  mime_type: string
  size: number
  width: number
  height: number
  duration: number
  status: string
  favorite: boolean
  created_at: string
  updated_at: string
  tags: string[]
  categories: number[]
}

export interface Meta {
  title: string
  description: string
  image: string
  video: string
  content_type: string
}

export interface Category {
  id: number
  name: string
  parent_id: number | null
  sort: number
  count: number
}

export interface Job {
  id: number
  item_id: number
  url: string
  adapter: string
  status: string
  progress: number
  info: string
  error: string
}

export interface Settings {
  ai: {
    base_url: string
    api_key: string
    model: string
    configured: boolean
  }
  has_password: boolean
}

export interface Stats {
  items: number
  images: number
  videos: number
  favorites: number
  pending_jobs: number
  disk_bytes: number
}

export interface ListResponse {
  items: Item[]
  total: number
}

export interface ListParams {
  keyword?: string
  type?: string
  status?: string
  favorite?: boolean
  tag?: string
  category?: number
  sort?: string
  page?: number
  page_size?: number
}
