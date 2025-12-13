import React, { forwardRef } from 'react'
import Link from 'next/link'
import { cn } from '@/utils/classnames'

export type TButtonSize = 'lg' | 'md' | 'sm' | 'xs'
export type TButtonVariant = 'danger' | 'ghost' | 'primary' | 'secondary'

interface IButtonBase {
  size?: TButtonSize
  variant?: TButtonVariant
  href?: string
}

export interface IButtonAsButton
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    IButtonBase {
  href?: undefined
}

export interface IButtonAsAnchor
  extends React.AnchorHTMLAttributes<HTMLAnchorElement>,
    IButtonBase {
  href: string
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
    border rounded-md bg-dark-grey-900 text-red-500
    hover:bg-[#1D0D10]
    focus:outline-red-500/50
    focus:bg-dark-grey-900
    active:outline-red-500/50
    active:bg-[#2E1013]
    disabled:opacity-50 disabled:hover:bg-dark-grey-700
  `,
  primary: `
    border border-transparent rounded-md bg-primary-600 text-white
    hover:bg-primary-700
    focus:outline-primary-400/80 focus:bg-primary-600
    active:bg-primary-900
    disabled:opacity-50 disabled:hover:bg-primary-600
  `,
  ghost: `
    border border-transparent rounded-md bg-inherit
    hover:bg-white/5
    focus:outline-primary-400/80 focus:bg-white/5
    active:bg-white/10
    disabled:opacity-50 disabled:hover:bg-transparent
  `,
  secondary: `
    border rounded-md bg-dark-grey-700 text-primary-400
    hover:bg-dark-grey-500
    focus:outline-primary-400/80 focus:bg-dark-grey-500
    active:bg-dark-grey-400
    disabled:opacity-50 disabled:hover:bg-dark-grey-700
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
      ...props
    },
    ref
  ) => {
    const classes = cn(
      `inline-flex items-center font-sans font-medium tracking-tight transition-colors whitespace-nowrap break-keep w-fit focus:outline-1 focus:outline-current cursor-pointer
      disabled:cursor-not-allowed`,
      VARIANT_CLASSES[variant],
      SIZE_CLASSES[size],
      'has-[svg]:flex has-[svg]:items-center has-[svg]:gap-1.5',
      className
    )

    if (href) {
      const isInternal = href.startsWith('/')
      if (isInternal) {
        return (
          <Link
            prefetch={false}
            href={href}
            className={classes}
            ref={ref as React.Ref<HTMLAnchorElement>}
            {...(props as React.AnchorHTMLAttributes<HTMLAnchorElement>)}
          >
            {children}
          </Link>
        )
      }
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
