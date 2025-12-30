import { api } from './client'
import type { ScriptPolicy } from '@/types/api'

export interface CreateScriptPolicyRequest {
  template_id: string
  group_id?: string
  config?: Record<string, unknown>
  priority?: number
  enabled?: boolean
}

export interface UpdateScriptPolicyRequest {
  template_id?: string
  group_id?: string
  config?: Record<string, unknown>
  priority?: number
  enabled?: boolean
}

export const scriptPoliciesApi = {
  list: () => api.get<ScriptPolicy[]>('/api/v1/script-policies'),

  getById: (id: string) => api.get<ScriptPolicy>(`/api/v1/script-policies/${id}`),

  getByGroup: (groupId: string) =>
    api.get<ScriptPolicy[]>(`/api/v1/script-policies/group/${groupId}`),

  getByTemplate: (templateId: string) =>
    api.get<ScriptPolicy[]>(`/api/v1/script-policies/template/${templateId}`),

  create: (data: CreateScriptPolicyRequest) =>
    api.post<ScriptPolicy>('/api/v1/script-policies', data),

  update: (id: string, data: UpdateScriptPolicyRequest) =>
    api.put<ScriptPolicy>(`/api/v1/script-policies/${id}`, data),

  delete: (id: string) =>
    api.delete('/api/v1/script-policies/delete', { id }),
}
