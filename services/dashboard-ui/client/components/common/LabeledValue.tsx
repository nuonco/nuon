import React from 'react'
import { cn } from '@/utils/classnames'
import { Text } from './Text'

export interface ILabeledValue extends React.HTMLAttributes<HTMLDivElement> {
  label: React.ReactNode
  loading?: boolean
  loadingWidth?: number
}

export const LabeledValue = ({
  children,
  className,
  label,
  loading,
  loadingWidth,
  ...props
}: ILabeledValue) => {
  return (
    <div className={cn('flex flex-col gap-1', className)} {...props}>
      {typeof label === 'string' ? (
        <Text variant="subtext" theme="neutral">
          {label}
        </Text>
      ) : (
        label
      )}
      {loading ? (
        <Text variant="subtext" loading loadingWidth={loadingWidth} />
      ) : typeof children === 'string' ? (
        <Text variant="subtext">{children}</Text>
      ) : (
        children
      )}
    </div>
  )
}
