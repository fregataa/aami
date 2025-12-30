import { api } from './client'
import type { PrometheusStatusResponse } from '@/types/api'

export interface RegenerateRulesResponse {
  message: string
  success: boolean
  groups_updated?: number
}

export interface ReloadResponse {
  message: string
  success: boolean
  healthy: boolean
}

export const prometheusApi = {
  getStatus: () => api.get<PrometheusStatusResponse>('/api/v1/prometheus/status'),

  reload: () => api.post<ReloadResponse>('/api/v1/prometheus/reload'),

  regenerateAllRules: () =>
    api.post<RegenerateRulesResponse>('/api/v1/prometheus/rules/regenerate'),

  regenerateGroupRules: (groupId: string) =>
    api.post<RegenerateRulesResponse>(`/api/v1/prometheus/rules/regenerate/${groupId}`),
}
