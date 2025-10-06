'use client'

import { useEffect } from 'react'

export function useAutoFocusOnVisible(
  ref: React.RefObject<HTMLElement>,
  isVisible: boolean,
  delay = 155
) {
  useEffect(() => {
    if (isVisible) {
      const timeout = setTimeout(() => {
        ref?.current?.focus()
      }, delay)
      return () => clearTimeout(timeout)
    }
  }, [isVisible, ref, delay])
}
