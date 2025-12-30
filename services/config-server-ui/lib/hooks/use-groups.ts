import useSWR from 'swr'
import { groupsApi } from '@/lib/api/groups'

export function useGroups() {
  const { data, error, isLoading, mutate } = useSWR(
    '/api/v1/groups',
    groupsApi.list
  )

  return {
    groups: data ?? [],
    isLoading,
    error,
    mutate,
  }
}

export function useGroup(id: string | undefined) {
  const { data, error, isLoading, mutate } = useSWR(
    id ? `/api/v1/groups/${id}` : null,
    () => groupsApi.getById(id!)
  )

  return {
    group: data,
    isLoading,
    error,
    mutate,
  }
}
