'use client'

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
        setIsMounted(true) // Mount the component
        setIsExiting(false) // Remove the exit class
      } else {
        setIsExiting(true) // Add the exit class
        const timeout = setTimeout(() => {
          setIsMounted(false) // Unmount the component after the animation
          onExited?.() // Notify parent that the component has exited
        }, 155) // Duration should match CSS animation duration

        return () => clearTimeout(timeout) // Cleanup timeout on unmount
      }
    }, [isVisible, onExited])

    if (!isMounted) {
      return null // Don't render anything if the component is not mounted
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
