import React, { forwardRef } from 'react'
import { Link } from 'react-router'
import { cn } from '@/utils/classnames'

export type TButtonSize = 'lg' | 'md' | 'sm' | 'xs'
export type TButtonVariant =
  | 'danger'
  | 'ghost'
  | 'primary'
  | 'secondary'
  | 'tab'

interface IButtonBase {
  size?: TButtonSize
  variant?: TButtonVariant
  href?: string
  isActive?: boolean
  isMenuButton?: boolean
}

export interface IButtonAsButton
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    IButtonBase {
  href?: undefined
  isAnchorTag?: undefined
}

export interface IButtonAsAnchor
  extends React.AnchorHTMLAttributes<HTMLAnchorElement>,
    IButtonBase {
  href: string
  isAnchorTag?: boolean
}

export type TButton = IButtonAsButton | IButtonAsAnchor

const SIZE_CLASSES: Record<TButtonSize, string> = {
  lg: 'text-sm h-9 px-3 py-1 leading-[21px]',
  md: 'text-sm h-8 px-3 py-1 leading-[21px]',
  sm: 'text-xs h-6 px-2 py-0.5 leading-[15px]',
  xs: 'text-xs h-4 leading-[15px]',
}

const VARIANT_CLASSES: Record<TButtonVariant, string> = {
  danger: `
    border rounded-lg bg-white dark:bg-dark-grey-900 text-red-800 dark:text-red-500
    hover:bg-red-50 dark:hover:bg-[#1D0D10]
    focus:outline-red-400 dark:focus:outline-red-500/50
    focus:bg-white dark:focus:bg-dark-grey-900
    active:outline-red-400 dark:active:outline-red-500/50
    active:bg-red-100 dark:active:bg-[#2E1013]
    disabled:opacity-50 disabled:hover:bg-white disabled:hover:dark:bg-dark-grey-700
  `,
  primary: `
    border border-transparent rounded-lg bg-primary-600 text-white
    hover:bg-primary-700
    focus:outline-primary-400/80 focus:bg-primary-600
    active:bg-primary-900
    disabled:opacity-50 disabled:hover:bg-primary-600
  `,
  ghost: `
    border border-transparent rounded-lg bg-inherit
    hover:bg-cool-grey-500/8 dark:hover:bg-cool-grey-500/8
    focus:outline-none focus:shadow-[0_0_0_1px_white,0_0_0_3px_rgba(128,64,191,0.64)] dark:focus:shadow-[0_0_0_1px_#141217,0_0_0_3px_rgba(128,64,191,0.64)]
    active:bg-cool-grey-500/16 dark:active:bg-cool-grey-500/16
    disabled:opacity-50 disabled:hover:bg-inherit disabled:hover:dark:bg-inherit
  `,
  secondary: `
    border rounded-lg bg-white dark:bg-dark-grey-700 text-primary-600 dark:text-primary-400
    shadow-[0px_1px_2px_0px_rgba(0,0,0,0.08)]
    hover:bg-cool-grey-50 dark:hover:bg-dark-grey-500
    focus:outline-none focus:shadow-[0_0_0_1px_white,0_0_0_3px_rgba(128,64,191,0.64)] dark:focus:shadow-[0_0_0_1px_#141217,0_0_0_3px_rgba(128,64,191,0.64)]
    focus:bg-white dark:focus:bg-dark-grey-700
    active:bg-cool-grey-100 dark:active:bg-dark-grey-400
    disabled:opacity-50 disabled:hover:bg-white disabled:hover:dark:bg-dark-grey-700
  `,
  tab: `
    border-b-3 border-transparent bg-transparent dark:bg-transparent text-primary-600 dark:text-primary-400
    hover:!border-primary-600/50
    focus:outline-none focus:bg-cool-grey-50/20 dark:focus:bg-dark-grey-500/20
    active:!border-primary-600/80
    disabled:opacity-50 disabled:hover:bg-white disabled:hover:dark:bg-dark-grey-700 disabled:hover:!border-transparent
  `,
}

export const Button = forwardRef<
  HTMLButtonElement | HTMLAnchorElement,
  TButton
>(
  (
    {
      className,
      children,
      size = 'md',
      variant = 'secondary',
      href,
      isActive,
      isAnchorTag = false,
      isMenuButton,
      ...props
    },
    ref
  ) => {
    const classes = cn(
      `inline-flex items-center font-sans font-strong tracking-tight transition-colors whitespace-nowrap break-keep w-fit focus:outline-1 focus:outline-current cursor-pointer
      disabled:cursor-not-allowed`,
      VARIANT_CLASSES[variant],
      SIZE_CLASSES[size],
      'has-[svg]:flex has-[svg]:items-center has-[svg]:gap-1.5',
      {
        '!border-primary-600 !hover:!border-primary-600':
          isActive && variant === 'tab',
        '!p-2 text-sm !leading-none h-8 w-full flex justify-between !rounded-md !text-cool-grey-800 dark:!text-white/70':
          isMenuButton,
      },
      className
    )

    if (href) {
      if (isAnchorTag) {
        return (
          <a
            ref={ref as React.Ref<HTMLAnchorElement>}
            className={classes}
            href={href}
            {...(props as React.AnchorHTMLAttributes<HTMLAnchorElement>)}
          >
            {children}
          </a>
        )
      }

      const isInternal = href.startsWith('/')
      if (isInternal) {
        return (
          <Link
            to={href}
            className={classes}
            ref={ref as React.Ref<HTMLAnchorElement>}
            {...(props as React.AnchorHTMLAttributes<HTMLAnchorElement>)}
          >
            {children}
          </Link>
        )
      }
      // External link
      return (
        <a
          ref={ref as React.Ref<HTMLAnchorElement>}
          className={classes}
          href={href}
          target="_blank"
          rel="noopener noreferrer"
          {...(props as React.AnchorHTMLAttributes<HTMLAnchorElement>)}
        >
          {children}
        </a>
      )
    }

    // Regular button
    return (
      <button
        ref={ref as React.Ref<HTMLButtonElement>}
        className={classes}
        {...(props as React.ButtonHTMLAttributes<HTMLButtonElement>)}
      >
        {children}
      </button>
    )
  }
)

Button.displayName = 'Button'
