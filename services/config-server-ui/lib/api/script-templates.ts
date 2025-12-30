import { api } from './client'
import type { ScriptTemplate } from '@/types/api'

export interface CreateScriptTemplateRequest {
  name: string
  description?: string
  script_type: string
  script_content: string
  config_schema?: Record<string, unknown>
  enabled?: boolean
}

export interface UpdateScriptTemplateRequest {
  name?: string
  description?: string
  script_type?: string
  script_content?: string
  config_schema?: Record<string, unknown>
  enabled?: boolean
}

export interface VerifyHashResponse {
  valid: boolean
  expected_hash?: string
  actual_hash?: string
}

export const scriptTemplatesApi = {
  list: () => api.get<ScriptTemplate[]>('/api/v1/script-templates'),

  getById: (id: string) => api.get<ScriptTemplate>(`/api/v1/script-templates/${id}`),

  create: (data: CreateScriptTemplateRequest) =>
    api.post<ScriptTemplate>('/api/v1/script-templates', data),

  update: (id: string, data: UpdateScriptTemplateRequest) =>
    api.put<ScriptTemplate>(`/api/v1/script-templates/${id}`, data),

  delete: (id: string) =>
    api.delete('/api/v1/script-templates/delete', { id }),

  verifyHash: (id: string) =>
    api.get<VerifyHashResponse>(`/api/v1/script-templates/${id}/verify-hash`),
}
