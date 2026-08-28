import {
  useState,
  useRef,
  useEffect,
  type CSSProperties,
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
  // Overrides a tab's label. Otherwise derived from the key, which lowercases
  // everything after the first character.
  tabLabels?: Record<string, string>
  tabsClassName?: string
  tabControlsClassName?: string
}

export const Tabs = ({
  className,
  initActiveTab,
  tabControlsClassName,
  tabs,
  tabLabels,
  tabsClassName,
  ...props
}: ITabs) => {
  const tabKeys = Object.keys(tabs)
  const [activeTabWidth, setActiveTabWidth] = useState<number | undefined>(
    undefined
  )
  const [activeTabLeft, setActiveTabLeft] = useState<number | undefined>(
    undefined
  )
  const [hoveredTabWidth, setHoveredTabWidth] = useState<number | undefined>(
    undefined
  )
  const [hoveredTabLeft, setHoveredTabLeft] = useState<number | undefined>(
    undefined
  )
  const [activeTab, setActiveTab] = useState(initActiveTab || tabKeys.at(0))
  const [containerHeight, setContainerHeight] = useState<number | undefined>(
    undefined
  )
  const contentRefs = useRef<Record<string, HTMLDivElement | null>>({})
  const tabButtonRefs = useRef<Record<string, HTMLButtonElement | null>>({})
  const containerRef = useRef<HTMLDivElement>(null)
  const heightMeasurementTimeout = useRef<NodeJS.Timeout | null>(null)
  const resizeObserver = useRef<ResizeObserver | null>(null)

  useEffect(() => {
    const activeButton = activeTab ? tabButtonRefs.current[activeTab] : null
    if (activeButton) {
      setActiveTabWidth(activeButton.offsetWidth)
      const parentRect = activeButton.parentElement?.getBoundingClientRect()
      const buttonRect = activeButton.getBoundingClientRect()
      let left = parentRect
        ? buttonRect.left - parentRect.left
        : buttonRect.left
      setActiveTabLeft(left)
      setHoveredTabLeft(left) // Always sync hovered left to active left on load/active change
    }
  }, [activeTab, tabKeys.length])

  const handleTabsMouseLeave = () => {
    setHoveredTabWidth(undefined)
    setHoveredTabLeft(activeTabLeft)
  }

  const handleTabsBlur = () => {
    setHoveredTabWidth(undefined)
    setHoveredTabLeft(activeTabLeft)
  }

  const handleTabHoverOrFocus = (tabKey: string) => {
    const button = tabButtonRefs.current[tabKey]
    if (button) {
      setHoveredTabWidth(button.offsetWidth)
      const parentRect = button.parentElement?.getBoundingClientRect()
      const buttonRect = button.getBoundingClientRect()
      let left = parentRect
        ? buttonRect.left - parentRect.left
        : buttonRect.left
      setHoveredTabLeft(left)
    }
  }

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

    if (heightMeasurementTimeout.current) {
      clearTimeout(heightMeasurementTimeout.current)
    }
    if (resizeObserver.current) {
      resizeObserver.current.disconnect()
    }

    // Wait for TransitionDiv to complete its transition (155ms + small buffer)
    heightMeasurementTimeout.current = setTimeout(() => {
      updateHeight()

      const activeContent = activeTab ? contentRefs.current[activeTab] : null
      if (activeContent) {
        resizeObserver.current = new ResizeObserver(updateHeight)
        resizeObserver.current.observe(activeContent)
      }
    }, 180)

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
          'tab-group flex items-center gap-6 border-b w-full',
          tabControlsClassName
        )}
        onMouseLeave={handleTabsMouseLeave}
        style={
          {
            '--active-tab-width': `${activeTabWidth ?? 0}px`,
            '--active-tab-left': `${activeTabLeft ?? 0}px`,
            '--hovered-tab-width': `${hoveredTabWidth ?? 0}px`,
            '--hovered-tab-left': `${hoveredTabLeft ?? 0}px`,
          } as CSSProperties
        }
      >
        {tabKeys.map((tabKey, idx) => (
          <Button
            className="!border-b-transparent hover:!border-b-transparent"
            key={`${tabKey}-${idx}-btn`}
            isActive={tabKey === activeTab}
            onClick={() => {
              setActiveTab(tabKey)
            }}
            variant="tab"
            ref={(el) => {
              tabButtonRefs.current[tabKey] = el as HTMLButtonElement
            }}
            onMouseEnter={() => handleTabHoverOrFocus(tabKey)}
            onFocus={() => handleTabHoverOrFocus(tabKey)}
            onBlur={handleTabsBlur}
          >
            {tabLabels?.[tabKey] ?? toSentenceCase(camelToWords(tabKey))}
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
            className="absolute top-0 left-0 w-full tab-content"
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
