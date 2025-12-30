import { api } from './client'
import type { AlertRule } from '@/types/api'

export interface CreateAlertRuleRequest {
  group_id: string
  name: string
  description?: string
  severity: 'critical' | 'warning' | 'info'
  query_template: string
  default_config?: Record<string, unknown>
  config?: Record<string, unknown>
  enabled?: boolean
  merge_strategy?: string
  priority?: number
}

export interface UpdateAlertRuleRequest {
  name?: string
  description?: string
  severity?: 'critical' | 'warning' | 'info'
  query_template?: string
  default_config?: Record<string, unknown>
  config?: Record<string, unknown>
  enabled?: boolean
  merge_strategy?: string
  priority?: number
}

export interface CreateFromTemplateRequest {
  group_id: string
  template_id: string
  name?: string
  config?: Record<string, unknown>
  enabled?: boolean
}

export const alertRulesApi = {
  list: () => api.get<AlertRule[]>('/api/v1/alert-rules'),

  getById: (id: string) => api.get<AlertRule>(`/api/v1/alert-rules/${id}`),

  getByGroup: (groupId: string) =>
    api.get<AlertRule[]>(`/api/v1/alert-rules/group/${groupId}`),

  getByTemplate: (templateId: string) =>
    api.get<AlertRule[]>(`/api/v1/alert-rules/template/${templateId}`),

  create: (data: CreateAlertRuleRequest) =>
    api.post<AlertRule>('/api/v1/alert-rules', data),

  createFromTemplate: (data: CreateFromTemplateRequest) =>
    api.post<AlertRule>('/api/v1/alert-rules/from-template', data),

  update: (id: string, data: UpdateAlertRuleRequest) =>
    api.put<AlertRule>(`/api/v1/alert-rules/${id}`, data),

  delete: (id: string) =>
    api.delete('/api/v1/alert-rules/delete', { id }),

  toggleEnabled: (id: string, enabled: boolean) =>
    api.put<AlertRule>(`/api/v1/alert-rules/${id}`, { enabled }),
}
