import React from 'react'
import { cn } from '@/utils/classnames'
import { toSentenceCase } from '@/utils/string-utils'
import { Badge, type IBadge } from './Badge'
import { Duration } from './Duration'
import { Icon } from './Icon'
import { Status, type TStatusType } from './Status'
import { Text } from './Text'
import { Time } from './Time'
import { Tooltip } from './Tooltip'

export interface ITimelineEvent
  extends Omit<React.HTMLAttributes<HTMLDivElement>, 'children' | 'title'> {
  actions?: React.ReactNode
  additionalCaption?: React.ReactNode | string
  badge?: IBadge
  caption?: React.ReactNode | string
  createdAt: string
  createdBy?: string
  duration?: number
  status: TStatusType
  title: React.ReactNode | string
  underline?: React.ReactNode | string
  updatedAt?: string
}

export const TimelineEvent = ({
  actions,
  additionalCaption,
  badge,
  caption,
  className,
  createdAt,
  createdBy,
  duration,
  status,
  title,
  underline,
  updatedAt,
  ...props
}: ITimelineEvent) => {
  return (
    <div
      className={cn(
        'flex py-4 gap-6 relative w-full items-start',
        "[&:before]:content-[''] [&:before]:absolute [&:before]:top-0 [&:before]:left-[0.813rem] [&:before]:w-px [&:before]:h-full [&:before]:border-l [&:before]:border-solid",
        '[&:first-child:before]:h-[calc(100%-1.5rem)] [&:first-child:before]:top-[1.5rem]',
        '[&:last-child:before]:h-[1.5rem]',
        '[&:only-child:before]:h-0',
        className
      )}
      {...props}
    >
      <Tooltip
        tipContentClassName="flex"
        tipContent={
          <Text variant="subtext" family="mono">
            {toSentenceCase(status)}
          </Text>
        }
        position="right"
      >
        <Status
          status={status}
          variant="timeline"
          isWithoutText
          className="relative z-1"
        />
      </Tooltip>
      <div className="w-full flex flex-col gap-1">
        <hgroup className="w-full flex items-center justify-between">
          <div className="flex items-center gap-3 min-w-0 flex-wrap">
            {title}
            {additionalCaption}
            {badge?.children ? <Badge {...badge} /> : null}
          </div>

          <span className="flex items-center gap-2 shrink-0">
            <Text variant="subtext" theme="neutral">
              <Time time={updatedAt || createdAt} format="relative" variant="subtext" />
              {createdBy ? ` by ${createdBy}` : null}
            </Text>
            {actions ? <span>{actions}</span> : null}
          </span>
        </hgroup>
        <div className="flex items-center gap-3 flex-wrap">
          <Text flex className="gap-1" variant="subtext" theme="neutral">
            <Icon variant="CalendarBlankIcon" size={14} />
            <Time time={createdAt} variant="subtext" />
          </Text>
          {duration ? (
            <Text flex className="gap-1" variant="subtext" theme="neutral">
              <Icon variant="TimerIcon" size={14} />
              <Duration nanoseconds={duration} variant="subtext" />
            </Text>
          ) : null}
          {caption ? (
            <Text variant="subtext" theme="neutral">
              {caption}
            </Text>
          ) : null}
        </div>
        {underline ? (
          <Text variant="label" theme="neutral">
            {underline}
          </Text>
        ) : null}
      </div>
    </div>
  )
}
