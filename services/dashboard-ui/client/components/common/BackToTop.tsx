import { useEffect, useState } from 'react'
import { useScrollToTop } from '@/hooks/use-scroll-to-top'
import { cn } from '@/utils/classnames'
import { Button } from './Button'
import { Icon } from './Icon'
import { TransitionDiv } from './TransitionDiv'

interface IBackToTop {
  containerId?: string
  scrollOffset?: number
}

export const BackToTop = ({ containerId = 'page-scroll-container', scrollOffset = 400 }: IBackToTop) => {
  const [isVisable, setIsVisable] = useState(false)
  const scrollToTop = useScrollToTop()

  useEffect(() => {
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

      setIsVisable(scrollTop > scrollOffset)
    }

    if (containerId) {
      const container = document.getElementById(containerId)
      if (container) {
        container.addEventListener('scroll', handleScroll)
        return () => container.removeEventListener('scroll', handleScroll)
      }
    } else {
      window.addEventListener('scroll', handleScroll)
      return () => window.removeEventListener('scroll', handleScroll)
    }
  }, [containerId, scrollOffset])

  return (
    <div className="fixed bottom-10 right-0 z-[2] flex justify-end pointer-events-none">
      <TransitionDiv className="fade pointer-events-auto mb-4 md:mb-6 mr-8 md:mr-12" isVisible={isVisable}>
        <Button
          className={cn(
            '!p-3 drop-shadow-lg',
            'bg-btn-gradient-light bg-btn-bg-light dark:bg-btn-gradient-dark dark:bg-btn-bg-dark'
          )}
          onClick={() => scrollToTop(containerId)}
          size="lg"
        >
          <Icon size={18} variant="ArrowUp" />
          <span className="!text-foreground text-sm font-strong">
            Back to top
          </span>
        </Button>
      </TransitionDiv>
    </div>
  )
}
