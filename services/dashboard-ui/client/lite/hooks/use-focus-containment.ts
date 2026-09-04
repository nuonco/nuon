import { useLayoutEffect, type RefObject } from 'react'

const FOCUSABLE =
  'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'

const focusableElements = (container: HTMLElement) =>
  Array.from(container.querySelectorAll<HTMLElement>(FOCUSABLE)).filter(
    (element) =>
      !element.hidden && element.getAttribute('aria-hidden') !== 'true'
  )

export const useFocusContainment = ({
  active,
  containerRef,
  initialFocus = 'first',
  initialFocusRef,
  restoreFocusRef,
  onEscape,
}: {
  active: boolean
  containerRef: RefObject<HTMLElement | null>
  initialFocus?: 'first' | 'container'
  initialFocusRef?: RefObject<HTMLElement | null>
  restoreFocusRef?: RefObject<HTMLElement | null>
  onEscape?: () => void
}) => {
  useLayoutEffect(() => {
    if (!active) return

    const container = containerRef.current
    if (!container) return

    const initial =
      initialFocusRef?.current ??
      container.querySelector<HTMLElement>('[autofocus]') ??
      (initialFocus === 'container'
        ? container
        : (focusableElements(container)[0] ?? container))
    initial.focus()

    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        event.preventDefault()
        onEscape?.()
        return
      }
      if (event.key !== 'Tab') return

      const elements = focusableElements(container)
      if (!elements.length) {
        event.preventDefault()
        container.focus()
        return
      }

      const first = elements[0]
      const last = elements.at(-1)
      if (
        event.shiftKey &&
        (document.activeElement === first ||
          document.activeElement === container)
      ) {
        event.preventDefault()
        last?.focus()
      } else if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault()
        first?.focus()
      }
    }

    container.addEventListener('keydown', onKeyDown)
    return () => {
      container.removeEventListener('keydown', onKeyDown)
      const restore = restoreFocusRef?.current
      const target = restore?.matches(FOCUSABLE)
        ? restore
        : restore?.querySelector<HTMLElement>(FOCUSABLE)
      setTimeout(() => target?.focus(), 0)
    }
  }, [
    active,
    containerRef,
    initialFocus,
    initialFocusRef,
    onEscape,
    restoreFocusRef,
  ])
}
