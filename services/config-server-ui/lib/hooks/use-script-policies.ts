import useSWR from 'swr'
import { scriptPoliciesApi } from '@/lib/api/script-policies'

export function useScriptPolicies() {
  const { data, error, isLoading, mutate } = useSWR(
    '/api/v1/script-policies',
    scriptPoliciesApi.list
  )

  return {
    policies: data ?? [],
    isLoading,
    error,
    mutate,
  }
}

export function useScriptPolicy(id: string) {
  const { data, error, isLoading, mutate } = useSWR(
    id ? `/api/v1/script-policies/${id}` : null,
    () => scriptPoliciesApi.getById(id)
  )

  return {
    policy: data,
    isLoading,
    error,
    mutate,
  }
}

export function useScriptPoliciesByGroup(groupId: string) {
  const { data, error, isLoading, mutate } = useSWR(
    groupId ? `/api/v1/script-policies/group/${groupId}` : null,
    () => scriptPoliciesApi.getByGroup(groupId)
  )

  return {
    policies: data ?? [],
    isLoading,
    error,
    mutate,
  }
}

export function useScriptPoliciesByTemplate(templateId: string) {
  const { data, error, isLoading, mutate } = useSWR(
    templateId ? `/api/v1/script-policies/template/${templateId}` : null,
    () => scriptPoliciesApi.getByTemplate(templateId)
  )

  return {
    policies: data ?? [],
    isLoading,
    error,
    mutate,
  }
}
