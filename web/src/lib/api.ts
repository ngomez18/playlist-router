import { getAuthToken, removeAuthToken } from './auth'
import type { User } from '../types/auth'
import type { BasePlaylist, CreateBasePlaylistRequest } from '../types/playlist'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || ''

class ApiClient {
  private baseURL: string

  constructor(baseURL: string) {
    this.baseURL = baseURL
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseURL}${endpoint}`
    const token = getAuthToken()

    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    }

    if (options.headers) {
      Object.assign(headers, options.headers)
    }

    if (token) {
      headers.Authorization = `Bearer ${token}`
    }

    const config: RequestInit = {
      ...options,
      headers,
    }

    const response = await fetch(url, config)

    if (!response.ok) {
      if (response.status === 401) {
        // Token expired or invalid, should trigger logout
        removeAuthToken()
        window.location.href = '/'
      }
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    return response.json()
  }

  // Auth endpoints
  async validateToken(): Promise<User> {
    return this.request<User>('/api/auth/validate')
  }

  // Base playlist endpoints
  async getBasePlaylist(id: string): Promise<BasePlaylist> {
    return this.request<BasePlaylist>(`/api/base_playlist/${id}`)
  }

  async createBasePlaylist(data: CreateBasePlaylistRequest): Promise<BasePlaylist> {
    return this.request<BasePlaylist>('/api/base_playlist', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async deleteBasePlaylist(id: string): Promise<void> {
    return this.request<void>(`/api/base_playlist/${id}`, {
      method: 'DELETE',
    })
  }
}

export const apiClient = new ApiClient(API_BASE_URL)