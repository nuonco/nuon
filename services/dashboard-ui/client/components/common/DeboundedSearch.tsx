import { useNavigate, useSearchParams } from 'react-router'
import { useState, useEffect, useRef } from 'react'
import { SearchInput } from './SearchInput'

interface IDebouncedSearchInput {
  searchParamKey?: string
  placeholder?: string
  debounceMs?: number
  className?: string
  labelClassName?: string
}

export const DebouncedSearchInput = ({
  searchParamKey = 'q',
  placeholder = 'Search…',
  debounceMs = 300,
  className,
  labelClassName,
}: IDebouncedSearchInput) => {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const [value, setValue] = useState(searchParams.get(searchParamKey) ?? '')
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const isMounted = useRef(false)

  useEffect(() => {
    if (!isMounted.current) {
      isMounted.current = true
      return
    }
    if (debounceRef.current) clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(() => {
      const params = new URLSearchParams(window.location.search)
      if (value) {
        params.set(searchParamKey, value)
      } else {
        params.delete(searchParamKey)
      }
      params.delete('offset')
      const query = params.toString()
      navigate(`${window.location.pathname}${query ? `?${query}` : ''}`, { replace: true })
    }, debounceMs)

    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current)
    }
  }, [value, debounceMs, navigate, searchParamKey])

  return (
    <SearchInput
      className={className}
      labelClassName={labelClassName}
      placeholder={placeholder}
      value={value}
      onChange={setValue}
    />
  )
}
