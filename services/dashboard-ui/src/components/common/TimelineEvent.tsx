import React from 'react'
import { cn } from '@/utils/classnames'
import { Badge, type IBadge } from './Badge'
import { Status, type TStatusType } from './Status'
import { Text } from './Text'
import { Time } from './Time'

export interface ITimelineEvent
  extends Omit<React.HTMLAttributes<HTMLDivElement>, 'children' | 'title'> {
  actions?: React.ReactNode
  additionalCaption?: React.ReactNode | string
  badge?: IBadge
  caption?: string
  createdAt: string
  createdBy?: string
  status: TStatusType
  title: React.ReactNode | string
}

export const TimelineEvent = ({
  actions,
  additionalCaption,
  badge,
  caption,
  className,
  createdAt,
  createdBy,
  status,
  title,
  ...props
}: ITimelineEvent) => {
  return (
    <div
      className={cn(
        'flex py-4 gap-6 relative w-full items-start',
        "[&:before]:content-[''] [&:before]:absolute [&:before]:top-0 [&:before]:left-[0.813rem] [&:before]:w-px [&:before]:h-full [&:before]:border-l [&:before]:border-solid",
        '[&:first-child:before]:h-[calc(100%-1.5rem)] [&:first-child:before]:top-[1.5rem]',
        '[&:last-child:before]:h-[1.5rem]',
        className
      )}
      {...props}
    >
      <Status
        status={status}
        variant="timeline"
        isWithoutText
        className="relative z-10"
      />
      <div className="w-full">
        <hgroup className="w-full flex items-center justify-between">
          <Text variant="body" weight="strong">
            {title}
          </Text>

          <span className="flex items-center gap-2">
            {actions ? <span>{actions}</span> : null}
            <Text variant="subtext" theme="neutral">
              <Time time={createdAt} format="relative" variant="subtext" />{' '}
              {createdBy ? `by ${createdBy}` : null}
            </Text>
          </span>
        </hgroup>
        <span className="flex items-center gap-2">
          {caption ? (
            <Text variant="subtext" theme="neutral">
              {caption}
            </Text>
          ) : null}{' '}
          {additionalCaption ? (
            <Text variant="subtext" theme="neutral">
              {additionalCaption}
            </Text>
          ) : null}
          {badge?.children ? <Badge size="sm" {...badge} /> : null}
        </span>
      </div>
    </div>
  )
}
