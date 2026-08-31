import { useEffect, useRef, useState } from 'react'

const scrollParent = (node: HTMLElement | null) => {
  let current = node?.parentElement ?? null

  while (current) {
    const { overflowY } = getComputedStyle(current)
    if (overflowY === 'auto' || overflowY === 'scroll') return current
    current = current.parentElement
  }

  return null
}

export const useScrolled = () => {
  const ref = useRef<HTMLDivElement>(null)
  const [isScrolled, setIsScrolled] = useState(false)

  useEffect(() => {
    const container = scrollParent(ref.current)
    if (!container) return

    const onScroll = () => setIsScrolled(container.scrollTop > 0)

    onScroll()
    container.addEventListener('scroll', onScroll, { passive: true })
    return () => container.removeEventListener('scroll', onScroll)
  }, [])

  return { ref, isScrolled }
}
