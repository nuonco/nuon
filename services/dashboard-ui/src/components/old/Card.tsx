import classNames from 'classnames'
import React, { type FC } from 'react'
import { Text } from '@/components/old/Typography'

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

export interface ISection extends React.HTMLAttributes<HTMLDivElement> {
  actions?: React.ReactNode | null
  childrenClassName?: string
  heading?: React.ReactNode
  headingClassName?: string
  isHeadingFixed?: boolean
}

export const Section: FC<ISection> = ({
  actions,
  className,
  children,
  childrenClassName,
  heading,
  headingClassName,
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
      {heading || actions ? (
        <SectionHeader
          actions={actions}
          heading={heading}
          className={headingClassName}
          isHeadingFixed={isHeadingFixed}
        />
      ) : null}
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

export const SectionHeader: FC<
  Omit<ISection, 'children' | 'childrenClassName'>
> = ({ actions, className, heading, isHeadingFixed }) => {
  return (
    <div
      className={classNames('flex items-center justify-between', {
        'px-6 pt-8 pb-4': isHeadingFixed,
        [`${className}`]: Boolean(className),
      })}
    >
      <Text variant="semi-14">{heading}</Text>
      <div>{actions}</div>
    </div>
  )
}
