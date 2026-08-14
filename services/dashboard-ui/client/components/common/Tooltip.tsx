import React, { useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { cn } from '@/utils/classnames'
import { Icon } from './Icon'
import { Text } from './Text'
import './Tooltip.css'

export interface ITooltip extends React.HTMLAttributes<HTMLSpanElement> {
  isOpen?: boolean
  defaultOpen?: boolean
  disableHover?: boolean
  onOpenChange?: (open: boolean) => void
  position?: 'top' | 'bottom' | 'left' | 'right'
  showIcon?: boolean
  tipContent: React.ReactNode
  tipContentClassName?: string
}

export const Tooltip = ({
  className,
  children,
  isOpen: controlledOpen,
  defaultOpen = false,
  disableHover = false,
  onOpenChange,
  position = 'top',
  showIcon = false,
  tipContent,
  tipContentClassName,
  ...props
}: ITooltip) => {
  const [uncontrolledOpen, setUncontrolledOpen] = useState(defaultOpen)
  const isControlled = controlledOpen !== undefined
  const isOpen = isControlled ? controlledOpen : uncontrolledOpen

  const setOpen = (open: boolean) => {
    if (!isControlled) setUncontrolledOpen(open)
    onOpenChange?.(open)
  }
  const [styles, setStyles] = useState<{
    top: number
    left: number
    arrow: number
  } | null>(null)
  const [effPosition, setEffPosition] = useState(position)
  const tooltipRef = useRef<HTMLDivElement>(null)
  const triggerRef = useRef<HTMLDivElement>(null)

  const calculatePosition = () => {
    if (!triggerRef.current || !tooltipRef.current) return

    const trigger = triggerRef.current.getBoundingClientRect()
    const tip = tooltipRef.current.getBoundingClientRect()
    const vw = window.innerWidth
    const vh = window.innerHeight
    const gap = 8
    const margin = 8

    let pos = position
    if (
      pos === 'right' &&
      trigger.right + gap + tip.width > vw - margin &&
      trigger.left - gap - tip.width >= margin
    ) {
      pos = 'left'
    } else if (
      pos === 'left' &&
      trigger.left - gap - tip.width < margin &&
      trigger.right + gap + tip.width <= vw - margin
    ) {
      pos = 'right'
    } else if (
      pos === 'bottom' &&
      trigger.bottom + gap + tip.height > vh - margin &&
      trigger.top - gap - tip.height >= margin
    ) {
      pos = 'top'
    } else if (
      pos === 'top' &&
      trigger.top - gap - tip.height < margin &&
      trigger.bottom + gap + tip.height <= vh - margin
    ) {
      pos = 'bottom'
    }

    let top = 0
    let left = 0
    if (pos === 'top') {
      top = trigger.top - tip.height - gap
      left = trigger.left + trigger.width / 2 - tip.width / 2
    } else if (pos === 'bottom') {
      top = trigger.bottom + gap
      left = trigger.left + trigger.width / 2 - tip.width / 2
    } else if (pos === 'left') {
      top = trigger.top + trigger.height / 2 - tip.height / 2
      left = trigger.left - tip.width - gap
    } else if (pos === 'right') {
      top = trigger.top + trigger.height / 2 - tip.height / 2
      left = trigger.right + gap
    }

    left = Math.min(Math.max(margin, left), vw - tip.width - margin)
    top = Math.min(Math.max(margin, top), vh - tip.height - margin)

    const edge = 12
    const arrow =
      pos === 'top' || pos === 'bottom'
        ? Math.min(
            Math.max(edge, trigger.left + trigger.width / 2 - left),
            tip.width - edge
          )
        : Math.min(
            Math.max(edge, trigger.top + trigger.height / 2 - top),
            tip.height - edge
          )

    setStyles({ top, left, arrow })
    setEffPosition(pos)
  }

  useEffect(() => {
    calculatePosition()

    window.addEventListener('resize', calculatePosition)
    window.addEventListener('scroll', calculatePosition, true)
    return () => {
      window.removeEventListener('resize', calculatePosition)
      window.removeEventListener('scroll', calculatePosition, true)
    }
  }, [])

  useEffect(() => {
    if (!isOpen) return
    let raf = 0
    let prevKey = ''
    let frames = 0
    const settle = () => {
      calculatePosition()
      const t = triggerRef.current?.getBoundingClientRect()
      const key = t ? `${t.top},${t.left},${t.width},${t.height}` : ''
      if (key !== prevKey && frames < 30) {
        prevKey = key
        frames += 1
        raf = requestAnimationFrame(settle)
      }
    }
    settle()
    return () => cancelAnimationFrame(raf)
  }, [isOpen])

  const tooltipContent = (
    <span
      ref={tooltipRef}
      className={cn(
        `tooltip-content bg-background text-foreground fixed flex items-center px-2 py-1 rounded-md drop-shadow-lg w-max whitespace-nowrap ${effPosition}`,
        {
          enter: isOpen,
          exit: !isOpen,
        },
        tipContentClassName
      )}
      role="tooltip"
      style={
        styles
          ? ({
              top: `${styles.top}px`,
              left: `${styles.left}px`,
              '--arrow': `${styles.arrow}px`,
            } as React.CSSProperties)
          : undefined
      }
    >
      {typeof tipContent === 'string' ? (
        <Text variant="subtext">{tipContent}</Text>
      ) : (
        tipContent
      )}
    </span>
  )

  return (
    <span
      className={cn('tooltip-wrapper w-fit leading-none', className)}
      ref={triggerRef}
      onMouseEnter={() => {
        if (disableHover) return
        calculatePosition()
        setOpen(true)
      }}
      onMouseLeave={() => {
        if (disableHover) return
        setOpen(false)
      }}
      onFocus={() => {
        if (disableHover) return
        calculatePosition()
        setOpen(true)
      }}
      onBlur={() => {
        if (disableHover) return
        setOpen(false)
      }}
      {...props}
    >
      {showIcon ? (
        <span className="inline-flex items-center gap-1 mr-1">
          {children} <Icon variant="QuestionIcon" />
        </span>
      ) : (
        children
      )}

      {createPortal(tooltipContent, document.body)}
    </span>
  )
}
