import useSWR from 'swr'
import { exportersApi } from '@/lib/api/exporters'

export function useExporters() {
  const { data, error, isLoading, mutate } = useSWR(
    '/api/v1/exporters',
    exportersApi.list
  )

  return {
    exporters: data ?? [],
    isLoading,
    error,
    mutate,
  }
}

export function useExporter(id: string | undefined) {
  const { data, error, isLoading, mutate } = useSWR(
    id ? `/api/v1/exporters/${id}` : null,
    () => exportersApi.getById(id!)
  )

  return {
    exporter: data,
    isLoading,
    error,
    mutate,
  }
}
