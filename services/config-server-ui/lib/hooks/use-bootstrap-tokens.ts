import useSWR from 'swr'
import { bootstrapTokensApi } from '@/lib/api/bootstrap-tokens'

export function useBootstrapTokens() {
  const { data, error, isLoading, mutate } = useSWR(
    '/api/v1/bootstrap-tokens',
    bootstrapTokensApi.list
  )

  return {
    tokens: data ?? [],
    isLoading,
    error,
    mutate,
  }
}

export function useBootstrapToken(id: string | undefined) {
  const { data, error, isLoading, mutate } = useSWR(
    id ? `/api/v1/bootstrap-tokens/${id}` : null,
    () => bootstrapTokensApi.getById(id!)
  )

  return {
    token: data,
    isLoading,
    error,
    mutate,
  }
}
