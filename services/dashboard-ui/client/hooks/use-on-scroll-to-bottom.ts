import { useEffect, useRef, useState } from 'react'

interface IUseScrollToBottom {
  onScrollToBottom: () => void
  threshold?: number
}

export function useScrollToBottom({
  onScrollToBottom,
  threshold = 10,
}: IUseScrollToBottom) {
  const elementRef = useRef<HTMLDivElement>(null)
  const [hasTriggered, setHasTriggered] = useState(false)

  useEffect(() => {
    const element = elementRef.current
    if (!element) return

    const handleScroll = () => {
      // Don't trigger if already triggered
      if (hasTriggered) return

      const { scrollTop, scrollHeight, clientHeight } = element
      const isAtBottom = scrollTop + clientHeight >= scrollHeight - threshold

      if (isAtBottom) {
        setHasTriggered(true)
        onScrollToBottom()
      }
    }

    element.addEventListener('scroll', handleScroll)

    return () => {
      element.removeEventListener('scroll', handleScroll)
    }
  }, [onScrollToBottom, threshold, hasTriggered])

  const reset = () => {
    setHasTriggered(false)
  }

  return { elementRef, reset }
}
