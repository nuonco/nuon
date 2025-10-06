'use client'

import { v4 as uuid } from 'uuid'
import React, {
  useEffect,
  createContext,
  useLayoutEffect,
  useRef,
  useState,
} from 'react'
import { createPortal } from 'react-dom'
import { TransitionDiv } from '@/components/common/TransitionDiv'
import { type IToast } from '@/components/surfaces/Toast'

type TToast = React.ReactElement<IToast & { ref?: React.Ref<HTMLDivElement> }>

type TToasts = Array<{
  id: string
  content: TToast
  isVisible: boolean
}>

interface IToastContext {
  addToast: (toast: TToast) => void
  removeToast: (id: string) => void
}

export const ToastContext = createContext<IToastContext>({
  addToast: () => {
    throw new Error('addToast called outside ToastProvider')
  },
  removeToast: () => {
    throw new Error('removeToast called outside ToastProvider')
  },
})

export const ToastProvider = ({ children }: { children: React.ReactNode }) => {
  const [toasts, setToasts] = useState<TToasts>([])

  function addToast(content: TToast) {
    setToasts([...toasts, { id: uuid(), content, isVisible: true }])
  }

  function removeToast(id: string) {
    setToasts((ts) =>
      ts.map((t) => (t?.id === id ? { ...t, isVisible: false } : t))
    )

    setTimeout(() => {
      setToasts((ts) => ts.filter((t) => t?.id !== id))
    }, 160)
  }

  return (
    <ToastContext.Provider
      value={{
        addToast,
        removeToast,
      }}
    >
      {children}
      <ToastPortal toasts={toasts} />
    </ToastContext.Provider>
  )
}

const ToastPortal = ({ toasts }: { toasts: TToasts }) => {
  const [mounted, setMounted] = useState(false)
  const [pauseTimeout, setPauseTimeout] = useState(false)
  const [toastHeights, setToastHeights] = useState<number[]>([])
  const containerRef = useRef<HTMLDivElement>(null)
  const toastsRef = useRef<(HTMLDivElement | null)[]>([])
  const GAP = 24

  useEffect(() => {
    setMounted(true)
  }, [])

  useLayoutEffect(() => {
    if (!mounted) return

    const heights = toastsRef.current
      .map((el) => (el ? el.offsetHeight : 0))
      .filter((h) => h > 0)
    setToastHeights(heights)

    const last3 = heights.slice(-3)
    const hoverHeight =
      last3.reduce((sum, h) => sum + h, 0) +
      GAP * (last3.length > 1 ? last3.length - 1 : 0)

    // --height: all toasts + GAP*2 for every toast beyond 3
    let height = 0
    if (last3.length > 0) {
      height =
        last3[last3.length - 1] +
        (last3.length > 1 ? GAP : 0) +
        (last3.length > 2 ? GAP : 0)
    }

    containerRef.current.style.setProperty('--hover-height', `${hoverHeight}px`)
    containerRef.current.style.setProperty('--height', `${height}px`)
  }, [toasts, mounted])

  if (!mounted) return null

  const len = toasts?.length
  const capIndex = 3

  return createPortal(
    <div
      className="fixed bottom-6 right-6 z-20 w-full max-w-72 block overflow-visible hover:bottom-10"
      id="toast-portal"
      onMouseEnter={() => {
        setPauseTimeout(true)
      }}
      onMouseLeave={() => {
        setPauseTimeout(false)
      }}
      ref={containerRef}
    >
      {toasts.map((t, i) => {
        const index = len - i
        const effectiveIndex = index > capIndex ? capIndex : index - 1
        const prevHeight = toastHeights
          .slice(len - effectiveIndex)
          .reduce((sum, h) => sum + h, 0)

        return (
          <TransitionDiv
            className={`toast-wrapper toast-wrapper-${index >= 4 ? 4 : index}`}
            key={t.id}
            isVisible={t.isVisible}
            style={
              {
                '--hover-offset-y': `-${prevHeight}px`,
                '--index': index >= 4 ? 4 : index,
              } as React.CSSProperties
            }
          >
            {React.isValidElement(t.content)
              ? React.cloneElement(t.content, {
                  pauseTimeout,
                  toastId: t.id,
                  ref: (el: HTMLDivElement) => {
                    toastsRef.current[i] = el
                  },
                })
              : null}
          </TransitionDiv>
        )
      })}
    </div>,
    document.body
  )
}
