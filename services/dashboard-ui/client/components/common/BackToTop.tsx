import { useCallback, useEffect, useRef, useState } from 'react'
import { useScrollToTop } from '@/hooks/use-scroll-to-top'
import { cn } from '@/utils/classnames'
import { Button } from './Button'
import { Icon } from './Icon'
import { TransitionDiv } from './TransitionDiv'

interface IBackToTop {
  containerId?: string
  scrollOffset?: number
  idleHideDelay?: number
}

export const BackToTop = ({
  containerId = 'page-scroll-container',
  scrollOffset = 400,
  idleHideDelay = 1500,
}: IBackToTop) => {
  const [isScrolledPast, setIsScrolledPast] = useState(false)
  const [isScrolling, setIsScrolling] = useState(false)
  const [isHovered, setIsHovered] = useState(false)
  const [isOverlappingAction, setIsOverlappingAction] = useState(false)
  const buttonRef = useRef<HTMLDivElement>(null)
  const buttonRectRef = useRef<DOMRect>()
  const scrollToTop = useScrollToTop()

  const checkForOverlappingAction = useCallback(() => {
    const currentRect = buttonRef.current?.getBoundingClientRect()
    if (currentRect) buttonRectRef.current = currentRect
    const buttonRect = currentRect ?? buttonRectRef.current
    if (!buttonRect) return

    const container = containerId
      ? document.getElementById(containerId)
      : document
    if (!container) return

    const actions = container.querySelectorAll<HTMLElement>(
      'button, [role="button"], input[type="button"], input[type="submit"], input[type="reset"]'
    )
    const overlaps = [...actions].some((action) => {
      if (buttonRef.current?.contains(action)) return false
      const rect = action.getBoundingClientRect()
      return (
        rect.width > 0 &&
        rect.height > 0 &&
        rect.left < buttonRect.right &&
        rect.right > buttonRect.left &&
        rect.top < buttonRect.bottom &&
        rect.bottom > buttonRect.top
      )
    })

    setIsOverlappingAction(overlaps)
  }, [containerId])

  const setButtonNode = useCallback(
    (node: HTMLDivElement | null) => {
      buttonRef.current = node
      if (node) requestAnimationFrame(checkForOverlappingAction)
    },
    [checkForOverlappingAction]
  )

  useEffect(() => {
    let idleTimer: ReturnType<typeof setTimeout>
    let collisionFrame: number
    const handleScroll = () => {
      let scrollTop = 0

      if (containerId) {
        const container = document.getElementById(containerId)
        if (container) {
          scrollTop = container.scrollTop
        }
      } else {
        scrollTop = window.scrollY || document.documentElement.scrollTop
      }

      setIsScrolledPast(scrollTop > scrollOffset)
      setIsScrolling(true)
      cancelAnimationFrame(collisionFrame)
      collisionFrame = requestAnimationFrame(checkForOverlappingAction)
      clearTimeout(idleTimer)
      idleTimer = setTimeout(() => setIsScrolling(false), idleHideDelay)
    }

    const target = containerId ? document.getElementById(containerId) : window
    if (!target) return
    target.addEventListener('scroll', handleScroll)
    window.addEventListener('resize', checkForOverlappingAction)
    return () => {
      cancelAnimationFrame(collisionFrame)
      clearTimeout(idleTimer)
      target.removeEventListener('scroll', handleScroll)
      window.removeEventListener('resize', checkForOverlappingAction)
    }
  }, [checkForOverlappingAction, containerId, scrollOffset, idleHideDelay])

  const shouldShow = isScrolledPast && (isScrolling || isHovered)

  useEffect(() => {
    if (!shouldShow) return
    const frame = requestAnimationFrame(checkForOverlappingAction)
    return () => cancelAnimationFrame(frame)
  }, [checkForOverlappingAction, shouldShow])

  return (
    <div className="fixed bottom-10 right-0 z-[2] flex justify-end pointer-events-none">
      <TransitionDiv
        className="fade pointer-events-auto mb-4 md:mb-6 mr-8 md:mr-12"
        isVisible={shouldShow && !isOverlappingAction}
        onMouseEnter={() => setIsHovered(true)}
        onMouseLeave={() => setIsHovered(false)}
      >
        <div ref={setButtonNode}>
          <Button
            className={cn(
              '!p-3 drop-shadow-lg',
              'bg-btn-gradient-light bg-btn-bg-light dark:bg-btn-gradient-dark dark:bg-btn-bg-dark'
            )}
            onClick={() => scrollToTop(containerId)}
            size="lg"
          >
            <Icon size={18} variant="ArrowUpIcon" />
            <span className="!text-foreground text-sm font-strong">
              Back to top
            </span>
          </Button>
        </div>
      </TransitionDiv>
    </div>
  )
}
