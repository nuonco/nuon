import React from 'react'
import type { TKeyValue } from '@/types'
import { cn } from '@/utils/classnames'
import { CodeBlock } from './CodeBlock'
import { EmptyState, type IEmptyState } from './EmptyState'
import { Skeleton } from './Skeleton'
import { Text } from './Text'

export interface IKeyValueList extends React.HTMLAttributes<HTMLDivElement> {
  emptyStateProps?: IEmptyState
  values: TKeyValue[]
}

export const KeyValueList = ({
  className,
  emptyStateProps = { variant: 'table', size: 'sm' },
  values,
  ...props
}: IKeyValueList) => {
  return values?.length ? (
    <div
      className={cn('grid grid-cols-[max-content_1fr] gap-0', className)}
      {...props}
    >
      {/* Header row */}
      <Text className="py-2 border-b" variant="subtext" theme="neutral">
        Name
      </Text>
      <Text className="py-2 pl-8 border-b" variant="subtext" theme="neutral">
        Value
      </Text>

      {/* Data rows */}
      {values.map(({ key, value, type }, index) => {
        const isLast = index === values.length - 1

        return (
          <React.Fragment key={key}>
            <Text
              className={cn(
                'py-2 break-all whitespace-nowrap',
                !isLast && 'border-b'
              )}
              variant="subtext"
              family="mono"
            >
              {key}
            </Text>
            <Text
              className={cn(
                'block py-2 pl-8 break-all !w-full overlfow-x-auto',
                !isLast && 'border-b'
              )}
              variant="subtext"
              family="mono"
            >
              {value ? (
                type === 'object' || type === 'array' ? (
                  <CodeBlock className="!w-full !overflow-auto" language="json">
                    {value}
                  </CodeBlock>
                ) : (
                  value
                )
              ) : (
                <Text variant="subtext" theme="neutral">
                  —
                </Text>
              )}
            </Text>
          </React.Fragment>
        )
      })}
    </div>
  ) : (
    <EmptyState variant="table" size="sm" {...emptyStateProps} />
  )
}

export const KeyValueListSkeleton = ({ count = 5 }) => {
  return (
    <div className="grid grid-cols-[max-content_1fr] gap-0">
      {/* Header */}
      <Text className="py-2 border-b" variant="subtext" theme="neutral">
        Name
      </Text>
      <Text className="py-2 pl-8 border-b" variant="subtext" theme="neutral">
        Value
      </Text>

      {/* Skeleton rows */}
      {Array.from({ length: count }).map((_, idx) => {
        const isLast = idx === count - 1

        return (
          <React.Fragment key={idx}>
            <div className={cn('py-2', !isLast && 'border-b')}>
              <Skeleton height="17px" width="120px" />
            </div>
            <div className={cn('py-2 pl-8', !isLast && 'border-b')}>
              <Skeleton height="17px" width="60%" />
            </div>
          </React.Fragment>
        )
      })}
    </div>
  )
}
