import { api } from './client'
import type { Group } from '@/types/api'

export interface CreateGroupRequest {
  name: string
  description?: string
  priority?: number
  is_default_own?: boolean
  metadata?: Record<string, unknown>
}

export interface UpdateGroupRequest {
  name?: string
  description?: string
  priority?: number
  is_default_own?: boolean
  metadata?: Record<string, unknown>
}

export const groupsApi = {
  list: () => api.get<Group[]>('/api/v1/groups'),

  getById: (id: string) => api.get<Group>(`/api/v1/groups/${id}`),

  create: (data: CreateGroupRequest) =>
    api.post<Group>('/api/v1/groups', data),

  update: (id: string, data: UpdateGroupRequest) =>
    api.put<Group>(`/api/v1/groups/${id}`, data),

  delete: (id: string) =>
    api.delete('/api/v1/groups/delete', { id }),

  restore: (id: string) =>
    api.post('/api/v1/groups/restore', { id }),

  purge: (id: string) =>
    api.post('/api/v1/groups/purge', { id }),
}
