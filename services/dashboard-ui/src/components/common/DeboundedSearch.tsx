'use client'

import { useRouter, useSearchParams } from 'next/navigation'
import {
  useState,
  useEffect,
  useRef,
  useCallback,
  startTransition,
} from 'react'
import { usePagination } from '@/hooks/use-pagination'
import { SearchInput } from './SearchInput'

interface IDebouncedSearchInput {
  searchParamKey?: string
  initialValue?: string
  placeholder?: string
  debounceMs?: number
  className?: string
  onDebouncedChange?: (value: string) => void
}

export const DebouncedSearchInput = ({
  searchParamKey = 'q',
  initialValue,
  placeholder = 'Search…',
  debounceMs = 200,
  className,
  onDebouncedChange,
}: IDebouncedSearchInput) => {
  const router = useRouter()
  const searchParams = useSearchParams()
  const { setIsPaginating } = usePagination()
  const valFromUrl = searchParams?.get(searchParamKey) || ''

  const [value, setValue] = useState(initialValue ?? valFromUrl)

  // refs to avoid stale closures
  const valueRef = useRef<string>(initialValue ?? valFromUrl)
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const isFocusedRef = useRef(false)
  const isComposingRef = useRef(false)

  useEffect(() => {
    valueRef.current = value
  }, [value])

  // Sync from URL when url changes — but do not clobber while user is focused/typing
  useEffect(() => {
    if (isFocusedRef.current) return
    setValue(initialValue ?? valFromUrl)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [valFromUrl, initialValue])

  // Build URL helper
  const buildUrlForValue = useCallback(
    (v: string) => {
      const params = new URLSearchParams(window.location.search)
      const existing = params.get(searchParamKey) || ''

      if (v) {
        if (existing !== v) {
          params.delete('offset')
        }
        params.set(searchParamKey, v)
      } else {
        params.delete(searchParamKey)
        params.delete('offset')
      }

      const query = params.toString()
      return `${window.location.pathname}${query ? `?${query}` : ''}${window.location.hash || ''}`
    },
    [searchParamKey]
  )

  // Apply the URL update to history (no navigation) and then trigger Next navigation (router.replace)
  const applyUrlAndTriggerRouter = useCallback(
    (triggerRouter = true) => {
      if (typeof window === 'undefined') return
      const currentValue = valueRef.current
      const newUrl = buildUrlForValue(currentValue)

      // Update the visible URL immediately without causing a Next navigation
      // (prevents useSearchParams / app-router from updating while typing)
      window.history.replaceState({}, '', newUrl)

      // Let caller react to the debounced value
      onDebouncedChange?.(currentValue)

      // Trigger a Next navigation to make the app re-run loaders / server fetches that depend on the query.
      // Do this in a transition so it won't block UI updates.
      if (!triggerRouter) return
      startTransition(() => {
        // Only call router.replace if the router API differs from the current URL
        // (it will still replace even if same, but this check reduces unnecessary navigations)
        // router.replace expects a URL (string). This will update app router state.
        router.replace(newUrl)
      })

      // Mark pagination state after scheduling the nav (keeps your existing behavior)
      //setIsPaginating(true)
    },
    [buildUrlForValue, onDebouncedChange, router, setIsPaginating]
  )

  // Debounce effect: schedule URL update + router navigation when user pauses typing.
  useEffect(() => {
    if (typeof window === 'undefined') return

    // If currently composing IME input, don't schedule the debounce now
    if (isComposingRef.current) return

    if (debounceRef.current) clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(() => {
      applyUrlAndTriggerRouter(true)
    }, debounceMs)

    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current)
    }
  }, [debounceMs, applyUrlAndTriggerRouter])

  // onChange wrapper that updates state + ref
  const handleChange = useCallback(
    (next: string) => {
      setValue(next)
      valueRef.current = next

      // Also reflect the change immediately in the URL bar without triggering router navigation.
      // This keeps history visible and shareable while the user types, but doesn't cause useSearchParams updates.
      if (typeof window !== 'undefined') {
        const immediateUrl = buildUrlForValue(next)
        window.history.replaceState({}, '', immediateUrl)
      }
    },
    [buildUrlForValue]
  )

  // Composition handlers for IME input
  const handleCompositionStart = useCallback(() => {
    isComposingRef.current = true
    if (debounceRef.current) clearTimeout(debounceRef.current)
  }, [])

  const handleCompositionEnd = useCallback(() => {
    isComposingRef.current = false
    // When composition ends, apply and trigger router immediately (user expects result).
    // Clear existing debounce then run an immediate short-scheduled update to avoid race.
    if (debounceRef.current) clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(
      () => {
        applyUrlAndTriggerRouter(true)
      },
      Math.max(0, debounceMs)
    )
  }, [applyUrlAndTriggerRouter, debounceMs])

  // Clear handler
  const handleClear = useCallback(() => {
    setValue('')
    valueRef.current = ''
    if (debounceRef.current) clearTimeout(debounceRef.current)

    if (typeof window === 'undefined') return

    const params = new URLSearchParams(window.location.search)
    params.delete(searchParamKey)
    params.delete('offset')
    const query = params.toString()
    const newUrl = `${window.location.pathname}${query ? `?${query}` : ''}${window.location.hash || ''}`

    // Replace history and trigger router so the search results clear
    window.history.replaceState({}, '', newUrl)
    startTransition(() => {
      router.replace(newUrl)
    })

    //setIsPaginating(true)
    onDebouncedChange?.('')
  }, [onDebouncedChange, router, searchParamKey, setIsPaginating])

  return (
    <SearchInput
      className={className}
      placeholder={placeholder}
      value={value}
      onChange={handleChange}
      onClear={handleClear}
      onFocus={() => (isFocusedRef.current = true)}
      onBlur={() => (isFocusedRef.current = false)}
      onCompositionStart={handleCompositionStart}
      onCompositionEnd={handleCompositionEnd}
    />
  )
}
