import { useEffect, useRef, useState, type CSSProperties } from 'react'
import { useLocation } from 'react-router'
import { Button } from '@/components/common/Button'
import type { TNavLink } from '@/types'
import './TabNav.css'

export interface ITabNav {
  activeIndex?: number
  basePath: string
  tabs: TNavLink[]
}

export const TabNav = ({ activeIndex, basePath, tabs }: ITabNav) => {
  const { pathname } = useLocation()

  const [activeTabWidth, setActiveTabWidth] = useState<number | undefined>()
  const [activeTabLeft, setActiveTabLeft] = useState<number | undefined>()
  const [hoveredTabWidth, setHoveredTabWidth] = useState<number | undefined>()
  const [hoveredTabLeft, setHoveredTabLeft] = useState<number | undefined>()
  const tabRefs = useRef<Record<string, HTMLButtonElement | null>>({})

  const getActiveIndex = () => {
    if (activeIndex !== undefined) return activeIndex
    return tabs.findIndex((tab) => {
      const href = `${basePath}${tab.path === '/' ? '' : tab.path}`
      return pathname === href
    })
  }

  const measureTab = (key: string) => {
    const button = tabRefs.current[key]
    if (!button) return undefined
    const parentRect = button.parentElement?.getBoundingClientRect()
    const buttonRect = button.getBoundingClientRect()
    const left = parentRect
      ? buttonRect.left - parentRect.left
      : buttonRect.left
    return { width: button.offsetWidth, left }
  }

  useEffect(() => {
    const idx = getActiveIndex()
    if (idx < 0) return
    const tab = tabs[idx]
    const m = measureTab(tab.path)
    if (m) {
      setActiveTabWidth(m.width)
      setActiveTabLeft(m.left)
      setHoveredTabLeft(m.left)
    }
  }, [pathname, activeIndex, tabs.length])

  const handleMouseEnter = (path: string) => {
    const m = measureTab(path)
    if (m) {
      setHoveredTabWidth(m.width)
      setHoveredTabLeft(m.left)
    }
  }

  const handleMouseLeave = () => {
    setHoveredTabWidth(undefined)
    setHoveredTabLeft(activeTabLeft)
  }

  return (
    <nav
      aria-label="tab navigation"
      className="tab-nav flex items-center gap-6 border-b w-full relative"
      onMouseLeave={handleMouseLeave}
      style={
        {
          '--active-tab-width': `${activeTabWidth ?? 0}px`,
          '--active-tab-left': `${activeTabLeft ?? 0}px`,
          '--hovered-tab-width': `${hoveredTabWidth ?? 0}px`,
          '--hovered-tab-left': `${hoveredTabLeft ?? 0}px`,
        } as CSSProperties
      }
    >
      {tabs.map((tab, i) => {
        const href = `${basePath}${tab.path === '/' ? '' : tab.path}`
        const isActive =
          activeIndex !== undefined ? i === activeIndex : pathname === href

        return (
          <Button
            className="!border-b-transparent hover:!border-b-transparent"
            key={tab.path}
            href={href}
            isActive={isActive}
            variant="tab"
            ref={(el) => {
              tabRefs.current[tab.path] = el as HTMLButtonElement
            }}
            onMouseEnter={() => handleMouseEnter(tab.path)}
          >
            <span className="flex items-center gap-2">
              {tab.badge ? (
                <span className="size-2.5 rounded-full bg-amber-500" />
              ) : null}
              {tab.text}
            </span>
          </Button>
        )
      })}
    </nav>
  )
}
