import type { HTMLAttributes, ReactNode } from 'react'
import type { TCompositeStatus } from '@/types/ctl-api.types'
import { cn } from '@/utils/classnames'
import type { TIconVariant } from '../atoms/Icon'
import { Status } from '../atoms/Status'
import { Text } from '../atoms/Text'
import { Time } from './Time'

export interface ITimelineItem
  extends Omit<HTMLAttributes<HTMLLIElement>, 'title'> {
  title: ReactNode
  status?: string | TCompositeStatus
  statusIcon?: TIconVariant
  time?: string
  actions?: ReactNode
  loading?: boolean
}

const MARKER_CLASSES = 'flex h-6 shrink-0 items-center'
const CONNECTOR_CLASSES = 'mt-0.5 w-px flex-1 bg-divider group-last:hidden'

export const TimelineItem = ({
  title,
  status,
  statusIcon,
  time,
  actions,
  loading = false,
  className,
  children,
  ...props
}: ITimelineItem) => {
  return (
    <li className={cn('group flex min-w-0 gap-2.5', className)} {...props}>
      <span className="flex shrink-0 flex-col items-center">
        <span data-timeline-marker className={MARKER_CLASSES}>
          {loading ? (
            <Status loading variant="icon" />
          ) : status ? (
            <Status status={status} icon={statusIcon} variant="icon" />
          ) : (
            <span
              aria-hidden
              className="size-6 rounded-full border border-divider"
            />
          )}
        </span>
        <span
          aria-hidden
          data-timeline-connector
          className={CONNECTOR_CLASSES}
        />
      </span>

      <div className="flex min-w-0 flex-1 flex-col gap-1 pb-5">
        <div className="flex min-w-0 flex-wrap items-start justify-between gap-x-3 gap-y-1">
          {loading ? (
            <Text weight="medium" loading loadingWidth={24} />
          ) : (
            <Text weight="medium" color="primary" className="min-w-0">
              {title}
            </Text>
          )}
          {actions && !loading ? (
            <span className="flex shrink-0 items-center gap-1">{actions}</span>
          ) : null}
        </div>

        {loading ? (
          <Text variant="caption" loading loadingWidth={14} />
        ) : time ? (
          <Time
            value={time}
            format="relative"
            color="tertiary"
            tooltipSide="bottom"
          />
        ) : null}

        {children && !loading ? (
          <div className="flex min-w-0 flex-col gap-1">{children}</div>
        ) : null}
      </div>
    </li>
  )
}
