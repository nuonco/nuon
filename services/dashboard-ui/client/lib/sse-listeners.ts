import type { QueryClient, QueryKey } from '@tanstack/react-query'

export type TSSEListener = (event: MessageEvent) => void
export type TSSEListenerMap = Record<string, TSSEListener>

export function createSSEQueryListener<T>(
  queryClient: QueryClient,
  queryKey: QueryKey | ((data: T) => QueryKey),
  options?: {
    transform?: (data: T) => unknown
    onData?: (data: T) => void
  }
): TSSEListener {
  return (event: MessageEvent) => {
    try {
      const data: T = JSON.parse(event.data)
      const key = typeof queryKey === 'function' ? queryKey(data) : queryKey
      queryClient.setQueryData(
        key,
        options?.transform ? options.transform(data) : data
      )
      options?.onData?.(data)
    } catch {}
  }
}
