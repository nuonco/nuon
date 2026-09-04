import {
  createContext,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react'
import type { TStatusTheme } from '@/utils/status-utils'
import { ToastStack } from '../components/organisms/toasts/ToastStack'
import { TOAST_EXIT_MS } from '../components/organisms/toasts/toast-motion'

export const DEFAULT_TOAST_TIMEOUT = 5000

export type TToastTheme = TStatusTheme | 'default'

export interface IToastAction {
  label: string
  onClick: () => void
}

export interface IToastInput {
  heading: string
  description?: ReactNode
  theme?: TToastTheme
  action?: IToastAction
  timeout?: number | null
}

export interface IToastDescriptor
  extends Omit<IToastInput, 'theme' | 'timeout'> {
  id: string
  theme: TToastTheme
  timeout: number | null
  exiting: boolean
}

export interface IToastContext {
  addToast: (toast: IToastInput) => string
  removeToast: (id: string) => void
  clearToasts: () => void
}

interface IToastTimer {
  remaining: number | null
  startedAt: number | null
  timerId: ReturnType<typeof setTimeout> | null
}

export const ToastContext = createContext<IToastContext | null>(null)

const resolveTimeout = (timeout: number | null | undefined) => {
  if (timeout === null) return null
  if (timeout === undefined) return DEFAULT_TOAST_TIMEOUT
  return timeout > 0 ? timeout : DEFAULT_TOAST_TIMEOUT
}

export const ToastProvider = ({ children }: { children: ReactNode }) => {
  const [toasts, setToasts] = useState<IToastDescriptor[]>([])
  const [paused, setPaused] = useState(false)
  const [portalRoot, setPortalRoot] = useState<HTMLDivElement | null>(null)
  const timersRef = useRef(new Map<string, IToastTimer>())
  const exitTimersRef = useRef(new Map<string, ReturnType<typeof setTimeout>>())

  const finalizeToast = useCallback((id: string) => {
    const exitTimer = exitTimersRef.current.get(id)
    if (exitTimer) clearTimeout(exitTimer)
    exitTimersRef.current.delete(id)
    timersRef.current.delete(id)
    setToasts((current) => current.filter((toast) => toast.id !== id))
  }, [])

  const removeToast = useCallback(
    (id: string) => {
      const timer = timersRef.current.get(id)
      if (timer?.timerId) clearTimeout(timer.timerId)
      if (timer) {
        timer.timerId = null
        timer.startedAt = null
      }
      setToasts((current) =>
        current.map((toast) =>
          toast.id === id && !toast.exiting
            ? { ...toast, exiting: true }
            : toast
        )
      )
      if (!exitTimersRef.current.has(id)) {
        const exitTimer = setTimeout(
          () => finalizeToast(id),
          TOAST_EXIT_MS + 40
        )
        exitTimersRef.current.set(id, exitTimer)
      }
    },
    [finalizeToast]
  )

  const addToast = useCallback((input: IToastInput) => {
    const id = crypto.randomUUID()
    const timeout = resolveTimeout(input.timeout)
    timersRef.current.set(id, {
      remaining: timeout,
      startedAt: null,
      timerId: null,
    })
    setToasts((current) => [
      ...current,
      {
        ...input,
        id,
        theme: input.theme ?? 'default',
        timeout,
        exiting: false,
      },
    ])
    return id
  }, [])

  const clearToasts = useCallback(() => {
    for (const toast of toasts) removeToast(toast.id)
  }, [removeToast, toasts])

  const setTimersPaused = useCallback((nextPaused: boolean) => {
    if (nextPaused) {
      const now = Date.now()
      for (const timer of timersRef.current.values()) {
        if (timer.timerId && timer.startedAt !== null) {
          clearTimeout(timer.timerId)
          timer.timerId = null
          timer.remaining =
            timer.remaining === null
              ? null
              : Math.max(0, timer.remaining - (now - timer.startedAt))
          timer.startedAt = null
        }
      }
    }
    setPaused(nextPaused)
  }, [])

  useEffect(() => {
    if (paused) return

    for (const toast of toasts) {
      if (toast.exiting) continue
      const timer = timersRef.current.get(toast.id)
      if (!timer || timer.remaining === null || timer.timerId) continue

      timer.startedAt = Date.now()
      timer.timerId = setTimeout(() => removeToast(toast.id), timer.remaining)
    }
  }, [paused, removeToast, toasts])

  useEffect(
    () => () => {
      for (const timer of timersRef.current.values()) {
        if (timer.timerId) clearTimeout(timer.timerId)
      }
      for (const timer of exitTimersRef.current.values()) clearTimeout(timer)
    },
    []
  )

  const value = useMemo<IToastContext>(
    () => ({ addToast, clearToasts, removeToast }),
    [addToast, clearToasts, removeToast]
  )

  return (
    <ToastContext.Provider value={value}>
      {children}
      <div ref={setPortalRoot} id="lite-toast-root" />
      {portalRoot ? (
        <ToastStack
          portalRoot={portalRoot}
          toasts={toasts}
          onDismiss={removeToast}
          onExitComplete={finalizeToast}
          onPausedChange={setTimersPaused}
        />
      ) : null}
    </ToastContext.Provider>
  )
}
