import { useEffect, useState, type HTMLAttributes } from 'react'
import { cn } from '@/utils/classnames'
import {
  SURFACE_ENTER_EASING,
  SURFACE_ENTER_MS,
  SURFACE_EXIT_EASING,
  SURFACE_TRANSITION_MS,
} from '../../../lib/surface-motion'

export interface ISurfaceTransition extends HTMLAttributes<HTMLDivElement> {
  coveredBy?: number
  variant: 'modal' | 'panel'
  visible: boolean
}

const panelPositionClass = (coveredBy: number) => {
  if (coveredBy >= 3) return '-translate-x-9'
  if (coveredBy === 2) return '-translate-x-6'
  if (coveredBy === 1) return '-translate-x-3'
  return 'translate-x-0'
}

const transitionStateClass = ({
  coveredBy,
  variant,
  visible,
}: Pick<ISurfaceTransition, 'coveredBy' | 'variant' | 'visible'>) => {
  if (!visible) {
    return variant === 'modal'
      ? 'translate-y-4 opacity-0'
      : 'translate-x-20 opacity-0'
  }
  if (variant === 'modal') {
    return coveredBy
      ? '-translate-y-1 opacity-100'
      : 'translate-y-0 opacity-100'
  }
  return `${panelPositionClass(coveredBy ?? 0)} opacity-100`
}

export const SurfaceTransition = ({
  children,
  className,
  coveredBy = 0,
  style,
  variant,
  visible,
  ...props
}: ISurfaceTransition) => {
  const [entered, setEntered] = useState(false)

  useEffect(() => {
    if (!visible) {
      setEntered(false)
      return
    }
    const frame = requestAnimationFrame(() => setEntered(true))
    return () => cancelAnimationFrame(frame)
  }, [visible])

  const active = visible && entered

  return (
    <div
      className={cn(
        'pointer-events-auto relative will-change-[opacity,translate] transition-[opacity,translate,width] motion-reduce:transition-none motion-reduce:translate-none',
        transitionStateClass({ coveredBy, variant, visible: active }),
        className
      )}
      style={{
        ...style,
        transitionDuration: `${
          active ? SURFACE_ENTER_MS : SURFACE_TRANSITION_MS
        }ms`,
        transitionTimingFunction: active
          ? SURFACE_ENTER_EASING
          : SURFACE_EXIT_EASING,
      }}
      {...props}
    >
      {children}
    </div>
  )
}
