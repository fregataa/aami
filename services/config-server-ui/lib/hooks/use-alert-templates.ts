import useSWR from 'swr'
import { alertTemplatesApi } from '@/lib/api/alert-templates'

export function useAlertTemplates() {
  const { data, error, isLoading, mutate } = useSWR(
    '/api/v1/alert-templates',
    alertTemplatesApi.list
  )

  return {
    templates: data ?? [],
    isLoading,
    error,
    mutate,
  }
}

export function useAlertTemplate(id: string) {
  const { data, error, isLoading, mutate } = useSWR(
    id ? `/api/v1/alert-templates/${id}` : null,
    () => alertTemplatesApi.getById(id)
  )

  return {
    template: data,
    isLoading,
    error,
    mutate,
  }
}
