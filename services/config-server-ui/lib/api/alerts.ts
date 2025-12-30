import { api } from './client'
import type { ActiveAlertsResponse } from '@/types/api'

export const alertsApi = {
  getActive: () => api.get<ActiveAlertsResponse>('/api/v1/alerts/active'),
}
