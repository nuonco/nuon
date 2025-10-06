'use client'

import { useEffect } from 'react'

interface IUseArrowKeysProps {
  onUpArrow: () => void
  onDownArrow: () => void
  enabled?: boolean
}

export const useArrowKeys = ({
  onUpArrow,
  onDownArrow,
  enabled = true,
}: IUseArrowKeysProps) => {
  useEffect(() => {
    if (!enabled) return

    const handleKeyDown = (event: KeyboardEvent) => {
      // Check if the event is from an input field to avoid conflicts
      const target = event.target as HTMLElement
      const isInputField =
        target.tagName === 'INPUT' ||
        target.tagName === 'TEXTAREA' ||
        target.contentEditable === 'true'

      // Skip if user is typing in an input field
      if (isInputField) return

      switch (event.key) {
        case 'ArrowUp':
          event.preventDefault() // Prevent default scroll behavior
          onUpArrow()
          break
        case 'ArrowDown':
          event.preventDefault() // Prevent default scroll behavior
          onDownArrow()
          break
        default:
          break
      }
    }

    // Add event listener to document for global key handling
    document.addEventListener('keydown', handleKeyDown)

    // Cleanup function
    return () => {
      document.removeEventListener('keydown', handleKeyDown)
    }
  }, [onUpArrow, onDownArrow, enabled])
}
