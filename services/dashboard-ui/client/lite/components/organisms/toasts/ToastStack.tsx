import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
  type CSSProperties,
} from 'react'
import { createPortal } from 'react-dom'
import { cn } from '@/utils/classnames'
import type { IToastDescriptor } from '../../../providers/toast-provider'
import {
  TOAST_ENTER_EASING,
  TOAST_ENTER_MS,
  TOAST_EXIT_EASING,
  TOAST_EXIT_MS,
  TOAST_REFLOW_EASING,
  TOAST_REFLOW_MS,
} from './toast-motion'
import { Card } from '../../atoms/Card'
import { Toast } from './Toast'

export interface IToastStack {
  portalRoot: HTMLElement
  toasts: IToastDescriptor[]
  onDismiss: (id: string) => void
  onExitComplete: (id: string) => void
  onPausedChange: (paused: boolean) => void
}

interface IToastSlot {
  descriptor: IToastDescriptor
  expanded: boolean
  offset: number
  opacity: number
  scale: number
  contentVisible: boolean
  zIndex: number
  onDismiss: (id: string) => void
  onExitComplete: (id: string) => void
  setElement: (id: string, element: HTMLDivElement | null) => void
}

const STACK_GAP = 12
const COMPACT_OFFSET = 10
const FALLBACK_HEIGHT = 80

const prefersReducedMotion = () =>
  typeof window !== 'undefined' &&
  window.matchMedia?.('(prefers-reduced-motion: reduce)').matches

const ToastSlot = ({
  descriptor,
  expanded,
  offset,
  opacity,
  scale,
  contentVisible,
  zIndex,
  onDismiss,
  onExitComplete,
  setElement,
}: IToastSlot) => {
  const [entered, setEntered] = useState(false)
  const finalizedRef = useRef(false)
  const slotRef = useRef<HTMLDivElement | null>(null)
  const interactive = expanded || (opacity === 1 && scale === 1)

  useEffect(() => {
    const frame = requestAnimationFrame(() => setEntered(true))
    return () => cancelAnimationFrame(frame)
  }, [])

  useEffect(() => {
    if (!descriptor.exiting) return
    setEntered(false)

    if (prefersReducedMotion()) {
      onExitComplete(descriptor.id)
    }
  }, [descriptor.exiting, descriptor.id, onExitComplete])

  useLayoutEffect(() => {
    if (slotRef.current) slotRef.current.inert = !interactive
  }, [interactive])

  const finishExit = () => {
    if (!descriptor.exiting || finalizedRef.current) return
    finalizedRef.current = true
    onExitComplete(descriptor.id)
  }

  const runAction = () => {
    try {
      descriptor.action?.onClick()
    } finally {
      onDismiss(descriptor.id)
    }
  }

  return (
    <div
      ref={(element) => {
        slotRef.current = element
        setElement(descriptor.id, element)
      }}
      data-toast-slot
      data-toast-id={descriptor.id}
      data-state={
        descriptor.exiting ? 'exiting' : entered ? 'open' : 'entering'
      }
      aria-hidden={!interactive || undefined}
      className="absolute inset-x-0 bottom-0 origin-bottom motion-reduce:transform-none motion-reduce:transition-none"
      style={{
        zIndex,
        opacity,
        pointerEvents: interactive && !descriptor.exiting ? 'auto' : 'none',
        transform: `translate3d(0, ${offset}px, 0) scale(${scale})`,
        transition: `transform ${TOAST_REFLOW_MS}ms ${TOAST_REFLOW_EASING}, opacity ${TOAST_REFLOW_MS}ms ${TOAST_REFLOW_EASING}`,
        willChange: 'transform, opacity',
      }}
    >
      <div
        className="motion-reduce:transform-none motion-reduce:opacity-100 motion-reduce:transition-none"
        style={
          {
            opacity: entered && !descriptor.exiting ? 1 : 0,
            translate: entered && !descriptor.exiting ? '0 0' : '0 0.75rem',
            scale: entered && !descriptor.exiting ? 1 : 0.98,
            transitionProperty: 'opacity, translate, scale',
            transitionDuration: `${
              entered && !descriptor.exiting ? TOAST_ENTER_MS : TOAST_EXIT_MS
            }ms`,
            transitionTimingFunction:
              entered && !descriptor.exiting
                ? TOAST_ENTER_EASING
                : TOAST_EXIT_EASING,
            willChange: 'opacity, translate, scale',
          } as CSSProperties
        }
        onTransitionEnd={(event) => {
          if (
            event.target === event.currentTarget &&
            event.propertyName === 'opacity'
          ) {
            finishExit()
          }
        }}
      >
        <div className="relative">
          <div
            className="transition-opacity duration-150 motion-reduce:transition-none"
            style={{ opacity: contentVisible ? 1 : 0 }}
          >
            <Toast
              heading={descriptor.heading}
              description={descriptor.description}
              theme={descriptor.theme}
              actionLabel={descriptor.action?.label}
              onAction={descriptor.action ? runAction : undefined}
              onDismiss={() => onDismiss(descriptor.id)}
            />
          </div>
          <Card
            aria-hidden
            padding="none"
            blur="lg"
            opacity="strong"
            shadow="floating"
            className="pointer-events-none absolute inset-0 transition-opacity duration-150 motion-reduce:transition-none"
            style={
              {
                opacity: contentVisible ? 0 : 1,
                '--card-shadow-floating': 'var(--toast-shadow)',
              } as CSSProperties
            }
          />
        </div>
      </div>
    </div>
  )
}

export const ToastStack = ({
  portalRoot,
  toasts,
  onDismiss,
  onExitComplete,
  onPausedChange,
}: IToastStack) => {
  const [hovered, setHovered] = useState(false)
  const [focused, setFocused] = useState(false)
  const [heights, setHeights] = useState<Record<string, number>>({})
  const [scrollable, setScrollable] = useState(false)
  const elementsRef = useRef(new Map<string, HTMLDivElement>())
  const stackRef = useRef<HTMLDivElement>(null)
  const viewportRef = useRef<HTMLDivElement>(null)
  const expanded = hovered || focused

  const setElement = useCallback(
    (id: string, element: HTMLDivElement | null) => {
      if (element) elementsRef.current.set(id, element)
      else elementsRef.current.delete(id)
    },
    []
  )

  useLayoutEffect(() => {
    const measure = () => {
      const next: Record<string, number> = {}
      for (const toast of toasts) {
        const element = elementsRef.current.get(toast.id)
        next[toast.id] =
          element?.offsetHeight ||
          element?.getBoundingClientRect().height ||
          FALLBACK_HEIGHT
      }
      setHeights((current) => {
        const ids = Object.keys(next)
        const unchanged =
          ids.length === Object.keys(current).length &&
          ids.every((id) => current[id] === next[id])
        return unchanged ? current : next
      })
    }

    measure()
    if (typeof ResizeObserver === 'undefined') return

    const observer = new ResizeObserver(measure)
    for (const element of elementsRef.current.values())
      observer.observe(element)
    return () => observer.disconnect()
  }, [toasts])

  useEffect(() => {
    onPausedChange(expanded)
    return () => onPausedChange(false)
  }, [expanded, onPausedChange])

  useLayoutEffect(() => {
    const activeElement = document.activeElement
    const activeSlot =
      activeElement instanceof Element
        ? activeElement.closest<HTMLElement>('[data-toast-slot]')
        : null
    if (
      focused &&
      stackRef.current &&
      (!stackRef.current.contains(activeElement) ||
        activeSlot?.dataset.state === 'exiting')
    ) {
      setFocused(false)
    }
  }, [focused, toasts])

  const layout = useMemo(() => {
    const heightOf = (toast: IToastDescriptor) =>
      heights[toast.id] ?? FALLBACK_HEIGHT
    const totalHeight = toasts.reduce(
      (total, toast, index) =>
        total + heightOf(toast) + (index ? STACK_GAP : 0),
      0
    )
    const frontHeight = toasts.length ? heightOf(toasts[toasts.length - 1]) : 0
    const compactHeight = frontHeight
      ? frontHeight + Math.min(2, toasts.length - 1) * COMPACT_OFFSET
      : 0

    return {
      height: expanded ? totalHeight : compactHeight,
      slots: toasts.map((toast, index) => {
        const depth = toasts.length - 1 - index
        if (expanded) {
          const offset = -toasts
            .slice(index + 1)
            .reduce((total, item) => total + heightOf(item) + STACK_GAP, 0)
          return { offset, opacity: 1, scale: 1 }
        }
        const visibleDepth = Math.min(depth, 2)
        return {
          offset: -visibleDepth * COMPACT_OFFSET,
          opacity: depth > 2 ? 0 : 1,
          scale: 1 - visibleDepth * 0.04,
        }
      }),
    }
  }, [expanded, heights, toasts])

  useLayoutEffect(() => {
    const update = () => {
      const desktop = window.matchMedia('(min-width: 640px)').matches
      const availableHeight = window.innerHeight - (desktop ? 96 : 32)
      setScrollable(expanded && layout.height > availableHeight)
    }
    update()
    window.addEventListener('resize', update)
    return () => window.removeEventListener('resize', update)
  }, [expanded, layout.height])

  return createPortal(
    <div
      ref={stackRef}
      data-toast-stack
      data-expanded={expanded || undefined}
      aria-label="Notifications"
      className="pointer-events-none fixed inset-x-4 bottom-[calc(1rem+env(safe-area-inset-bottom))] z-[1000] sm:left-auto sm:right-6 sm:bottom-20 sm:w-96"
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      onFocusCapture={() => setFocused(true)}
      onBlurCapture={(event) => {
        if (!event.currentTarget.contains(event.relatedTarget as Node | null)) {
          setFocused(false)
        }
      }}
    >
      <div
        ref={viewportRef}
        className={cn(
          'max-h-[calc(100dvh-2rem-env(safe-area-inset-bottom))] sm:max-h-[calc(100dvh-6rem)]',
          scrollable
            ? '-m-4 overflow-y-auto overscroll-contain p-4'
            : 'overflow-visible'
        )}
      >
        <div
          className="pointer-events-auto relative transition-[height] motion-reduce:transition-none"
          style={{
            height: layout.height,
            transitionDuration: `${TOAST_REFLOW_MS}ms`,
            transitionTimingFunction: TOAST_REFLOW_EASING,
          }}
        >
          {toasts.map((toast, index) => (
            <ToastSlot
              key={toast.id}
              descriptor={toast}
              expanded={expanded}
              offset={layout.slots[index].offset}
              opacity={layout.slots[index].opacity}
              scale={layout.slots[index].scale}
              contentVisible={expanded || index === toasts.length - 1}
              zIndex={index + 1}
              onDismiss={onDismiss}
              onExitComplete={onExitComplete}
              setElement={setElement}
            />
          ))}
        </div>
      </div>
    </div>,
    portalRoot
  )
}
