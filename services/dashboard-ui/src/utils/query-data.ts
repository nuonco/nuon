import { API_URL } from '@/configs/api'
import { getFetchOpts } from '@/utils/get-fetch-opts'

export type TResponseError = {
  description: string
  error: string
  user_error: boolean
}

export type TOldQueryError = {
  errorMessage: string
  status: Response['status']
  error: TResponseError
}

export interface IQueryData {
  errorMessage?: string
  orgId?: string
  path: string
  pathVersion?: 'v1'
  abortTimeout?: number
}

export async function queryData<T>({
  errorMessage = 'Encountered an issue retrieving this information, please refresh the page to try again.',
  orgId,
  path,
  pathVersion = 'v1',
  abortTimeout = 10000,
}: IQueryData): Promise<T> {
  const res = await fetch(
    `${API_URL}/${pathVersion}/${path}`,
    await getFetchOpts(orgId, {}, abortTimeout)
  )

  if (!res.ok) {
    throw new Error(errorMessage)
    // const error = await res.json()

    // if (res.status >= 500) {

    // } else {
    //   return {
    //     status: res.status,
    //     error,
    //     errorMessage,
    //   }
    // }
  }

  return res.json()
}

export type TQueryError = {
  description: string
  error: string
  user_error: boolean
  meta?: any
}

export type TQuery<T> = {
  data: T | null
  error: null | TQueryError
  headers: Response['headers']
  status: Response['status']
}

interface INueQueryData {
  abortTimeout?: number
  headers?: Record<string, unknown>
  orgId?: string
  path: string
  pathVersion?: 'v1'
}

export async function nueQueryData<T>({
  abortTimeout = 10000,
  headers = {},
  orgId,
  path,
  pathVersion = 'v1',
}: INueQueryData): Promise<TQuery<T>> {
  try {
    const response = await fetch(
      `${API_URL}/${pathVersion}/${path}`,
      await getFetchOpts(orgId, headers, abortTimeout)
    )

    // Parse the response data
    const data = await response.json()

    if (response.ok) {
      // Handle successful responses
      return {
        data,
        error: null,
        status: response.status,
        headers: response.headers,
      }
    } else {
      // Explicitly handle 502 errors
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

      // Handle other non-OK responses
      return {
        data: null,
        error: data,
        status: response.status,
        headers: response.headers,
      }
    }
  } catch (error) {
    // Handle network errors or exceptions
    console.error('Error fetching data:', error)

    return {
      data: null,
      error: {
        description: 'An unexpected error occurred while fetching data.',
        error: error instanceof Error ? error.message : 'Unknown Error',
        user_error: false,
      },
      status: 500,
      headers: new Headers(),
    }
  }
}
