import { api } from './client'
import type { Exporter, PaginatedResponse } from '@/types/api'

export interface CreateExporterRequest {
  target_id: string
  type: string
  port: number
  path?: string
  enabled?: boolean
}

export interface UpdateExporterRequest {
  type?: string
  port?: number
  path?: string
  enabled?: boolean
}

export const exportersApi = {
  list: async () => {
    const response = await api.get<PaginatedResponse<Exporter>>('/api/v1/exporters')
    return response.data
  },

  getById: (id: string) => api.get<Exporter>(`/api/v1/exporters/${id}`),

  getByTarget: async (targetId: string) => {
    const response = await api.get<PaginatedResponse<Exporter>>(`/api/v1/exporters/target/${targetId}`)
    return response.data
  },

  create: (data: CreateExporterRequest) =>
    api.post<Exporter>('/api/v1/exporters', data),

  update: (id: string, data: UpdateExporterRequest) =>
    api.put<Exporter>(`/api/v1/exporters/${id}`, data),

  delete: (id: string) =>
    api.delete('/api/v1/exporters/delete', { id }),

  restore: (id: string) =>
    api.post('/api/v1/exporters/restore', { id }),

  purge: (id: string) =>
    api.post('/api/v1/exporters/purge', { id }),
}
