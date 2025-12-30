import useSWR from 'swr'
import { alertRulesApi } from '@/lib/api/alert-rules'

export function useAlertRules() {
  const { data, error, isLoading, mutate } = useSWR(
    '/api/v1/alert-rules',
    alertRulesApi.list
  )

  return {
    rules: data ?? [],
    isLoading,
    error,
    mutate,
  }
}

export function useAlertRule(id: string) {
  const { data, error, isLoading, mutate } = useSWR(
    id ? `/api/v1/alert-rules/${id}` : null,
    () => alertRulesApi.getById(id)
  )

  return {
    rule: data,
    isLoading,
    error,
    mutate,
  }
}

export function useAlertRulesByGroup(groupId: string) {
  const { data, error, isLoading, mutate } = useSWR(
    groupId ? `/api/v1/alert-rules/group/${groupId}` : null,
    () => alertRulesApi.getByGroup(groupId)
  )

  return {
    rules: data ?? [],
    isLoading,
    error,
    mutate,
  }
}

export function useAlertRulesByTemplate(templateId: string) {
  const { data, error, isLoading, mutate } = useSWR(
    templateId ? `/api/v1/alert-rules/template/${templateId}` : null,
    () => alertRulesApi.getByTemplate(templateId)
  )

  return {
    rules: data ?? [],
    isLoading,
    error,
    mutate,
  }
}
