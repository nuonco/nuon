import { useEffect, useRef, useState } from 'react'
import { ContextTooltip } from '@/components/common/ContextTooltip'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { useBreadcrumb } from '@/hooks/use-breadcrumb'
import type { TNavLink } from '@/types'

const Separator = () => <Icon variant="CaretRightIcon" className="muted" />

const BreadcrumbItem = ({
  crumb,
  isLast,
  isLoading,
}: {
  crumb: TNavLink
  isLast: boolean
  isLoading: boolean
}) => (
  <Text weight="strong">
    {isLoading ? (
      <Skeleton
        height="17px"
        width={`${crumb?.text?.length * 16 * 0.6}px`}
        maxWidth="200px"
      />
    ) : (
      <Link
        href={crumb.path}
        isActive={isLast}
        variant="breadcrumb"
        className="truncate max-w-48 inline-block align-bottom"
      >
        {crumb.text}
      </Link>
    )}
  </Text>
)

const GAP = 8 // gap-2
const ELLIPSIS_ITEM_WIDTH = 40

function computeCollapseCount(
  availableWidth: number,
  itemWidths: number[]
): number {
  if (itemWidths.length <= 2) return 0

  const totalWidth = itemWidths.reduce(
    (sum, w, i) => sum + w + (i > 0 ? GAP : 0),
    0
  )
  if (totalWidth <= availableWidth) return 0

  let hideCount = 0
  let currentWidth = totalWidth

  for (let i = 1; i < itemWidths.length - 1; i++) {
    if (currentWidth <= availableWidth) break
    currentWidth -= itemWidths[i] + GAP
    if (hideCount === 0) currentWidth += ELLIPSIS_ITEM_WIDTH + GAP
    hideCount++
  }

  return Math.min(hideCount, itemWidths.length - 2)
}

function useCollapsedCount(
  navRef: React.RefObject<HTMLElement | null>,
  listRef: React.RefObject<HTMLOListElement | null>,
  totalItems: number
) {
  const [collapsedCount, setCollapsedCount] = useState(0)
  const itemWidthsRef = useRef<number[]>([])
  const prevTotalRef = useRef(totalItems)

  // When breadcrumb items change, clear cached widths and show all for measurement
  if (prevTotalRef.current !== totalItems) {
    prevTotalRef.current = totalItems
    itemWidthsRef.current = []
    setCollapsedCount(0)
  }

  useEffect(() => {
    const nav = navRef.current
    const list = listRef.current
    if (!nav || !list) return

    const ro = new ResizeObserver(() => {
      // If no cached widths yet, we're in measurement mode (all items visible)
      if (itemWidthsRef.current.length === 0) {
        const items = Array.from(list.children) as HTMLElement[]
        if (items.length === 0) return
        itemWidthsRef.current = items.map(
          (el) => el.getBoundingClientRect().width
        )
      }

      const count = computeCollapseCount(
        nav.clientWidth,
        itemWidthsRef.current
      )
      setCollapsedCount(count)
    })
    ro.observe(nav)
    return () => ro.disconnect()
  }, [navRef, listRef, totalItems])

  return collapsedCount
}

export const BreadcrumbNav = () => {
  const { breadcrumbLinks, isLoading } = useBreadcrumb()
  const navRef = useRef<HTMLElement>(null)
  const listRef = useRef<HTMLOListElement>(null)
  const collapsedCount = useCollapsedCount(
    navRef,
    listRef,
    breadcrumbLinks.length
  )

  const firstCrumb = breadcrumbLinks[0]
  if (!firstCrumb) return null

  const shouldCollapse = collapsedCount > 0
  const collapsedCrumbs = shouldCollapse
    ? breadcrumbLinks.slice(1, 1 + collapsedCount)
    : []
  const visibleTail = shouldCollapse
    ? breadcrumbLinks.slice(1 + collapsedCount)
    : breadcrumbLinks.slice(1)

  return (
    <nav ref={navRef} aria-label="Breadcrumb" className="flex-1 min-w-0 overflow-hidden">
      <ol ref={listRef} className="flex items-center gap-2 w-max">
        <li className="flex items-center gap-2">
          <BreadcrumbItem
            crumb={firstCrumb}
            isLast={breadcrumbLinks.length === 1}
            isLoading={isLoading}
          />
        </li>

        {shouldCollapse && (
          <li className="flex items-center gap-2">
            <Separator />
            <ContextTooltip
              items={collapsedCrumbs.map((crumb) => ({
                id: crumb.path,
                title: <Text weight="strong">{crumb.text}</Text>,
                href: crumb.path,
              }))}
              position="bottom"
            >
              <Text weight="strong" className="cursor-default">
                …
              </Text>
            </ContextTooltip>
          </li>
        )}

        {visibleTail.map((crumb, idx) => (
          <li key={crumb.path} className="flex items-center gap-2">
            <Separator />
            <BreadcrumbItem
              crumb={crumb}
              isLast={idx === visibleTail.length - 1}
              isLoading={isLoading}
            />
          </li>
        ))}
      </ol>
    </nav>
  )
}

export const Breadcrumbs = ({ breadcrumbs }: { breadcrumbs: TNavLink[] }) => {
  const { updateBreadcrumb } = useBreadcrumb()
  const key = JSON.stringify(breadcrumbs)

  useEffect(() => {
    updateBreadcrumb(breadcrumbs)
  }, [key])

  return <></>
}
