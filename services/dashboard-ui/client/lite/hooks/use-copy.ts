import { useCallback, useEffect, useRef, useState } from 'react'

export type TCopyStatus = 'idle' | 'copied' | 'error'

export interface IUseCopy {
  resetAfter?: number
}

const legacyCopy = (value: string) => {
  const el = document.createElement('textarea')
  el.value = value
  el.setAttribute('readonly', '')
  el.style.position = 'fixed'
  el.style.opacity = '0'
  document.body.appendChild(el)
  el.select()
  try {
    return document.execCommand('copy')
  } catch {
    return false
  } finally {
    el.remove()
  }
}

export const useCopy = ({ resetAfter = 2000 }: IUseCopy = {}) => {
  const [status, setStatus] = useState<TCopyStatus>('idle')
  const timer = useRef<ReturnType<typeof setTimeout>>()

  useEffect(() => () => clearTimeout(timer.current), [])

  const schedule = useCallback(
    (next: TCopyStatus) => {
      setStatus(next)
      clearTimeout(timer.current)
      timer.current = setTimeout(() => setStatus('idle'), resetAfter)
    },
    [resetAfter]
  )

  const reset = useCallback(() => {
    clearTimeout(timer.current)
    setStatus('idle')
  }, [])

  const copy = useCallback(
    async (value: string) => {
      if (!value) {
        schedule('error')
        return false
      }

      try {
        await navigator.clipboard.writeText(value)
        schedule('copied')
        return true
      } catch {
        const ok = legacyCopy(value)
        schedule(ok ? 'copied' : 'error')
        return ok
      }
    },
    [schedule]
  )

  return { copy, status, reset }
}
