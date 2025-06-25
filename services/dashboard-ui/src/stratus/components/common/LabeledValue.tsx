import React from 'react'
import { cn } from '@/stratus/components/helpers'
import { Text } from './Text'

export interface ILabeledValue extends React.HTMLAttributes<HTMLDivElement> {
  label: React.ReactNode
}

export const LabeledValue = ({
  children,
  className,
  label,
  ...props
}: ILabeledValue) => {
  return (
    <div className={cn('flex flex-col gap-1', className)} {...props}>
      {typeof label === 'string' ? (
        <Text variant="subtext" theme="muted">
          {label}
        </Text>
      ) : (
        label
      )}
      {children}
    </div>
  )
}
