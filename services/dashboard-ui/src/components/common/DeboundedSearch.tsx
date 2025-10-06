'use client'

import { useRouter, useSearchParams } from 'next/navigation'
import { type FC, useState, useEffect, useRef } from 'react'
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

export const DebouncedSearchInput: FC<IDebouncedSearchInput> = ({
  searchParamKey = 'q',
  initialValue,
  placeholder = 'Search…',
  debounceMs = 200,
  className,
  onDebouncedChange,
}) => {
  const router = useRouter()
  const searchParams = useSearchParams()
  const { setIsPaginating } = usePagination()
  const valFromUrl = searchParams?.get(searchParamKey) || ''
  const [value, setValue] = useState(initialValue ?? valFromUrl)
  const debounceRef = useRef<NodeJS.Timeout | null>(null)

  // Keep in sync if URL changes outside this component
  useEffect(() => {
    setValue(initialValue ?? valFromUrl)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [valFromUrl])

  useEffect(() => {
    if (debounceRef.current) clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(() => {
      const params = new URLSearchParams(window.location.search)
      if (value) {
        setIsPaginating(true)
        params.set(searchParamKey, value)
      } else {
        params.delete(searchParamKey)
      }
      router.replace(`?${params.toString()}`)

      onDebouncedChange?.(value)
    }, debounceMs)

    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current)
    }
  }, [
    value,
    searchParamKey,
    router,
    debounceMs,
    onDebouncedChange,
    setIsPaginating,
  ])

  // --- Snappy clear handler ---
  const handleClear = () => {
    setValue('')
    if (debounceRef.current) clearTimeout(debounceRef.current)

    // Preserve all existing params except this one
    const params = new URLSearchParams(window.location.search)
    params.delete(searchParamKey)

    setIsPaginating(true)
    router.replace(`?${params.toString()}`)

    onDebouncedChange?.('')
  }

  return (
    <SearchInput
      className={className}
      placeholder={placeholder}
      value={value}
      onChange={setValue}
      onClear={handleClear}
    />
  )
}
