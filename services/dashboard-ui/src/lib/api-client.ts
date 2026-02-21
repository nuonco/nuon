import type { TAPIResponse } from '@/types'

/**
 * Browser-compatible API client for SPA mode.
 * Relies on the browser automatically sending the HttpOnly X-Nuon-Auth
 * cookie via credentials: 'include'. The Go BFF auth middleware reads
 * the cookie from the request — no need to read it via JS.
 */

interface IAPIClientOptions {
  path: string
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  body?: any
  orgId?: string
  headers?: Record<string, string>
  timeout?: number
}

export async function apiClient<T>({
  path,
  method = 'GET',
  body,
  orgId,
  headers = {},
  timeout = 10000,
}: IAPIClientOptions): Promise<TAPIResponse<T>> {
  const fetchOptions: RequestInit = {
    method,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      'x-nuon-pagination-enabled': 'true',
      ...(orgId && { 'X-Nuon-Org-ID': orgId }),
      ...headers,
    },
    signal: AbortSignal.timeout(timeout),
  }

  if (body) {
    fetchOptions.body = JSON.stringify(body)
  }

  try {
    const response = await fetch(path, fetchOptions)
    const headersObj = Object.fromEntries(response.headers.entries())

    // Return 401 to caller — let the caller decide whether to redirect
    if (response.status === 401) {
      return {
        data: null,
        error: { error: 'unauthorized', description: 'Session expired' },
        status: 401,
        headers: headersObj,
      }
    }

    let data = null
    const contentType = response.headers.get('content-type')
    if (contentType?.includes('application/json')) {
      const text = await response.text()
      if (text) {
        data = JSON.parse(text)
      }
    }

    // The Go BFF wraps responses in { data, error, status, headers } (TAPIResponse).
    // Unwrap the envelope so callers get the actual data.
    if (data && typeof data === 'object' && 'data' in data && 'status' in data) {
      return {
        data: data.data,
        error: data.error ?? null,
        status: data.status ?? response.status,
        headers: headersObj,
      }
    }

    if (response.ok) {
      return { data, error: null, status: response.status, headers: headersObj }
    } else {
      return { data: null, error: data, status: response.status, headers: headersObj }
    }
  } catch (error) {
    return {
      data: null,
      error: {
        error: 'network_error',
        description: error instanceof Error ? error.message : 'Network request failed',
      },
      status: 500,
      headers: {},
    }
  }
}