import { useCallback, useEffect, useRef, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router'
import { SearchInput } from '@/components/common/SearchInput'

const SEARCH_PARAM = 'search'
const DEBOUNCE_MS = 300

export const WorkflowSearch = () => {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const urlValue = searchParams.get(SEARCH_PARAM) || ''
  const [value, setValue] = useState(urlValue)
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  // Keep local input in sync when the URL param changes externally (e.g.
  // navigating to the page with a pre-set search).
  useEffect(() => {
    setValue(urlValue)
  }, [urlValue])

  const writeParam = useCallback(
    (next: string) => {
      const params = new URLSearchParams(searchParams.toString())
      if (next) {
        params.set(SEARCH_PARAM, next)
      } else {
        params.delete(SEARCH_PARAM)
      }
      params.delete('offset')
      navigate(`?${params.toString()}`, { replace: true })
    },
    [navigate, searchParams]
  )

  const handleChange = useCallback(
    (next: string) => {
      setValue(next)
      if (timerRef.current) clearTimeout(timerRef.current)
      timerRef.current = setTimeout(() => writeParam(next), DEBOUNCE_MS)
    },
    [writeParam]
  )

  const handleClear = useCallback(() => {
    setValue('')
    if (timerRef.current) clearTimeout(timerRef.current)
    writeParam('')
  }, [writeParam])

  useEffect(() => {
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current)
    }
  }, [])

  return (
    <SearchInput
      placeholder="Search workflows..."
      value={value}
      onChange={handleChange}
      onClear={handleClear}
    />
  )
}
