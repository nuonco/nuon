'use client'

import { v4 as uuid } from 'uuid'
import React, {
  type FC,
  useEffect,
  createContext,
  useContext,
  useLayoutEffect,
  useRef,
  useState,
} from 'react'
import { createPortal } from 'react-dom'
import { TransitionDiv } from '@/stratus/components/common'
import { type IToast } from '@/stratus/components/surfaces'
import { setDashboardSidebarCookie } from '@/stratus/actions'

type TDashboardToasts = Array<{
  id: string
  content: React.ReactNode
  isVisible: boolean
}>

interface IDashboardContext {
  isSidebarOpen?: boolean
  closeSidebar?: () => void
  openSidebar?: () => void
  toggleSidebar?: () => void
  addToast?: (toast: React.ReactNode) => void
  removeToast?: (id: string) => void
}

const DashboardContext = createContext<IDashboardContext>({})

export const DashboardProvider: FC<{
  children: React.ReactNode
  initIsSidebarOpen?: boolean
}> = ({ children, initIsSidebarOpen = false }) => {
  const [isSidebarOpen, setIsSidebarOpen] = useState(initIsSidebarOpen)
  const [toasts, setToasts] = useState<TDashboardToasts>([])

  function addToast(content: React.ReactNode) {
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

  function closeSidebar() {
    setDashboardSidebarCookie(false)
    setIsSidebarOpen(false)
  }

  function openSidebar() {
    setDashboardSidebarCookie(true)
    setIsSidebarOpen(true)
  }

  function toggleSidebar() {
    setDashboardSidebarCookie(!isSidebarOpen)
    setIsSidebarOpen(!isSidebarOpen)
  }

  return (
    <DashboardContext.Provider
      value={{
        addToast,
        closeSidebar,
        isSidebarOpen,
        openSidebar,
        removeToast,
        toggleSidebar,
      }}
    >
      {children}
      <ToastPortal toasts={toasts} />
    </DashboardContext.Provider>
  )
}

export const useDashboard = (): IDashboardContext =>
  useContext(DashboardContext)

const ToastPortal: FC<{
  toasts: TDashboardToasts
}> = ({ toasts }) => {
  const [mounted, setMounted] = useState(false)
  const [toastHeights, setToastHeights] = useState<number[]>([])
  const containerRef = useRef<HTMLDivElement | null>(null)
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
    <div id="toast-portal" ref={containerRef}>
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
            style={{
              // @ts-ignore for custom CSS variables
              '--hover-offset-y': `-${prevHeight}px`,
              '--index': index >= 4 ? 4 : index,
            }}
          >
            {React.isValidElement<IToast>(t.content)
              ? React.cloneElement<IToast>(t.content, {
                  toastId: t.id,
                  ref: (el) => (toastsRef.current[i] = el),
                })
              : null}
          </TransitionDiv>
        )
      })}
    </div>,
    document.body
  )
}
