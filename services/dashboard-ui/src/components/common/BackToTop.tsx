'use client'

import { useEffect, useState } from 'react'
import { useScrollToTop } from '@/hooks/use-scroll-to-top'
import { cn } from '@/utils/classnames'
import { Button } from './Button'
import { Icon } from './Icon'
import { TransitionDiv } from './TransitionDiv'

interface IBackToTop {
  containerId?: string
  scrollOffset?: number // Show button after scrolling this many pixels
}

export const BackToTop = ({ containerId, scrollOffset = 400 }: IBackToTop) => {
  const [isVisable, setIsVisable] = useState(false)
  const scrollToTop = useScrollToTop()

  useEffect(() => {
    const handleScroll = () => {
      let scrollTop = 0

      if (containerId) {
        // Get scroll position of specific container
        const container = document.getElementById(containerId)
        if (container) {
          scrollTop = container.scrollTop
        }
      } else {
        // Get scroll position of window
        scrollTop = window.scrollY || document.documentElement.scrollTop
      }

      // Show button if scrolled past threshold
      setIsVisable(scrollTop > scrollOffset)
    }

    // Add scroll listener to appropriate element
    if (containerId) {
      const container = document.getElementById(containerId)
      if (container) {
        container.addEventListener('scroll', handleScroll)

        // Cleanup
        return () => {
          container.removeEventListener('scroll', handleScroll)
        }
      }
    } else {
      // Listen to window scroll
      window.addEventListener('scroll', handleScroll)

      // Cleanup
      return () => {
        window.removeEventListener('scroll', handleScroll)
      }
    }
  }, [containerId, scrollOffset])

  return (
    <TransitionDiv className="fade" isVisible={isVisable}>
      <Button
        className={cn(
          'absolute bottom-4 md:bottom-6 right-4 md:right-6 !p-3 drop-shadow-lg',
          'bg-btn-gradient-light bg-btn-bg-light dark:bg-btn-gradient-dark dark:bg-btn-bg-dark'
        )}
        onClick={() => {
          scrollToTop(containerId)
        }}
        size="lg"
      >
        <Icon size={18} variant="ArrowUp" />
        <span className="!text-foreground text-sm font-strong">
          Back to top
        </span>
      </Button>
    </TransitionDiv>
  )
}
