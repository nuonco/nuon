'use client'

import { useRef, useState } from 'react'
import { cn } from '@/utils/classnames'
import { Icon } from '@/components/common/Icon'
import { usePageSidebar } from '@/hooks/use-page-sidebar'
import type { TNavLink } from '@/types'
import { SubNavLink } from './SubNavLink'

interface ISubNav {
  basePath: string
  links: Array<TNavLink>
}

export const SubNav = ({ basePath, links }: ISubNav) => {
  const {
    isPageSidebarOpen,
    closePageSidebar,
    openPageSidebar,
    togglePageSidebar,
  } = usePageSidebar()
  const [dragging, setDragging] = useState(false)
  const handleRef = useRef<HTMLDivElement>(null)
  const startXRef = useRef<number | null>(null)

  const handleDragStart = (e: React.MouseEvent | React.TouchEvent) => {
    setDragging(true)
    const startX = 'touches' in e ? e.touches[0].clientX : e.clientX
    startXRef.current = startX
  }

  const handleDragMove = (e: React.MouseEvent | React.TouchEvent) => {
    if (!dragging || startXRef.current === null) return

    const currentX = 'touches' in e ? e.touches[0].clientX : e.clientX
    const deltaX = currentX - startXRef.current

    if (deltaX < -1 && isPageSidebarOpen) {
      closePageSidebar()
      setDragging(false)
    } else if (deltaX > 1 && !isPageSidebarOpen) {
      openPageSidebar()
      setDragging(false)
    }
  }

  const handleDragEnd = () => {
    setDragging(false)
    startXRef.current = null
  }

  return (
    <aside
      className={cn(
        // Base styles (mobile)
        'border-b flex shrink-0 overflow-x-auto overflow-y-visible w-screen md:w-[4.5rem]',
        // Desktop styles
        'md:overflow-visible md:relative md:max-h-screen  md:transition-[width] md:duration-fastest md:ease-cubic md:border-b-0 md:border-r md:flex-none',
        {
          'md:w-[17.5rem]': isPageSidebarOpen,
        }
      )}
    >
      <nav
        className={cn(
          // Mobile nav
          'flex shrink-0 gap-8 px-4 py-3 h-16',
          // Desktop nav
          'md:flex-col md:gap-1 md:px-4 md:py-4 md:w-full md:h-full'
        )}
      >
        {links.map((link) => (
          <SubNavLink key={link.path} basePath={basePath} {...link} />
        ))}
      </nav>
      <div
        ref={handleRef}
        className={cn(
          'hidden group', // hide by default
          'md:flex md:absolute md:right-[-1rem] md:w-4 md:h-full md:cursor-pointer md:border-l md:border-transparent',
          'md:transition-[border-color] md:duration-fastest md:ease-cubic',
          'page-nav-handle', // for event handling
          // Add border color on hover
          'hover:!border-primary-600'
        )}
        onMouseDown={handleDragStart}
        onMouseMove={handleDragMove}
        onMouseUp={handleDragEnd}
        onTouchStart={handleDragStart}
        onTouchMove={handleDragMove}
        onTouchEnd={handleDragEnd}
      >
        <button
          className={cn(
            // Base styles (hidden by default)
            'absolute left-[-0.875rem] top-1/2 opacity-0  md:cursor-pointer',
            'border rounded-lg shadow-md p-1 bg-white dark:bg-dark-grey-300',
            'transition-opacity duration-fastest ease-cubic',
            'translate-y-[-50%]',
            // Show on parent hover
            'group-hover:opacity-100' // If you wrap handle in a group, or use peer
            // Show on hover (simulate with parent hover via CSS, or just always show for demo)
          )}
          onClick={() => {
            togglePageSidebar()
          }}
        >
          <Icon variant="SplitHorizontal" />
        </button>
      </div>
    </aside>
  )
}
