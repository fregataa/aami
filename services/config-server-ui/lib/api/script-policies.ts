import { api } from './client'
import type { ScriptPolicy, PaginatedResponse } from '@/types/api'

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
  list: async () => {
    const response = await api.get<PaginatedResponse<ScriptPolicy>>('/api/v1/script-policies')
    return response.data
  },

  getById: (id: string) => api.get<ScriptPolicy>(`/api/v1/script-policies/${id}`),

  getByGroup: async (groupId: string) => {
    const response = await api.get<PaginatedResponse<ScriptPolicy>>(`/api/v1/script-policies/group/${groupId}`)
    return response.data
  },

  getByTemplate: async (templateId: string) => {
    const response = await api.get<PaginatedResponse<ScriptPolicy>>(`/api/v1/script-policies/template/${templateId}`)
    return response.data
  },

  create: (data: CreateScriptPolicyRequest) =>
    api.post<ScriptPolicy>('/api/v1/script-policies', data),

  update: (id: string, data: UpdateScriptPolicyRequest) =>
    api.put<ScriptPolicy>(`/api/v1/script-policies/${id}`, data),

  delete: (id: string) =>
    api.delete('/api/v1/script-policies/delete', { id }),
}
