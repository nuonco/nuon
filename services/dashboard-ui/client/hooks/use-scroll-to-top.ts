import { useCallback } from 'react'

interface IUseScrollToTop {
  behavior?: 'smooth' | 'instant' | 'auto'
}

export const useScrollToTop = (options: IUseScrollToTop = {}) => {
  const { behavior = 'smooth' } = options

  const scrollToTop = useCallback(
    (elementId?: string, offset: number = 0) => {
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

      const scrollPosition = Math.max(0, offset)

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
