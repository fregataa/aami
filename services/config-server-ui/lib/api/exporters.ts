import { api } from './client'
import type { Exporter } from '@/types/api'

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
  list: () => api.get<Exporter[]>('/api/v1/exporters'),

  getById: (id: string) => api.get<Exporter>(`/api/v1/exporters/${id}`),

  getByTarget: (targetId: string) =>
    api.get<Exporter[]>(`/api/v1/exporters/target/${targetId}`),

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
