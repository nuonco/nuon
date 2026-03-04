import { API_URL } from '@/configs/api'
import type { TAPIError } from '@/types'

export type TPaginationMeta = {
  hasNext: boolean
  offset: number
  limit: number
}

export type TPaginatedResult<T> = {
  data: T
  pagination: TPaginationMeta
}

export type TWithMetaResult<T> = {
  data: T
  nextOffset: string | null
}

export type TWithHeadersResult<T> = {
  data: T
  headers: Record<string, string>
}

interface IAPIData {
  abortTimeout?: number
  headers?: Record<string, unknown>
  orgId?: string
  path: string
  pathVersion?: '/v1' | ''
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  body?: any
  paginated?: boolean
  withMeta?: boolean
  withHeaders?: boolean
}

export async function api<T>(opts: IAPIData & { withHeaders: true }): Promise<TWithHeadersResult<T>>
export async function api<T>(opts: IAPIData & { paginated: true }): Promise<TPaginatedResult<T>>
export async function api<T>(opts: IAPIData & { withMeta: true }): Promise<TWithMetaResult<T>>
export async function api<T>(opts: IAPIData & { paginated?: false; withMeta?: false; withHeaders?: false }): Promise<T>
export async function api<T>({
  abortTimeout = 10000,
  headers = {},
  orgId,
  path,
  pathVersion = '/v1',
  method = 'GET',
  body,
  paginated,
  withMeta,
  withHeaders,
}: IAPIData): Promise<T | TPaginatedResult<T> | TWithMetaResult<T> | TWithHeadersResult<T>> {
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
        contentType?.includes('application/octet-stream') ||
        contentType?.includes('text/plain') ||
        contentType?.includes('application/toml'))
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
      if (paginated) {
        return {
          data: data as T,
          pagination: {
            hasNext: response.headers.get('X-Nuon-Page-Next') === 'true',
            offset: Number(response.headers.get('X-Nuon-Page-Offset') ?? 0),
            limit: Number(response.headers.get('X-Nuon-Page-Limit') ?? 20),
          },
        }
      }
      if (withMeta) {
        return { data: data as T, nextOffset: response.headers.get('x-nuon-api-next') }
      }
      if (withHeaders) {
        const responseHeaders: Record<string, string> = {}
        response.headers.forEach((value, key) => { responseHeaders[key] = value })
        return { data: data as T, headers: responseHeaders }
      }
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
          status: 502,
        } satisfies TAPIError
      }

      throw {
        ...(data ?? {
          error: 'Unknown error',
          description: 'No error details provided',
          user_error: false,
        }),
        status: response.status,
      } as TAPIError
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
