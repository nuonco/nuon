import { API_URL } from '@/configs/api'
import type { TAPIError } from '@/types'

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
}: IAPIData): Promise<T> {
  let response: Response | undefined
  try {
    const fetchOpts: RequestInit = {
      cache: 'no-store',
      credentials: 'include',
      method,
      headers: {
        Accept: 'application/json',
        'Content-Type': 'application/json',
        'X-Nuon-Org-ID': orgId || '',
        ...headers,
      },
      signal: AbortSignal.timeout(abortTimeout),
    }
    if (body !== undefined && method !== 'GET') {
      fetchOpts.body = JSON.stringify(body)
    }

    response = await fetch(`${API_URL}${pathVersion}/${path}`, fetchOpts)

    let data = null
    const contentType = response.headers.get('content-type')
    const contentLength = response.headers.get('content-length')

    if (contentLength !== '0' && contentType?.includes('application/json')) {
      const text = await response.text()

      if (text) {
        try {
          data = JSON.parse(text)
        } catch (parseError) {
          console.warn('Failed to parse response as JSON:', parseError)
          data = text
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
      return data as T
    } else {
      if (response.status === 401) {
        window.location.href = '/login'
      }

      if (response.status === 502) {
        console.warn('Received 502 Bad Gateway from API')
        throw {
          description:
            'The server is temporarily unavailable. Please try again later.',
          error: 'Bad Gateway',
          user_error: true,
        } satisfies TAPIError
      }

      throw (data ?? {
        error: 'Unknown error',
        description: 'No error details provided',
        user_error: false,
      }) as TAPIError
    }
  } catch (error) {
    if (error && typeof error === 'object' && 'error' in error) {
      throw error
    }

    const isTimeout =
      (error instanceof DOMException && error.name === 'TimeoutError') ||
      (error instanceof Error && error.name === 'AbortError')

    if (isTimeout) {
      console.warn('API request timed out:', error)
      throw {
        description:
          'The request timed out. Please check your connection and try again.',
        error: 'Timeout',
        user_error: true,
      } satisfies TAPIError
    }

    console.error('Error fetching data:', error)
    throw {
      description: 'An unexpected error occurred while fetching data.',
      error: error instanceof Error ? error.message : 'Unknown Error',
      user_error: false,
    } satisfies TAPIError
  }
}
