import type { TCatalog, THealth, TRun } from './types'

const request = async <T,>(path: string, init?: RequestInit): Promise<T> => {
  const response = await fetch(path, init)
  if (!response.ok) {
    const body = (await response.json().catch(() => null)) as { error?: string } | null
    throw new Error(body?.error ?? `${response.status} ${response.statusText}`)
  }
  return response.json() as Promise<T>
}

export const getCatalog = () => request<TCatalog>('/api/catalog')
export const getHealth = () => request<THealth>('/api/health')
export const getRuns = () => request<TRun[]>('/api/runs')
export const getRun = (id: string) => request<TRun>(`/api/runs/${encodeURIComponent(id)}`)

export const dispatchRef = (refId: string) => {
  const csrf = document.querySelector<HTMLMetaElement>('meta[name="csrf-token"]')?.content
  return request<{ dispatch_id: string }>('/api/dispatch', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-CSRF-Token': csrf ?? '',
    },
    body: JSON.stringify({ ref_id: refId }),
  })
}
