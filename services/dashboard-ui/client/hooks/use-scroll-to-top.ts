'use client'

import { useCallback } from 'react'

interface IUseScrollToTop {
  behavior?: 'smooth' | 'instant' | 'auto'
}

export const useScrollToTop = (options: IUseScrollToTop = {}) => {
  const { behavior = 'smooth' } = options

  const scrollToTop = useCallback(
    (elementId?: string, offset: number = 0) => {
      // Determine the target element
      let targetElement: HTMLElement | Window

      if (elementId) {
        const element = document.getElementById(elementId)
        if (element) {
          targetElement = element
        } else {
          console.warn(
            `Element with ID "${elementId}" not found, falling back to window`
          )
          targetElement = window
        }
      } else {
        targetElement = window
      }

      // Calculate scroll position with offset
      const scrollPosition = Math.max(0, offset)

      // Scroll to position
      if (targetElement === window) {
        window.scrollTo({
          top: scrollPosition,
          behavior,
        })
      } else {
        ;(targetElement as HTMLElement).scrollTo({
          top: scrollPosition,
          behavior,
        })
      }
    },
    [behavior]
  )

  return scrollToTop
}
