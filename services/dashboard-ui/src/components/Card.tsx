import classNames from 'classnames'
import React, { type FC } from 'react'
import { Heading } from '@/components/Typography'

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

export interface ISection extends React.HTMLAttributes<HTMLSelectElement> {
  actions?: React.ReactNode | null
  childrenClassName?: string
  heading: React.ReactNode
  isHeadingFixed?: boolean
}

export const Section: FC<ISection> = ({
  actions,
  className,
  children,
  childrenClassName,
  heading,
  isHeadingFixed = false,
  ...props
}) => {
  return (
    <section
      className={classNames('flex flex-auto flex-col gap-4', {
        'px-6 py-8': !isHeadingFixed,
        [`${className}`]: Boolean(className),
      })}
      {...props}
    >
      <div
        className={classNames('flex items-center justify-between', {
          'px-6 pt-8 pb-4': isHeadingFixed,
        })}
      >
        <Heading>{heading}</Heading>
        <div>{actions}</div>
      </div>
      <div
        className={classNames('h-fit', {
          'px-6 overflow-auto': isHeadingFixed,
          [`${childrenClassName}`]: Boolean(childrenClassName),
        })}
      >
        {children}
      </div>
    </section>
  )
}
