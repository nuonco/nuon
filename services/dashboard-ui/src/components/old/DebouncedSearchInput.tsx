'use client'

import { useRouter, useSearchParams } from 'next/navigation'
import React, { type FC, useState, useEffect, useRef } from 'react'
import { MagnifyingGlassIcon, XCircleIcon } from '@phosphor-icons/react'
import { Button } from './Button'

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
  const valFromUrl = searchParams.get(searchParamKey) || ''
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
  }, [value, searchParamKey, router, debounceMs, onDebouncedChange])

  return (
    <label className="relative">
      <MagnifyingGlassIcon className="text-cool-grey-500 dark:text-cool-grey-700 absolute top-2.5 left-2" />
      <input
        className={`rounded-md pl-8 pr-3.5 py-1.5 h-[36px] text-base border bg-white dark:bg-dark-grey-800 placeholder:text-cool-grey-500 dark:placeholder:text-cool-grey-700 md:min-w-80 ${className}`}
        type="text"
        placeholder={placeholder}
        value={value}
        onChange={(e) => setValue(e.target.value)}
        autoComplete="off"
      />
      {value !== '' ? (
        <Button
          className="!p-0.5 absolute top-1/2 right-1.5 -translate-y-1/2"
          variant="ghost"
          title="clear search"
          value=""
          onClick={(e) => setValue((e.target as HTMLButtonElement).value)}
        >
          <XCircleIcon />
        </Button>
      ) : null}
    </label>
  )
}
