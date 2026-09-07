import { useEffect, useState } from 'react'
import { cn } from '@/utils/classnames'
import {
  SURFACE_ENTER_EASING,
  SURFACE_EXIT_EASING,
  SURFACE_TRANSITION_MS,
} from '../../../lib/surface-motion'

export interface ISurfaceOverlay {
  visible: boolean
  topmost: boolean
  onClose?: () => void
}

export const SurfaceOverlay = ({
  visible,
  topmost,
  onClose,
}: ISurfaceOverlay) => {
  const [entered, setEntered] = useState(false)

  useEffect(() => {
    if (!visible || !topmost) {
      setEntered(false)
      return
    }
    const frame = requestAnimationFrame(() => setEntered(true))
    return () => cancelAnimationFrame(frame)
  }, [topmost, visible])

  const active = visible && entered

  return topmost ? (
    <div
      aria-hidden="true"
      data-surface-overlay
      className={cn(
        'pointer-events-auto absolute inset-0 bg-[var(--surface-overlay)] will-change-[opacity,backdrop-filter] transition-[opacity,backdrop-filter] motion-reduce:transition-none',
        active ? 'opacity-100 backdrop-blur-sm' : 'opacity-0 backdrop-blur-none'
      )}
      style={{
        transitionDuration: `${SURFACE_TRANSITION_MS}ms`,
        transitionTimingFunction: active
          ? SURFACE_ENTER_EASING
          : SURFACE_EXIT_EASING,
      }}
      onClick={onClose}
    />
  ) : null
}
