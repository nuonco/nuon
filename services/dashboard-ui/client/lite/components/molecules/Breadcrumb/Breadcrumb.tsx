import {
  Fragment,
  useCallback,
  useLayoutEffect,
  useRef,
  useState,
  type RefObject,
} from 'react'
import { cn } from '@/utils/classnames'
import type { IBreadcrumbItem } from '../../../providers/breadcrumb-provider'
import { Button } from '../../atoms/Button'
import { Dropdown } from '../../atoms/Dropdown'
import { Icon } from '../../atoms/Icon'
import { Link } from '../../atoms/Link'
import { Text } from '../../atoms/Text'
import { Menu, MenuItem } from '../Menu'

const ITEM_GAP = 8
const OVERFLOW_WIDTH = 40

export const hiddenBreadcrumbIndices = (
  availableWidth: number,
  itemWidths: number[]
) => {
  if (availableWidth <= 0 || itemWidths.length < 2) return []

  let width =
    itemWidths.reduce((total, itemWidth) => total + itemWidth, 0) +
    ITEM_GAP * (itemWidths.length - 1)
  if (width <= availableWidth) return []

  const hidden: number[] = []
  for (let index = 1; index < itemWidths.length - 1; index++) {
    width -= itemWidths[index] + ITEM_GAP
    if (hidden.length === 0) width += OVERFLOW_WIDTH + ITEM_GAP
    hidden.push(index)
    if (width <= availableWidth) return hidden
  }

  if (itemWidths.length > 1) hidden.unshift(0)
  return hidden
}

const useHiddenIndices = (
  navRef: RefObject<HTMLElement | null>,
  measureRef: RefObject<HTMLOListElement | null>,
  signature: string
) => {
  const [hidden, setHidden] = useState<number[]>([])

  const measure = useCallback(() => {
    const nav = navRef.current
    const list = measureRef.current
    if (!nav || !list) return
    const widths = [...list.children].map(
      (child) => child.getBoundingClientRect().width
    )
    setHidden(hiddenBreadcrumbIndices(nav.clientWidth, widths))
  }, [measureRef, navRef])

  useLayoutEffect(() => {
    measure()
    const nav = navRef.current
    if (!nav || typeof ResizeObserver === 'undefined') return
    const observer = new ResizeObserver(measure)
    observer.observe(nav)
    return () => observer.disconnect()
  }, [measure, navRef, signature])

  return hidden
}

const Separator = () => (
  <Icon
    variant="CaretRightIcon"
    size={14}
    className="shrink-0 text-tertiary"
    aria-hidden
  />
)

const Crumb = ({
  item,
  current,
  measure,
}: {
  item: IBreadcrumbItem
  current: boolean
  measure: boolean
}) => {
  if (measure || (!current && !item.href)) {
    return (
      <Text
        variant="caption"
        color={current ? 'primary' : 'secondary'}
        loading={!item.label}
        loadingWidth={item.loadingWidth}
        className="block max-w-48 truncate whitespace-nowrap"
      >
        {item.label}
      </Text>
    )
  }

  if (current) {
    return (
      <Text
        variant="caption"
        weight="medium"
        color="primary"
        loading={!item.label}
        loadingWidth={item.loadingWidth}
        aria-current="page"
        className="block truncate"
      >
        {item.label}
      </Text>
    )
  }

  return (
    <Link
      href={item.href ?? '#'}
      variant="caption"
      loading={!item.label}
      loadingWidth={item.loadingWidth}
      className="block max-w-48 truncate whitespace-nowrap no-underline"
    >
      {item.label}
    </Link>
  )
}

const CrumbItem = ({
  item,
  index,
  current,
  measure = false,
}: {
  item: IBreadcrumbItem
  index: number
  current: boolean
  measure?: boolean
}) => (
  <li
    className={cn(
      'flex min-w-0 items-center gap-2',
      current && !measure && 'flex-1'
    )}
    data-breadcrumb-index={index}
  >
    {index > 0 ? <Separator /> : null}
    <Crumb item={item} current={current} measure={measure} />
  </li>
)

export interface IBreadcrumb {
  items: IBreadcrumbItem[]
  className?: string
}

export const Breadcrumb = ({ items, className }: IBreadcrumb) => {
  const navRef = useRef<HTMLElement>(null)
  const measureRef = useRef<HTMLOListElement>(null)
  const signature = JSON.stringify(items)
  const hidden = useHiddenIndices(navRef, measureRef, signature)
  const hiddenItems = items.filter((_, index) => hidden.includes(index))

  if (!items.length) return null

  return (
    <nav
      ref={navRef}
      aria-label="Breadcrumb"
      className={cn('relative min-w-0 flex-1 overflow-hidden', className)}
    >
      <ol
        ref={measureRef}
        aria-hidden
        className="pointer-events-none invisible absolute flex w-max items-center gap-2"
      >
        {items.map((item, index) => (
          <CrumbItem
            key={`${item.href ?? 'current'}-${index}`}
            item={item}
            index={index}
            current={index === items.length - 1}
            measure
          />
        ))}
      </ol>

      <ol className="flex min-w-0 items-center gap-2">
        {items.map((item, index) => {
          if (hidden.includes(index)) return null
          const previousHidden = index > 0 && hidden.includes(index - 1)

          return (
            <Fragment key={`${item.href ?? 'current'}-${index}`}>
              {previousHidden ? (
                <li className="flex shrink-0 items-center gap-2">
                  <Separator />
                  <Dropdown
                    align="start"
                    haspopup="menu"
                    contentClassName="bg-popover-bg text-popover-text shadow-[var(--popover-shadow)]"
                    trigger={
                      <Button
                        variant="ghost"
                        size="sm"
                        iconOnly
                        aria-label="Show parent pages"
                      >
                        <Icon variant="DotsThreeIcon" size={18} aria-hidden />
                      </Button>
                    }
                  >
                    <Menu>
                      {hiddenItems.map((hiddenItem, hiddenIndex) => (
                        <MenuItem
                          key={`${hiddenItem.href ?? 'hidden'}-${hiddenIndex}`}
                          href={hiddenItem.href}
                          disabled={!hiddenItem.label || !hiddenItem.href}
                        >
                          {hiddenItem.label ?? 'Loading'}
                        </MenuItem>
                      ))}
                    </Menu>
                  </Dropdown>
                </li>
              ) : null}
              <CrumbItem
                item={item}
                index={index}
                current={index === items.length - 1}
              />
            </Fragment>
          )
        })}
      </ol>
    </nav>
  )
}
