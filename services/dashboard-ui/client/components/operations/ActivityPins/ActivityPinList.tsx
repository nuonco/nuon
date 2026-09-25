import type { ReactNode } from 'react'
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Checkbox } from '@/components/common/form/CheckboxInput'
import { cn } from '@/utils/classnames'

export type TActivityPinListItem = {
  checked: boolean
  description?: string
  disabled?: boolean
  id: string
  loading?: boolean
  name: string
}

export interface IActivityPinList {
  emptyMessage: string
  emptyTitle: string
  filtered?: boolean
  hint?: string
  items: TActivityPinListItem[]
  loading?: boolean
  onOffsetChange?: (offset: number) => void
  onToggle: (id: string) => void
  pagination?: {
    hasNext: boolean
    limit: number
    offset: number
  }
  search?: ReactNode
  title: string
}

const loadingItems = (): TActivityPinListItem[] =>
  Array.from({ length: 3 }, (_, index) => ({
    id: `loading-${index}`,
    name: '',
    loading: true,
    checked: false,
    disabled: true,
  }))

export const ActivityPinList = ({
  emptyMessage,
  emptyTitle,
  filtered = false,
  hint,
  items,
  loading = false,
  onOffsetChange,
  onToggle,
  pagination,
  search,
  title,
}: IActivityPinList) => {
  const rows = loading && !items.length ? loadingItems() : items
  const showPagination =
    !!pagination &&
    !!onOffsetChange &&
    (pagination.offset > 0 || pagination.hasNext)

  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-col gap-1">
        <Text variant="body" weight="strong" role="heading" level={3}>
          {title}
        </Text>
        {hint ? (
          <Text variant="subtext" theme="neutral">
            {hint}
          </Text>
        ) : null}
      </div>
      {search}
      {rows.length ? (
        <div className="flex flex-col">
          {rows.map((item) => (
            <label
              key={item.id}
              className={cn(
                'flex items-start gap-3 rounded-md p-2',
                item.disabled
                  ? 'opacity-60'
                  : 'cursor-pointer hover:bg-black/5 dark:hover:bg-white/5'
              )}
            >
              <Checkbox
                className="mt-1 shrink-0"
                checked={item.checked}
                disabled={item.disabled || item.loading}
                aria-label={item.name || title}
                title={
                  item.disabled && !item.checked
                    ? 'Remove a pin before adding another'
                    : undefined
                }
                onChange={() => onToggle(item.id)}
              />
              <span className="flex flex-col min-w-0">
                <Text
                  weight="strong"
                  className="truncate"
                  loading={item.loading}
                  loadingWidth={16}
                >
                  {item.name}
                </Text>
                {item.description ? (
                  <Text variant="subtext" theme="neutral" className="line-clamp-2">
                    {item.description}
                  </Text>
                ) : item.loading ? (
                  <Text variant="subtext" loading loadingWidth={28} />
                ) : null}
              </span>
            </label>
          ))}
        </div>
      ) : (
        <EmptyState
          size="sm"
          variant={filtered ? 'search' : 'table'}
          emptyTitle={emptyTitle}
          emptyMessage={emptyMessage}
        />
      )}
      {showPagination && pagination && onOffsetChange ? (
        <div className="flex items-center gap-3">
          <Button
            disabled={pagination.offset === 0}
            onClick={() =>
              onOffsetChange(Math.max(pagination.offset - pagination.limit, 0))
            }
            title="previous"
          >
            <Icon variant="ArrowLeftIcon" />
          </Button>
          <Button
            disabled={!pagination.hasNext}
            onClick={() => onOffsetChange(pagination.offset + pagination.limit)}
            title="next"
          >
            <Icon variant="ArrowRightIcon" />
          </Button>
        </div>
      ) : null}
    </div>
  )
}
