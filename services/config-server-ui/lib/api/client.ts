import { config } from '@/lib/config'

export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
    public details?: string
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

interface ErrorResponse {
  error: {
    code: string
    message: string
    details?: string
  }
}

async function handleResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const error: ErrorResponse = await response.json().catch(() => ({
      error: { code: 'UNKNOWN', message: 'Unknown error occurred' }
    }))
    throw new ApiError(
      response.status,
      error.error.code,
      error.error.message,
      error.error.details
    )
  }

  // Handle 204 No Content
  if (response.status === 204) {
    return undefined as T
  }

  return response.json()
}

export const api = {
  async get<T>(path: string): Promise<T> {
    const response = await fetch(`${config.apiUrl}${path}`)
    return handleResponse<T>(response)
  },

  async post<T>(path: string, body?: unknown): Promise<T> {
    const response = await fetch(`${config.apiUrl}${path}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: body ? JSON.stringify(body) : undefined,
    })
    return handleResponse<T>(response)
  },

  async put<T>(path: string, body: unknown): Promise<T> {
    const response = await fetch(`${config.apiUrl}${path}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    })
    return handleResponse<T>(response)
  },

  async delete(path: string, body: { id: string }): Promise<void> {
    const response = await fetch(`${config.apiUrl}${path}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    })
    return handleResponse<void>(response)
  },
}
