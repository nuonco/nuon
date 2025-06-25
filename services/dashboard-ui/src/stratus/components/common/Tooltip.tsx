'use client'

import React, { useEffect, useRef, useState } from 'react'
import { cn } from '@/stratus/components/helpers'
import { Icon } from './Icon'
import './Tooltip.css'

export interface ITooltip extends React.HTMLAttributes<HTMLSpanElement> {
  position?: 'top' | 'bottom' | 'left' | 'right'
  showIcon?: boolean
  tipContent: React.ReactNode
}

export const Tooltip = ({
  className,
  children,
  position = 'top',
  showIcon = false,
  tipContent,
  ...props
}: ITooltip) => {
  const [isOpen, setIsOpen] = useState(false)
  const [styles, setStyles] = useState<{
    top: string
    left: string
  } | null>(null)
  const tooltipRef = useRef<HTMLDivElement>(null)
  const triggerRef = useRef<HTMLDivElement>(null)

  const calculatePosition = () => {
    if (triggerRef.current && tooltipRef.current) {
      const triggerRect = triggerRef.current.getBoundingClientRect()
      const tooltipRect = tooltipRef.current.getBoundingClientRect()

      let top = 0
      let left = 0

      if (position === 'top') {
        top = -(tooltipRect.height + 8) // 8px spacing above
        left = triggerRect.width / 2 - tooltipRect.width / 2
      } else if (position === 'bottom') {
        top = triggerRect.height + 8 // 8px spacing below
        left = triggerRect.width / 2 - tooltipRect.width / 2
      } else if (position === 'left') {
        top = triggerRect.height / 2 - tooltipRect.height / 2
        left = -(tooltipRect.width + 8) // 8px spacing to the left
      } else if (position === 'right') {
        top = triggerRect.height / 2 - tooltipRect.height / 2
        left = triggerRect.width + 8 // 8px spacing to the right
      }

      setStyles({
        top: `${top}px`,
        left: `${left}px`,
      })
    }
  }

  useEffect(() => {
    calculatePosition()

    // Recalculate on window resize or scroll
    window.addEventListener('resize', calculatePosition)
    window.addEventListener('scroll', calculatePosition)
    return () => {
      window.removeEventListener('resize', calculatePosition)
      window.removeEventListener('scroll', calculatePosition)
    }
  }, [])

  return (
    <span
      className={cn('tooltip-wrapper', className)}
      ref={triggerRef}
      style={{ position: 'relative' }}
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
          {children} <Icon variant="Question" />
        </span>
      ) : (
        children
      )}

      <span
        ref={tooltipRef}
        className={cn(`tooltip-content ${position}`, {
          enter: isOpen,
          exit: !isOpen,
        })}
        style={styles}
      >
        {tipContent}
      </span>
    </span>
  )
}
