import { useCallback } from 'react'
import { useSearchParams } from 'react-router'
import { useClearSearchParamsOnUnmount } from '@/hooks/use-clear-search-params-on-unmount'

type TSetSearchParamValue = string | null

export const useSearchParamState = (
  key: string,
  { clearOnUnmount = false }: { clearOnUnmount?: boolean } = {}
): [string | null, (value: TSetSearchParamValue) => void] => {
  const [searchParams, setSearchParams] = useSearchParams()
  const value = searchParams.get(key)
  useClearSearchParamsOnUnmount(clearOnUnmount ? [key] : [])

  const setValue = useCallback(
    (next: TSetSearchParamValue) => {
      setSearchParams(
        (prev) => {
          const params = new URLSearchParams(prev)
          if (next) params.set(key, next)
          else params.delete(key)
          return params
        },
        { replace: true }
      )
    },
    [key, setSearchParams]
  )

  return [value, setValue]
}
