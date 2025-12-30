import { api } from './client'
import type { AlertTemplate } from '@/types/api'

export interface CreateAlertTemplateRequest {
  name: string
  description?: string
  severity: 'critical' | 'warning' | 'info'
  query_template: string
  default_config?: Record<string, unknown>
}

export interface UpdateAlertTemplateRequest {
  name?: string
  description?: string
  severity?: 'critical' | 'warning' | 'info'
  query_template?: string
  default_config?: Record<string, unknown>
}

export const alertTemplatesApi = {
  list: () => api.get<AlertTemplate[]>('/api/v1/alert-templates'),

  getById: (id: string) => api.get<AlertTemplate>(`/api/v1/alert-templates/${id}`),

  create: (data: CreateAlertTemplateRequest) =>
    api.post<AlertTemplate>('/api/v1/alert-templates', data),

  update: (id: string, data: UpdateAlertTemplateRequest) =>
    api.put<AlertTemplate>(`/api/v1/alert-templates/${id}`, data),

  delete: (id: string) =>
    api.delete('/api/v1/alert-templates/delete', { id }),
}
