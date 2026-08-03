import React, { forwardRef, useEffect, useState } from 'react'
import { cn } from '@/utils/classnames'

export interface ITransitionDiv extends React.HTMLAttributes<HTMLDivElement> {
  isVisible: boolean
  onExited?: () => void
}

export const TransitionDiv = forwardRef<HTMLDivElement, ITransitionDiv>(
  ({ children, className, isVisible, onExited, ...props }, ref) => {
    const [isExiting, setIsExiting] = useState(false)
    const [isMounted, setIsMounted] = useState(isVisible)

    useEffect(() => {
      if (isVisible) {
        setIsMounted(true)
        setIsExiting(false)
      } else {
        setIsExiting(true)
        const timeout = setTimeout(() => {
          setIsMounted(false)
          onExited?.()
        }, 155) // Duration should match CSS animation duration

        return () => clearTimeout(timeout)
      }
    }, [isVisible, onExited])

    if (!isMounted) {
      return null
    }

    return (
      <div
        className={cn(`${isExiting ? 'exit' : 'enter'}`, className)}
        ref={ref}
        {...props}
      >
        {children}
      </div>
    )
  }
)

TransitionDiv.displayName = 'TransitionDiv'
