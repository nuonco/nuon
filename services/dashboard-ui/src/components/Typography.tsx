import classNames from 'classnames'
import React, { type FC } from 'react'
import { ClickToCopy } from '@/components/ClickToCopy'

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
      className={classNames(
        'font-sans flex flex-wrap gap-1 items-center word-wrap',
        {
          'text-base font-semibold leading-normal': variant === 'heading',
          'text-base font-semibold': variant === 'subheading',
          'text-xl font-semibold leading-loose tracking-normal':
            variant === 'title',
          'text-3xl font-bold': variant === 'subtitle',
          [`${className}`]: Boolean(className),
        }
      )}
      {...props}
    >
      {children}
    </span>
  )
}

export interface IOldText extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: 'base' | 'caption' | 'label' | 'overline' | 'status' | 'id' | 'mono'
}

export const OldText: FC<IOldText> = ({
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
        ['text-base leading-normal tracking-normal font-normal']:
          variant === 'base',
        ['tracking-wide text-sm font-semibold uppercase leading-none word-wrap']:
          isStatus,
        ['text-sm tracking-wide leading-none text-gray-600 dark:text-gray-300']:
          isOverline,
        ['text-sm font-medium leading-tight']: isLabel,
        'text-sm': variant === 'caption',
        'font-mono text-sm tracking-wided font-normal text-cool-grey-600 dark:text-white/70':
          variant === 'id',
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
  variant?: 'default' | 'preformated' | 'inline'
}

export const Code: FC<ICode> = ({
  className,
  children,
  variant = 'default',
}) => {
  const classes = classNames(
    'text-sm p-4 bg-cool-grey-100 text-blue-800 dark:bg-dark-grey-200 dark:text-blue-500 font-mono break-all flex flex-col rounded shadow-sm min-h-[3rem] max-h-[40rem] max-w-5xl overflow-auto',
    {
      '!p-1 leading-3 min-h-min overflow-x-scroll': variant === 'inline',
      [`${className}`]: Boolean(className),
    }
  )

  return variant === 'preformated' ? (
    <pre className={classes}>{children}</pre>
  ) : (
    <code className={classes}>
      {variant === 'inline' ? (
        <span className="block min-w-max">{children}</span>
      ) : (
        children
      )}
    </code>
  )
}

export const CodeInline: FC<ICode> = ({ className, children }) => {
  const classes = classNames(
    'text-sm bg-cool-grey-50 text-blue-800 dark:bg-dark-grey-200 text-blue-500 font-mono break-all rounded-lg shadow-sm py-0.5 px-2 leading-3 border',
    {
      [`${className}`]: Boolean(className),
    }
  )

  return <code className={classes}>{children}</code>
}

export interface ITruncate extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: 'default' | 'small' | 'large' | 'extra-large'
}

export const Truncate: FC<ITruncate> = ({
  className,
  children,
  variant = 'default',
  ...props
}) => {
  return (
    <span
      className={classNames('truncate inline', {
        'max-w-[130px]': variant === 'default',
        'max-w-[70px]': variant === 'small',
        'max-w-[180px]': variant === 'large',
        'max-w-[280px]': variant === 'extra-large',
        [`${className}`]: Boolean(className),
      })}
      {...props}
    >
      {children}
    </span>
  )
}

export type TTextVariant =
  | 'reg-12'
  | 'reg-14'
  | 'mono-12'
  | 'mono-14'
  | 'med-8'
  | 'med-12'
  | 'med-14'
  | 'med-18'
  | 'semi-14'
  | 'semi-18'

export interface IText extends React.HTMLAttributes<HTMLSpanElement> {
  level?: 1 | 2 | 3 | 4 | 5 | 6
  role?: 'paragraph' | 'heading' | 'code' | 'time'
  variant?: TTextVariant
  isMuted?: boolean
}

export const Text: FC<IText> = ({
  className,
  children,
  level,
  role = 'paragraph',
  variant = 'reg-12',
  isMuted = false,
  ...props
}) => {
  return (
    <span
      aria-level={role === 'heading' && level ? level : undefined}
      className={classNames('font-sans flex flex-wrap items-center gap-1', {
        'font-stronger text-lg leading-loose tracking-normal':
          variant === 'semi-18',
        'font-stronger text-sm': variant === 'semi-14',
        'font-strong text-lg': variant === 'med-18',
        'font-strong text-base': variant === 'med-14',
        'font-strong text-xs tracking-wide': variant === 'med-12',
        'font-strong text-[10px] leading-tight tracking-wide': variant === 'med-8',
        'text-sm leading-normal': variant === 'reg-14',
        'text-xs leading-normal tracking-wide': variant === 'reg-12',
        '!font-mono font-normal text-sm leading-loose': variant === 'mono-14',
        '!font-mono font-normal text-xs leading-relaxed text-cool-grey-600 dark:text-white/70':
          variant === 'mono-12',
        'text-cool-grey-600 dark:text-white/70': isMuted,
        [`${className}`]: Boolean(className),
      })}
      role={role}
      {...props}
    >
      {children}
    </span>
  )
}

export const ID: FC<{
  id: React.ReactElement | string
  className?: string
}> = ({ className, id }) => (
  <Text variant="mono-12" className={className}>
    <ClickToCopy>{id}</ClickToCopy>
  </Text>
)
