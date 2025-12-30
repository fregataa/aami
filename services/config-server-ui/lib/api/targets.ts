import { api } from './client'
import type { Target } from '@/types/api'

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
  list: () => api.get<Target[]>('/api/v1/targets'),

  getById: (id: string) => api.get<Target>(`/api/v1/targets/${id}`),

  getByHostname: (hostname: string) =>
    api.get<Target>(`/api/v1/targets/hostname/${hostname}`),

  getByGroup: (groupId: string) =>
    api.get<Target[]>(`/api/v1/targets/group/${groupId}`),

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
