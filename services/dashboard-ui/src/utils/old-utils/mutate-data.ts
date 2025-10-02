import { API_URL } from '@/configs/api'
import { getFetchOpts, type TQuery } from '@/utils'

export interface IMutateData {
  data?: Record<string, unknown>
  errorMessage?: string
  method?: 'POST' | 'PATCH' | 'DELETE'
  orgId?: string
  path: string
  pathVersion?: 'v1'
}

export async function mutateData<T>({
  data,
  errorMessage = 'Encountered an issue retrieving this information, please refresh the page to try again.',
  method = 'POST',
  orgId,
  path,
  pathVersion = 'v1',
}: IMutateData): Promise<T> {
  const res = await fetch(`${API_URL}/${pathVersion}/${path}`, {
    ...(await getFetchOpts(orgId, {}, 10000)),
    body: data && JSON.stringify(data),
    method,
  })

  if (!res.ok) {
    throw new Error(errorMessage)
  }

  return res.json()
}

export interface INueMutateData {
  abortTimeout?: number
  body?: Record<string, unknown>
  method?: 'POST' | 'PATCH' | 'DELETE'
  orgId?: string
  path: string
  pathVersion?: 'v1'
}

export async function nueMutateData<T>({
  abortTimeout = 10000,
  body,
  method = 'POST',
  orgId,
  path,
  pathVersion = 'v1',
}: INueMutateData): Promise<TQuery<T>> {
  return fetch(`${API_URL}/${pathVersion}/${path}`, {
    ...(await getFetchOpts(orgId, {}, abortTimeout)),
    body: body ? JSON.stringify(body) : undefined,
    method,
  })
    .then((r) =>
      r.json().then((data) =>
        r.ok
          ? {
              data,
              error: null,
              status: r.status,
              headers: r.headers,
            }
          : { data: null, error: data, status: r.status, headers: r.headers }
      )
    )
    .catch((error) => {
      console.error(error)
      return { data: null, error, status: 500, headers: new Headers() }
    })
}
