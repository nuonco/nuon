import { useEffect, useRef, useState, type ReactNode } from 'react'
import { cn } from '@/utils/classnames'

export interface IAnimatedHeight {
  children: ReactNode
  className?: string
}

export const AnimatedHeight = ({ children, className }: IAnimatedHeight) => {
  const innerRef = useRef<HTMLDivElement>(null)
  const [height, setHeight] = useState<number>()

  useEffect(() => {
    const el = innerRef.current
    if (!el) return
    setHeight(el.offsetHeight)
    const observer = new ResizeObserver(() => setHeight(el.offsetHeight))
    observer.observe(el)
    return () => observer.disconnect()
  }, [])

  return (
    <div
      style={{ height }}
      className={cn(
        'overflow-hidden transition-[height] duration-(--duration-fast) ease-cubic motion-reduce:transition-none',
        className,
      )}
    >
      <div ref={innerRef}>{children}</div>
    </div>
  )
}
