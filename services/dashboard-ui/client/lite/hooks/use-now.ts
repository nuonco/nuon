import { useEffect, useState } from 'react'

export const useNow = (enabled: boolean, intervalMs = 30_000) => {
  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    if (!enabled) return
    const interval = window.setInterval(() => setNow(Date.now()), intervalMs)
    return () => window.clearInterval(interval)
  }, [enabled, intervalMs])

  return now
}
