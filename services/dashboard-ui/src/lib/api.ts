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
        Accept: 'application/json',
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

    // Convert headers to a plain object for serialization
    const headersObj = Object.fromEntries(response.headers.entries())

    // Handle empty response bodies (common with DELETE requests)
    let data = null
    const contentType = response.headers.get('content-type')
    const contentLength = response.headers.get('content-length')

    // Only try to parse JSON if there's actually content
    if (contentLength !== '0' && contentType?.includes('application/json')) {
      const text = await response.text()

      if (text) {
        try {
          data = JSON.parse(text)
        } catch (parseError) {
          console.warn('Failed to parse response as JSON:', parseError)
          data = text // Return as text if JSON parsing fails
        }
      }
    }

    if (
      contentLength !== '0' &&
      (contentType?.includes('text/csv') ||
        contentType?.includes('application/octet-stream'))
    ) {
      const content = await response.text()
      let filename = contentType?.includes('text/csv')
        ? 'data.csv'
        : 'download.bin'
      const contentDisposition = response.headers.get('content-disposition')
      const filenameMatch = contentDisposition?.match(/filename="?([^"]+)"?/)
      if (filenameMatch) {
        filename = filenameMatch[1].replace(/^["'_]+|["'_]+$/g, '').trim()
      }
      data = { content, filename }
    }

    if (response.ok) {
      return {
        data,
        error: null,
        status: response.status,
        headers: headersObj,
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
          headers: headersObj,
        }
      }

      return {
        data: null,
        error: data || {
          error: 'Unknown error',
          description: 'No error details provided',
        },
        status: response.status,
        headers: headersObj,
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

    // Convert headers to object if available, otherwise empty object
    const errorHeadersObj = response
      ? Object.fromEntries(response.headers.entries())
      : {}

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
      headers: errorHeadersObj,
    }

    if (timeoutError) {
      console.warn('API request timed out:', error)
    } else {
      console.error('Error fetching data:', error)
    }

    return errorResponse
  }
}
