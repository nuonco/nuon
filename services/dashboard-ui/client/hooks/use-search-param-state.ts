import { useCallback } from 'react'
import { useSearchParams } from 'react-router'
import { useClearSearchParamsOnUnmount } from '@/hooks/use-clear-search-params-on-unmount'
import { useReplaceSearchParams } from '@/hooks/use-replace-search-params'

type TSetSearchParamValue = string | null

export const useSearchParamState = (
  key: string,
  { clearOnUnmount = false }: { clearOnUnmount?: boolean } = {}
): [string | null, (value: TSetSearchParamValue) => void] => {
  const [searchParams] = useSearchParams()
  const value = searchParams.get(key)
  const replaceSearch = useReplaceSearchParams()
  useClearSearchParamsOnUnmount(clearOnUnmount ? [key] : [])

  const setValue = useCallback(
    (next: TSetSearchParamValue) => {
      replaceSearch((params) => {
        if (next) params.set(key, next)
        else params.delete(key)
      })
    },
    [key, replaceSearch]
  )

  return [value, setValue]
}
