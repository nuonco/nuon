import { useEffect, useState } from 'react'
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
  const scrollToTop = useScrollToTop()

  useEffect(() => {
    let idleTimer: ReturnType<typeof setTimeout>
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
      clearTimeout(idleTimer)
      idleTimer = setTimeout(() => setIsScrolling(false), idleHideDelay)
    }

    const target = containerId
      ? document.getElementById(containerId)
      : window
    if (!target) return
    target.addEventListener('scroll', handleScroll)
    return () => {
      clearTimeout(idleTimer)
      target.removeEventListener('scroll', handleScroll)
    }
  }, [containerId, scrollOffset, idleHideDelay])

  return (
    <div className="fixed bottom-10 right-0 z-[2] flex justify-end pointer-events-none">
      <TransitionDiv
        className="fade pointer-events-auto mb-4 md:mb-6 mr-8 md:mr-12"
        isVisible={isScrolledPast && (isScrolling || isHovered)}
        onMouseEnter={() => setIsHovered(true)}
        onMouseLeave={() => setIsHovered(false)}
      >
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
      </TransitionDiv>
    </div>
  )
}
