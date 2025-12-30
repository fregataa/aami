import useSWR from 'swr'
import { alertsApi } from '@/lib/api/alerts'

export function useActiveAlerts() {
  const { data, error, isLoading, mutate } = useSWR(
    '/api/v1/alerts/active',
    alertsApi.getActive,
    { refreshInterval: 10000 }  // Refresh every 10 seconds
  )

  return {
    alerts: data?.alerts ?? [],
    total: data?.total ?? 0,
    isLoading,
    error,
    mutate,
  }
}
