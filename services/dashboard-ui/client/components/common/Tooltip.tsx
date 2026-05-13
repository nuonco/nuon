import React, { useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { cn } from '@/utils/classnames'
import { Icon } from './Icon'
import './Tooltip.css'

export interface ITooltip extends React.HTMLAttributes<HTMLSpanElement> {
  isOpen?: boolean
  position?: 'top' | 'bottom' | 'left' | 'right'
  showIcon?: boolean
  tipContent: React.ReactNode
  tipContentClassName?: string
}

export const Tooltip = ({
  className,
  children,
  isOpen: initIsOpen = false,
  position = 'top',
  showIcon = false,
  tipContent,
  tipContentClassName,
  ...props
}: ITooltip) => {
  const [isOpen, setIsOpen] = useState(initIsOpen)
  const [styles, setStyles] = useState<{
    top: string
    left: string
  } | null>(null)
  const tooltipRef = useRef<HTMLDivElement>(null)
  const triggerRef = useRef<HTMLDivElement>(null)

  const calculatePosition = () => {
    if (triggerRef.current && tooltipRef.current) {
      const trigger = triggerRef.current.getBoundingClientRect()
      const tooltipRect = tooltipRef.current.getBoundingClientRect()

      let top = 0
      let left = 0

      if (position === 'top') {
        top = trigger.top - tooltipRect.height - 8
        left = trigger.left + trigger.width / 2 - tooltipRect.width / 2
      } else if (position === 'bottom') {
        top = trigger.bottom + 8
        left = trigger.left + trigger.width / 2 - tooltipRect.width / 2
      } else if (position === 'left') {
        top = trigger.top + trigger.height / 2 - tooltipRect.height / 2
        left = trigger.left - tooltipRect.width - 8
      } else if (position === 'right') {
        top = trigger.top + trigger.height / 2 - tooltipRect.height / 2
        left = trigger.right + 8
      }

      setStyles({
        top: `${top}px`,
        left: `${left}px`,
      })
    }
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

  const tooltipContent = (
    <span
      ref={tooltipRef}
      className={cn(
        `tooltip-content bg-background text-foreground fixed block px-2 py-1 rounded-md drop-shadow-lg w-max whitespace-nowrap ${position}`,
        {
          enter: isOpen,
          exit: !isOpen,
        },
        tipContentClassName
      )}
      role="tooltip"
      style={styles || undefined}
    >
      {tipContent}
    </span>
  )

  return (
    <span
      className={cn('tooltip-wrapper w-fit leading-none', className)}
      ref={triggerRef}
      onMouseEnter={() => {
        calculatePosition()
        setIsOpen(true)
      }}
      onMouseLeave={() => {
        setIsOpen(false)
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
