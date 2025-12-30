import useSWR from 'swr'
import { scriptTemplatesApi } from '@/lib/api/script-templates'

export function useScriptTemplates() {
  const { data, error, isLoading, mutate } = useSWR(
    '/api/v1/script-templates',
    scriptTemplatesApi.list
  )

  return {
    templates: data ?? [],
    isLoading,
    error,
    mutate,
  }
}

export function useScriptTemplate(id: string) {
  const { data, error, isLoading, mutate } = useSWR(
    id ? `/api/v1/script-templates/${id}` : null,
    () => scriptTemplatesApi.getById(id)
  )

  return {
    template: data,
    isLoading,
    error,
    mutate,
  }
}
