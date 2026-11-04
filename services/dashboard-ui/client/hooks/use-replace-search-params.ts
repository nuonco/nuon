import { useCallback, useEffect, useRef } from 'react'
import { useSearchParams } from 'react-router'

// setSearchParams does not update window.location until its navigation
// finishes, and the functional updater only sees the last render. A log
// filter mounting in that gap would replace a panel key that was just
// queued. pending is that in-flight write; pendingBase is the address bar
// it started from, so a render that still shows the old bar does not
// discard it.
let pending: URLSearchParams | null = null
let pendingBase: string | null = null

export const useReplaceSearchParams = () => {
  const [searchParams, setSearchParams] = useSearchParams()
  const setRef = useRef(setSearchParams)
  setRef.current = setSearchParams

  useEffect(() => {
    if (!pending) return
    const current = window.location.search
    if (pending.toString() === new URLSearchParams(current).toString()) {
      pending = null
      pendingBase = null
      return
    }
    if (pendingBase !== null && current !== pendingBase) {
      pending = null
      pendingBase = null
    }
  }, [searchParams])

  return useCallback((update: (params: URLSearchParams) => void) => {
    if (!pending) pendingBase = window.location.search
    const next = new URLSearchParams(pending ?? window.location.search)
    update(next)
    pending = next
    setRef.current(next, { replace: true })
  }, [])
}
