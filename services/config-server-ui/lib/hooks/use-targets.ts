import useSWR from 'swr'
import { targetsApi } from '@/lib/api/targets'

export function useTargets() {
  const { data, error, isLoading, mutate } = useSWR(
    '/api/v1/targets',
    targetsApi.list
  )

  return {
    targets: data ?? [],
    isLoading,
    error,
    mutate,
  }
}

export function useTarget(id: string | undefined) {
  const { data, error, isLoading, mutate } = useSWR(
    id ? `/api/v1/targets/${id}` : null,
    () => targetsApi.getById(id!)
  )

  return {
    target: data,
    isLoading,
    error,
    mutate,
  }
}
