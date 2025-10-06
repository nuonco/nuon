'use client'

import {
  useState,
  useRef,
  useEffect,
  type HTMLAttributes,
  type ReactNode,
} from 'react'
import { Button } from '@/components/common/Button'
import { TransitionDiv } from '@/components/common/TransitionDiv'
import { cn } from '@/utils/classnames'
import { camelToWords, toSentenceCase } from '@/utils/string-utils'
import './Tabs.css'

interface ITabs extends Omit<HTMLAttributes<HTMLDivElement>, 'children'> {
  initActiveTab?: string
  tabs: Record<string, ReactNode>
  tabsClassName?: string
  tabControlsClassName?: string
}

export const Tabs = ({
  className,
  initActiveTab,
  tabControlsClassName,
  tabs,
  tabsClassName,
  ...props
}: ITabs) => {
  const tabKeys = Object.keys(tabs)
  const [activeTab, setActiveTab] = useState(initActiveTab || tabKeys.at(0))
  const [containerHeight, setContainerHeight] = useState<number | undefined>(
    undefined
  )
  const contentRefs = useRef<Record<string, HTMLDivElement | null>>({})
  const containerRef = useRef<HTMLDivElement>(null)
  const heightMeasurementTimeout = useRef<NodeJS.Timeout | null>(null)
  const resizeObserver = useRef<ResizeObserver | null>(null)

  useEffect(() => {
    const updateHeight = () => {
      if (activeTab && contentRefs.current[activeTab]) {
        const activeContent = contentRefs.current[activeTab]
        if (activeContent) {
          const height = activeContent.scrollHeight
          setContainerHeight(height)
        }
      }
    }

    // Clear any existing timeout and observer
    if (heightMeasurementTimeout.current) {
      clearTimeout(heightMeasurementTimeout.current)
    }
    if (resizeObserver.current) {
      resizeObserver.current.disconnect()
    }

    // Wait for TransitionDiv to complete its transition (155ms + small buffer)
    heightMeasurementTimeout.current = setTimeout(() => {
      updateHeight()

      // Set up ResizeObserver for the active content
      const activeContent = activeTab ? contentRefs.current[activeTab] : null
      if (activeContent) {
        resizeObserver.current = new ResizeObserver(updateHeight)
        resizeObserver.current.observe(activeContent)
      }
    }, 180) // 155ms + 25ms buffer

    return () => {
      if (heightMeasurementTimeout.current) {
        clearTimeout(heightMeasurementTimeout.current)
      }
      if (resizeObserver.current) {
        resizeObserver.current.disconnect()
      }
    }
  }, [activeTab])

  return (
    <div className={cn('tabs flex flex-col', className)} {...props}>
      <div
        className={cn(
          'flex items-center gap-6 border-b w-full',
          tabControlsClassName
        )}
      >
        {tabKeys.map((tabKey, idx) => (
          <Button
            key={`${tabKey}-${idx}-btn`}
            isActive={tabKey === activeTab}
            onClick={() => {
              setActiveTab(tabKey)
            }}
            variant="tab"
          >
            {toSentenceCase(camelToWords(tabKey))}
          </Button>
        ))}
      </div>
      <div
        ref={containerRef}
        className={cn(
          'relative transition-all duration-300 ease-in-out',
          tabsClassName
        )}
        style={{
          height: containerHeight ? `${containerHeight}px` : 'auto',
          minHeight: containerHeight ? `${containerHeight}px` : 'auto',
        }}
      >
        {tabKeys.map((tabKey, idx) => (
          <TransitionDiv
            ref={(el) => {
              contentRefs.current[tabKey] = el
            }}
            className="absolute top-0 left-0 w-full"
            key={`${tabKey}-${idx}-tab`}
            isVisible={tabKey === activeTab}
          >
            {tabs[tabKey]}
          </TransitionDiv>
        ))}
      </div>
    </div>
  )
}
