import { API_URL } from '@/configs/api'
import type { TAPIResponse } from '@/types'
import { auth0 } from './auth'

interface IAPIData {
  abortTimeout?: number
  headers?: Record<string, unknown>
  orgId?: string
  path: string
  pathVersion?: '/v1' | ''
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  body?: any
}

export async function api<T>({
  abortTimeout = 10000,
  headers = {},
  orgId,
  path,
  pathVersion = '/v1',
  method = 'GET',
  body,
}: IAPIData): Promise<TAPIResponse<T>> {
  let response: Response | undefined
  try {
    const session = await auth0.getSession()
    const fetchOpts: RequestInit = {
      cache: 'no-store',
      method,
      headers: {
        Authorization: `Bearer ${session?.tokenSet?.accessToken}`,
        'Content-Type': 'application/json',
        'X-Nuon-Org-ID': orgId || '',
        'x-nuon-pagination-enabled': 'true',
        ...headers,
      },
      signal: AbortSignal.timeout(abortTimeout),
    }
    if (body !== undefined && method !== 'GET') {
      fetchOpts.body = JSON.stringify(body)
    }

    response = await fetch(`${API_URL}${pathVersion}/${path}`, fetchOpts)

    const data = await response.json()

    if (response.ok) {
      return {
        data,
        error: null,
        status: response.status,
        headers: response.headers,
      }
    } else {
      if (response.status === 502) {
        console.warn('Received 502 Bad Gateway from API')
        return {
          data: null,
          error: {
            description:
              'The server is temporarily unavailable. Please try again later.',
            error: 'Bad Gateway',
            user_error: true,
          },
          status: response.status,
          headers: response.headers,
        }
      }

      return {
        data: null,
        error: data,
        status: response.status,
        headers: response.headers,
      }
    }
  } catch (error) {
    // Handle timeout error specifically
    let timeoutError = false
    // DOMException for AbortSignal.timeout in Node 20+ and modern browsers
    if (error instanceof DOMException && error.name === 'TimeoutError') {
      timeoutError = true
    } else if (error instanceof Error && error.name === 'AbortError') {
      // Fallback for environments using AbortController/AbortSignal
      timeoutError = true
    }

    const errorResponse = {
      data: null,
      error: timeoutError
        ? {
            description:
              'The request timed out. Please check your connection and try again.',
            error: 'Timeout',
            user_error: true,
          }
        : {
            description: 'An unexpected error occurred while fetching data.',
            error: error instanceof Error ? error.message : 'Unknown Error',
            user_error: false,
          },
      status: timeoutError ? 408 : 500,
      headers: response?.headers ?? new Headers(),
    }

    if (timeoutError) {
      console.warn('API request timed out:', error)
    } else {
      console.error('Error fetching data:', error)
    }

    return errorResponse
  }
}
