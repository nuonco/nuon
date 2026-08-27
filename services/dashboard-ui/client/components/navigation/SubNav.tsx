import { Fragment, useRef, useState } from 'react'
import { useLocation } from 'react-router'
import { cn } from '@/utils/classnames'
import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import { usePageSidebar } from '@/hooks/use-page-sidebar'
import { useStoredRecord } from '@/hooks/use-stored-record'
import { isNavLinkActive } from '@/utils/nav-active'
import type { TNavAction, TNavItem, TNavLink, TNavSectionHeader } from '@/types'
import { SubNavLink } from './SubNavLink'
import { SubNavButton } from './SubNavButton'

function isSection(item: TNavItem): item is TNavSectionHeader {
  return 'type' in item && item.type === 'section'
}

function isAction(item: TNavItem): item is TNavAction {
  return 'type' in item && item.type === 'action'
}

type TNavGroup = {
  key: string
  header: TNavSectionHeader | null
  items: Array<TNavLink | TNavAction>
}

function groupItems(links: Array<TNavItem>): TNavGroup[] {
  const groups: TNavGroup[] = []
  let current: TNavGroup | null = null

  links.forEach((item, i) => {
    if (isSection(item)) {
      current = { key: `${item.label}-${i}`, header: i === 0 ? null : item, items: [] }
      groups.push(current)
    } else {
      if (!current) {
        current = { key: `group-${i}`, header: null, items: [] }
        groups.push(current)
      }
      current.items.push(item)
    }
  })

  return groups
}

interface ISubNav {
  basePath: string
  links: Array<TNavItem>
  storageKey?: string
}

export const SubNav = ({ basePath, links, storageKey = 'subnav-sections' }: ISubNav) => {
  const {
    isPageSidebarOpen,
    closePageSidebar,
    openPageSidebar,
    togglePageSidebar,
  } = usePageSidebar()
  const [sectionState, setSectionOpen] = useStoredRecord<boolean>(storageKey)
  const { pathname } = useLocation()
  const groups = groupItems(links)
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
          'md:sticky md:top-0 md:flex-col md:gap-1 md:px-4 md:py-4 md:w-full md:h-auto'
        )}
      >
        {groups.map((group) => {
          const renderItem = (item: TNavLink | TNavAction) =>
            isAction(item) ? (
              <SubNavButton
                key={item.key}
                iconVariant={item.iconVariant}
                text={item.text}
                onClick={item.onClick}
                isActive={item.isActive}
              />
            ) : (
              <SubNavLink key={item.path} basePath={basePath} {...item} />
            )

          if (!group.header) {
            return (
              <Fragment key={group.key}>{group.items.map(renderItem)}</Fragment>
            )
          }

          const label = group.header.label
          const userOpen =
            sectionState[label] ?? group.header.defaultOpen ?? true
          const hasActiveItem = group.items.some((item) =>
            isAction(item)
              ? item.isActive
              : isNavLinkActive(basePath, item.path, pathname, item.matchPaths)
          )
          const isOpen = !isPageSidebarOpen || userOpen || hasActiveItem

          return (
            <div key={group.key} className="contents md:block">
              <button
                type="button"
                onClick={() => {
                  if (isPageSidebarOpen) setSectionOpen(label, !userOpen)
                }}
                aria-expanded={isOpen}
                className={cn(
                  'group/section hidden md:flex items-center w-full text-left rounded-md transition-all duration-fast ease-cubic',
                  {
                    'px-3 py-1 mt-1.5 mb-0.5 cursor-pointer hover:bg-black/5 dark:hover:bg-white/5':
                      isPageSidebarOpen,
                    'px-2 mt-1 mb-1 pointer-events-none': !isPageSidebarOpen,
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
                      'md:grid-cols-[1fr] md:opacity-100 mr-2': isPageSidebarOpen,
                      'md:grid-cols-[0fr] md:opacity-0 mr-0': !isPageSidebarOpen,
                    }
                  )}
                >
                  <span className="overflow-hidden">{label}</span>
                </Text>

                <div className="h-px flex-1 bg-cool-grey-200 dark:bg-white/10" />

                <Icon
                  variant="CaretDownIcon"
                  size={12}
                  className={cn(
                    'shrink-0 text-cool-grey-400 transition-all duration-fast ease-cubic',
                    'group-hover/section:text-cool-grey-600 dark:group-hover/section:text-cool-grey-300',
                    {
                      'md:opacity-100 ml-2': isPageSidebarOpen,
                      'md:opacity-0 md:w-0 ml-0': !isPageSidebarOpen,
                      '-rotate-90': !isOpen,
                    }
                  )}
                />
              </button>

              <div
                className={cn(
                  'contents md:grid md:transition-[grid-template-rows] md:duration-fast md:ease-cubic',
                  isOpen ? 'md:grid-rows-[1fr]' : 'md:grid-rows-[0fr]'
                )}
              >
                <div className="contents md:flex md:min-h-0 md:flex-col md:gap-1 md:overflow-hidden">
                  {group.items.map(renderItem)}
                </div>
              </div>
            </div>
          )
        })}
      </nav>

      <div
        ref={handleRef}
        className={cn(
          'hidden',
          'md:flex md:absolute md:right-[-1rem] md:w-4 md:h-full md:cursor-pointer md:border-l md:border-transparent',
          'md:transition-[border-color] md:duration-fastest md:ease-cubic',
          'page-nav-handle', // for event handling
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
            'fixed top-1/2 opacity-0 md:cursor-pointer',
            'border rounded-lg shadow-md p-1 bg-white dark:bg-dark-grey-300',
            'transition-opacity duration-fastest ease-cubic',
            '-translate-x-1/2 -translate-y-1/2',
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
