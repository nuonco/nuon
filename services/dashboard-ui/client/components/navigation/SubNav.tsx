import { useRef, useState } from 'react'
import { cn } from '@/utils/classnames'
import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import { usePageSidebar } from '@/hooks/use-page-sidebar'
import type { TNavItem, TNavLink } from '@/types'
import { SubNavLink } from './SubNavLink'

function isSection(item: TNavItem): item is { type: 'section'; label: string } {
  return 'type' in item && item.type === 'section'
}

interface ISubNav {
  basePath: string
  links: Array<TNavItem>
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
        'group/sidebar border-b flex shrink-0 overflow-x-auto overflow-y-visible w-full md:w-[4.5rem]',
        'md:overflow-visible md:relative md:transition-[width] md:duration-fastest md:ease-cubic md:border-b-0 md:border-r md:flex-none',
        {
          'md:w-[17.5rem]': isPageSidebarOpen,
        }
      )}
    >
      <nav
        className={cn(
          'flex shrink-0 gap-8 px-4 py-3 h-16',
          'md:sticky md:top-0 md:flex-col md:gap-1 md:px-4 md:py-4 md:w-full'
        )}
      >
        {links.map((item, i) =>
          isSection(item) ? (
            i === 0 ? null : (
              <div
                key={`section-${item.label}`}
                className={cn(
                  'hidden md:flex items-center transition-all duration-fast ease-cubic',
                  {
                    'px-3 mt-2 mb-0.5': isPageSidebarOpen,
                    'mx-2 mt-4 mb-1': !isPageSidebarOpen,
                  }
                )}
              >
                <Text
                  variant="label"
                  theme="neutral"
                  family="mono"
                  className={cn(
                    'uppercase tracking-wider text-[10px] !grid duration-fast transition-all ease-cubic',
                    {
                      'md:grid-cols-[1fr] md:opacity-100 mr-2':
                        isPageSidebarOpen,
                      'md:grid-cols-[0fr] md:opacity-0 mr-0':
                        !isPageSidebarOpen,
                    }
                  )}
                >
                  <span className="overflow-hidden">{item.label}</span>
                </Text>

                <div className="h-px flex-1 bg-cool-grey-200 dark:bg-white/10" />
              </div>
            )
          ) : (
            <SubNavLink key={item.path} basePath={basePath} {...item} />
          )
        )}
      </nav>

      <div
        ref={handleRef}
        className={cn(
          'hidden',
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
            'fixed top-1/2 opacity-0 md:cursor-pointer',
            'border rounded-lg shadow-md p-1 bg-white dark:bg-dark-grey-300',
            'transition-opacity duration-fastest ease-cubic',
            '-translate-x-1/2 -translate-y-1/2',
            // Show when hovering anywhere on the sidebar
            'group-hover/sidebar:opacity-100'
          )}
          onClick={() => {
            togglePageSidebar()
          }}
        >
          <Tooltip
            position="right"
            tipContent={
              <div className="w-fit">
                <Text flex nowrap className="gap-2" variant="subtext">
                  {isPageSidebarOpen ? 'Collapse' : 'Expand'} sidebar
                  <span className="inline-flex gap-0.5">
                    <Badge variant="code" size="sm">
                      ALT
                    </Badge>
                    <Badge variant="code" size="sm">
                      SHIFT
                    </Badge>
                    <Badge variant="code" size="sm">
                      S
                    </Badge>
                  </span>
                </Text>
              </div>
            }
          >
            <Icon variant="SplitHorizontalIcon" />
          </Tooltip>
        </button>
      </div>
    </aside>
  )
}
