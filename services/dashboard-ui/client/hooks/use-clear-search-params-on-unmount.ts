import { useEffect, useRef } from 'react'
import { useSearchParams } from 'react-router'

const owners = new Map<string, number>()

export const useClearSearchParamsOnUnmount = (keys: readonly string[]) => {
  const [, setSearchParams] = useSearchParams()
  const setRef = useRef(setSearchParams)
  setRef.current = setSearchParams
  const keyList = keys.join(',')

  useEffect(() => {
    const owned = keyList.split(',').filter(Boolean)
    const mountedPathname = window.location.pathname
    owned.forEach((key) => owners.set(key, (owners.get(key) ?? 0) + 1))

    return () => {
      owned.forEach((key) => owners.set(key, (owners.get(key) ?? 1) - 1))
      // Deferred so a StrictMode remount or a sibling owner can claim the key
      // first, and so a panel close can drop its own param before this
      // snapshot. The setter's functional form sees this component's last
      // render, not the live URL.
      setTimeout(() => {
        if (window.location.pathname !== mountedPathname) return
        const next = new URLSearchParams(window.location.search)
        const orphaned = owned.filter(
          (key) => !owners.get(key) && next.has(key)
        )
        if (!orphaned.length) return
        orphaned.forEach((key) => next.delete(key))
        setRef.current(next, { replace: true })
      }, 300)
    }
  }, [keyList])
}
