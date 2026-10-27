import { useLayoutEffect, useRef, useState, type ReactNode } from 'react'
import { cn } from '@/utils/classnames'
import { Button } from './Button'
import { Icon, type TIconVariant } from './Icon'

export type TShowMoreIcons = 'caret' | 'plusMinus'

const ICONS: Record<TShowMoreIcons, { closed: TIconVariant; open: TIconVariant }> = {
  plusMinus: { closed: 'PlusIcon', open: 'MinusIcon' },
  caret: { closed: 'CaretDownIcon', open: 'CaretUpIcon' },
}

export interface IShowMore {
  children: ReactNode
  className?: string
  lines?: number
  maxHeight?: number
  label?: string
  expandedLabel?: string
  icons?: TShowMoreIcons
  expanded?: boolean
  defaultExpanded?: boolean
  onExpandedChange?: (expanded: boolean) => void
}

export const ShowMore = ({
  children,
  className,
  lines = 3,
  maxHeight,
  label = 'Show more',
  expandedLabel = 'Show less',
  icons = 'caret',
  expanded: expandedProp,
  defaultExpanded = false,
  onExpandedChange,
}: IShowMore) => {
  const [expandedState, setExpandedState] = useState(defaultExpanded)
  const expanded = expandedProp ?? expandedState
  const [canExpand, setCanExpand] = useState(false)
  const contentRef = useRef<HTMLDivElement>(null)
  const collapsed = !expanded
  const collapsedHeight = maxHeight != null ? `${maxHeight}px` : `calc(${lines} * 1lh)`

  useLayoutEffect(() => {
    const el = contentRef.current
    if (!el || expanded) return

    const check = () => setCanExpand(el.scrollHeight > el.clientHeight + 1)
    check()

    const observer = new ResizeObserver(check)
    observer.observe(el)
    return () => observer.disconnect()
  }, [expanded, children, collapsedHeight])

  const toggle = () => {
    const next = !expanded
    onExpandedChange?.(next)
    if (expandedProp === undefined) setExpandedState(next)
  }

  const icon = ICONS[icons]
  const showButton = canExpand || expanded

  return (
    <div className={cn('flex flex-col items-start gap-1 min-w-0 w-full', className)}>
      <div
        ref={contentRef}
        style={collapsed ? { maxHeight: collapsedHeight } : undefined}
        className={cn(
          'w-full',
          collapsed && 'overflow-hidden',
          collapsed &&
            canExpand &&
            '[mask-image:linear-gradient(to_bottom,black_65%,transparent)]',
        )}
      >
        {children}
      </div>
      {showButton ? (
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className="-ml-2"
          aria-expanded={expanded}
          onClick={toggle}
        >
          {expanded ? expandedLabel : label}
          <Icon variant={expanded ? icon.open : icon.closed} size={14} />
        </Button>
      ) : null}
    </div>
  )
}
