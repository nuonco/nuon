import classNames from 'classnames'
import React, { type FC } from 'react'

export type THeadingVariant = 'heading' | 'subheading' | 'title' | 'subtitle'

export interface IHeading extends React.HTMLAttributes<HTMLSpanElement> {
  level?: 1 | 2 | 3 | 4 | 5 | 6
  variant?: THeadingVariant
}

export const Heading: FC<IHeading> = ({
  className,
  children,
  level = 3,
  variant = 'heading',
  ...props
}) => {
  return (
    <span
      aria-level={level}
      className={classNames('flex flex-wrap gap-1 items-center word-wrap', {
        'text-xl font-bold': variant === 'heading',
        'text-md font-semibold': variant === 'subheading',
        'text-5xl font-semibold': variant === 'title',
        'text-3xl font-bold': variant === 'subtitle',
        [`${className}`]: Boolean(className),
      })}
      {...props}
    >
      {children}
    </span>
  )
}

export interface IText extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: 'base' | 'caption' | 'label' | 'overline' | 'status'
}

export const Text: FC<IText> = ({
  children,
  className,
  variant = 'base',
  ...props
}) => {
  const isLabel = variant === 'label'
  const isOverline = variant === 'overline'
  const isStatus = variant === 'status'

  return (
    <span
      className={classNames('flex flex-wrap items-center gap-1', {
        ['tracking-wider text-xs font-semibold uppercase leading-none word-wrap']:
          isStatus,
        ['text-xs tracking-wide leading-none text-gray-600 dark:text-gray-300']:
          isOverline,
        ['text-sm font-semibold']: isLabel,
        'text-xs': variant === 'caption',
        [`${className}`]: Boolean(className),
      })}
      role="paragraph"
      {...props}
    >
      {children}
    </span>
  )
}

export interface ICode extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: 'default' | 'preformated'
}

export const Code: FC<ICode> = ({
  className,
  children,
  variant = 'default',
}) => {
  const classes = classNames(
    'text-xs p-6 bg-gray-800 text-gray-100 font-mono break-all block rounded shadow-sm min-h-[3rem] max-h-[40rem] max-w-5xl overflow-auto',
    { [`${className}`]: Boolean(className) }
  )

  return variant === 'preformated' ? (
    <pre className={classes}>{children}</pre>
  ) : (
    <span className={classes}>{children}</span>
  )
}
