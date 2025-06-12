import classNames from 'classnames'
import React, { type FC } from 'react'
import { Text } from './Text'

export interface ILabeledValue extends React.HTMLAttributes<HTMLDivElement> {
  label: React.ReactNode
}

export const LabeledValue: FC<ILabeledValue> = ({
  children,
  className,
  label,
  ...props
}) => {
  return (
    <div
      className={classNames('flex flex-col gap-1', {
        [`${className}`]: Boolean(className),
      })}
      {...props}
    >
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
