export type TAPIError = {
  error: string
  description?: string
  status?: number
}

let _basePath = ''

export function setBasePath(path: string) {
  _basePath = path
}

export function getBasePath(): string {
  return _basePath
}

interface IAPIData {
  path: string
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  body?: any
  params?: Record<string, any>
}

export async function api<T>({
  path,
  method = 'GET',
  body,
  params,
}: IAPIData): Promise<T> {
  let url = `${_basePath}/api/${path}`

  if (params) {
    const searchParams = new URLSearchParams()
    for (const [key, value] of Object.entries(params)) {
      if (value === undefined || value === null || value === '') continue
      if (Array.isArray(value)) {
        for (const v of value) {
          searchParams.append(key, String(v))
        }
      } else {
        searchParams.set(key, String(value))
      }
    }
    const qs = searchParams.toString()
    if (qs) url += `?${qs}`
  }

  const fetchOpts: RequestInit = {
    cache: 'no-store',
    credentials: 'include',
    method,
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
    },
  }

  if (body !== undefined && method !== 'GET') {
    fetchOpts.body = JSON.stringify(body)
  }

  const response = await fetch(url, fetchOpts)

  let data = null
  const contentType = response.headers.get('content-type')
  const contentLength = response.headers.get('content-length')

  if (contentLength !== '0' && contentType?.includes('application/json')) {
    const text = await response.text()
    if (text) {
      try {
        data = JSON.parse(text)
      } catch {
        data = text
      }
    }
  }

  if (response.ok) {
    return data as T
  }

  if (response.status === 401) {
    window.location.reload()
    return undefined as any
  }

  throw {
    ...(data ?? {
      error: 'Unknown error',
      description: 'No error details provided',
    }),
    status: response.status,
  } as TAPIError
}
