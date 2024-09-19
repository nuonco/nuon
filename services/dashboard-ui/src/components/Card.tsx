import classNames from 'classnames'
import React, { type FC } from 'react'

export interface ICard extends React.HTMLAttributes<HTMLDivElement> {}

export const Card: FC<ICard> = ({ className, children, ...props }) => (
  <div
    className={classNames(
      'p-4 rounded-lg bg-gray-500/5 drop-shadow-sm flex flex-col gap-2 overflow-auto',
      {
        [`${className}`]: Boolean(className),
      }
    )}
    {...props}
  >
    {children}
  </div>
)

export const Config: FC<React.HTMLAttributes<HTMLDivElement>> = ({
  className,
  children,
  ...props
}) => {
  return (
    <div
      className={classNames(
        'flex flex-col md:flex-row flex-wrap gap-4 lg:gap-6 items-start justify-start',
        { [`${className}`]: Boolean(className) }
      )}
      {...props}
    >
      {children}
    </div>
  )
}

interface IConfigContent
  extends Omit<React.HTMLAttributes<HTMLSpanElement>, 'children'> {
  label: React.ReactNode
  value: React.ReactNode
}

export const ConfigContent: FC<IConfigContent> = ({
  className,
  label,
  value,
}) => {
  return (
    <span
      className={classNames('flex flex-col gap-2', {
        [`${className}`]: Boolean(className),
      })}
    >
      <span className="font-normal leading-normal text-sm tracking-wide text-cool-grey-600 dark:text-cool-grey-500">
        {label}
      </span>
      <span className="font-medium leading-normal text-sm tracking-wide max-w-52 break-all">
        {value}
      </span>
    </span>
  )
}
