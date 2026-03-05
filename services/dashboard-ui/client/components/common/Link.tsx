import React from 'react'
import { Link as RouterLink } from 'react-router'
import { cn } from '@/utils/classnames'

export type TLinkVariant = 'default' | 'ghost' | 'nav' | 'breadcrumb'

export interface ILink extends React.AnchorHTMLAttributes<HTMLAnchorElement> {
  isActive?: boolean
  isATag?: boolean
  isExternal?: boolean
  variant?: TLinkVariant
  href?: string
}

const VARIANT_CLASSES: Record<TLinkVariant, string> = {
  default: [
    'text-primary-600 dark:text-primary-500',
    'hover:text-primary-800 hover:dark:text-primary-400',
    'focus:text-primary-800 focus:dark:text-primary-400',
    'active:text-primary-900 active:dark:text-primary-600',
    'focus-visible:rounded',
    'focus-visible:px-0.5',
  ].join(' '),
  ghost: [
    'px-3 py-1 border-none rounded-md bg-inherit align-middle font-strong tracking-tight',
    'hover:bg-cool-grey-50 hover:dark:bg-white/10',
    'focus:outline-1 focus:outline-offset-0 focus:outline-primary-400/80 focus:bg-cool-grey-50 focus:dark:bg-white/10',
    'focus-visible:outline-1 focus-visible:outline-offset-0 focus-visible:outline-primary-400/80',
    'active:bg-cool-grey-100 active:dark:bg-white/15',
  ].join(' '),
  nav: [
    'flex items-center gap-4 overflow-hidden rounded-md py-2.5 px-3 transition-colors w-full',
    'text-[14px] h-[36px] leading-[21px] tracking-[-0.2px]',
    'hover:bg-black/5 hover:dark:bg-white/10',
  ].join(' '),
  breadcrumb: [
    'whitespace-nowrap break-keep',
    'hover:text-primary-800 hover:dark:text-primary-300',
  ].join(' '),
}

const NAV_ACTIVE =
  'text-primary-800 dark:text-primary-400 bg-primary-200 dark:bg-primary-600/25'
const NAV_INACTIVE = 'text-cool-grey-800 dark:text-cool-grey-400'
const BREADCRUMB_ACTIVE = 'text-primary-600 dark:text-primary-400'
const BREADCRUMB_INACTIVE = 'text-cool-grey-600 dark:text-cool-grey-400'

export const Link = ({
  className,
  children,
  href,
  isATag = false,
  isActive = false,
  isExternal = false,
  variant = 'default',
  ...props
}: ILink) => {
  const baseClasses = [
    'link',
    'font-sans',
    'transition-colors w-fit',
    'focus-visible:outline',
    'focus-visible:outline-1',
    'focus-visible:outline-offset-0',
    'focus-visible:outline-primary-400/80',
    'text-inherit font-inherit text-[inherit] leading-[inherit] tracking-[inherit]',
    'has-[svg]:flex has-[svg]:items-center has-[svg]:gap-1.5',
  ].join(' ')

  const variantStateClass =
    variant === 'nav'
      ? isActive
        ? NAV_ACTIVE
        : NAV_INACTIVE
      : variant === 'breadcrumb'
        ? isActive
          ? BREADCRUMB_ACTIVE
          : BREADCRUMB_INACTIVE
        : undefined

  const classes = cn(
    baseClasses,
    VARIANT_CLASSES[variant],
    variantStateClass,
    className
  )

  return isATag ? (
    <a className={classes} href={href as string} {...props}>
      {children}
    </a>
  ) : isExternal ? (
    <a
      className={classes}
      href={href as string}
      target="_blank"
      rel="noopener noreferrer"
      {...props}
    >
      {children}
    </a>
  ) : (
    <RouterLink className={classes} to={href ?? ''} {...(props as any)}>
      {children}
    </RouterLink>
  )
}
