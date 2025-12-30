import { api } from './client'
import type { Target, PaginatedResponse } from '@/types/api'

export interface CreateTargetRequest {
  hostname: string
  ip_address: string
  port?: number
  labels?: Record<string, string>
  group_ids?: string[]
}

export interface UpdateTargetRequest {
  hostname?: string
  ip_address?: string
  port?: number
  labels?: Record<string, string>
  group_ids?: string[]
}

export const targetsApi = {
  list: async () => {
    const response = await api.get<PaginatedResponse<Target>>('/api/v1/targets')
    return response.data
  },

  getById: (id: string) => api.get<Target>(`/api/v1/targets/${id}`),

  getByHostname: (hostname: string) =>
    api.get<Target>(`/api/v1/targets/hostname/${hostname}`),

  getByGroup: async (groupId: string) => {
    const response = await api.get<PaginatedResponse<Target>>(`/api/v1/targets/group/${groupId}`)
    return response.data
  },

  create: (data: CreateTargetRequest) =>
    api.post<Target>('/api/v1/targets', data),

  update: (id: string, data: UpdateTargetRequest) =>
    api.put<Target>(`/api/v1/targets/${id}`, data),

  delete: (id: string) =>
    api.delete('/api/v1/targets/delete', { id }),

  restore: (id: string) =>
    api.post('/api/v1/targets/restore', { id }),

  purge: (id: string) =>
    api.post('/api/v1/targets/purge', { id }),
}
