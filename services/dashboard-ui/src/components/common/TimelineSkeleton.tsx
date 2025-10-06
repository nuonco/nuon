import React from 'react'
import { cn } from '@/utils/classnames'
import { Skeleton } from './Skeleton'
import { Status } from './Status'

export const TimelineSkeleton = ({
  eventCount = 5,
  ...props
}: Omit<React.HTMLAttributes<HTMLDivElement>, 'children'> & {
  eventCount?: number
}) => {
  return (
    <div {...props}>
      <div>
        <Skeleton height="24px" width="65px" />

        <div className="flex flex-col">
          {Array.from({ length: eventCount }).map((_, i) => (
            <TimelineEventSkeleton key={i} />
          ))}
        </div>
      </div>
    </div>
  )
}

const TimelineEventSkeleton = ({ className }: { className?: string }) => (
  <div
    className={cn(
      'flex py-4 gap-6 relative w-full items-start',
      "[&:before]:content-[''] [&:before]:absolute [&:before]:top-0 [&:before]:left-[0.813rem] [&:before]:w-px [&:before]:h-full [&:before]:border-l [&:before]:border-solid",
      '[&:first-child:before]:h-[calc(100%-1.5rem)] [&:first-child:before]:top-[1.5rem]',
      '[&:last-child:before]:h-[1.5rem]',
      className
    )}
  >
    <Status
      status="skeleton"
      variant="timeline"
      isWithoutText
      className="relative z-10"
    />
    <div className="flex flex-col gap-1.5 w-full">
      <hgroup className="w-full flex items-center justify-between">
        <Skeleton height="24px" width="230px" />

        <Skeleton height="15px" width="65px" />
      </hgroup>
      <span className="flex items-center gap-2">
        <Skeleton height="15px" width="180px" />
      </span>
    </div>
  </div>
)
