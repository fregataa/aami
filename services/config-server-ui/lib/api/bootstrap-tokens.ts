import { api } from './client'
import type { BootstrapToken } from '@/types/api'

export interface CreateBootstrapTokenRequest {
  name: string
  description?: string
  group_id: string
  expires_at?: string
  max_uses?: number
}

export interface UpdateBootstrapTokenRequest {
  name?: string
  description?: string
  group_id?: string
  expires_at?: string
  max_uses?: number
}

export interface BootstrapTokenListResponse {
  tokens: BootstrapToken[]
  total: number
}

export const bootstrapTokensApi = {
  list: () => api.get<BootstrapTokenListResponse>('/api/v1/bootstrap-tokens'),

  getById: (id: string) => api.get<BootstrapToken>(`/api/v1/bootstrap-tokens/${id}`),

  getByToken: (token: string) =>
    api.get<BootstrapToken>(`/api/v1/bootstrap-tokens/token/${token}`),

  create: (data: CreateBootstrapTokenRequest) =>
    api.post<BootstrapToken>('/api/v1/bootstrap-tokens', data),

  update: (id: string, data: UpdateBootstrapTokenRequest) =>
    api.put<BootstrapToken>(`/api/v1/bootstrap-tokens/${id}`, data),

  delete: (id: string) =>
    api.delete('/api/v1/bootstrap-tokens/delete', { id }),

  restore: (id: string) =>
    api.post('/api/v1/bootstrap-tokens/restore', { id }),

  purge: (id: string) =>
    api.post('/api/v1/bootstrap-tokens/purge', { id }),
}
