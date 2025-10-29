import classNames from 'classnames'
import NextLink from 'next/link'
import { type FC, HTMLAttributes } from 'react'

export interface ILink extends HTMLAttributes<HTMLAnchorElement> {
  href: string
  target?: string
  isActive?: boolean
  isExternal?: boolean
  variant?: 'default' | 'breadcrumb' | 'ghost'
}

export const Link: FC<ILink> = ({
  className,
  children,
  href,
  isActive = false,
  isExternal = false,
  variant = 'default',
  ...props
}) => {
  return (
    <NextLink
      className={classNames('flex flex-initial items-center w-fit gap-1', {
        'text-primary-700 hover:text-primary-600 focus:text-primary-600 active:text-primary-800 dark:text-primary-500 dark:hover:text-primary-400 dark:focus:text-primary-400 dark:active:text-primary-600':
          variant === 'default',
        'text-cool-grey-600 hover:text-primary-600 focus:text-primary-600 active:text-primary-800 dark:text-cool-grey-400 dark:hover:text-primary-400 dark:focus:text-primary-400 dark:active:text-primary-600':
          variant === 'breadcrumb',
        'text-cool-grey-950 dark:text-white':
          variant === 'breadcrumb' && isActive,
        'p-2 rounded-md text-sm font-medium text-cool-grey-6000 hover:bg-black/5 dark:hover:bg-white/5':
          variant === 'ghost',
        [`${className}`]: Boolean(className),
      })}
      prefetch={false}
      href={href}
      {...props}
    >
      {children}
    </NextLink>
  )
}
