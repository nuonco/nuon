import React from 'react'
import NextLink, { type LinkProps as NextLinkProps } from 'next/link'
import { cn } from '@/stratus/components/helpers'
import './Link.css'

export type TLinkVariant = 'default' | 'ghost' | 'nav' | 'breadcrumb'

export interface ILink
  extends Omit<React.AnchorHTMLAttributes<HTMLAnchorElement>, 'href'>,
    Partial<NextLinkProps> {
  isActive?: boolean
  isExternal?: boolean
  variant?: TLinkVariant
}

export const Link = ({
  className,
  children,
  href,
  isActive = false,
  isExternal = false,
  variant = 'default',
  ...props
}: ILink) => {
  const classes = cn(
    'link',
    variant,
    {
      active: isActive,
      inactive: !isActive,
    },
    className
  )

  return isExternal ? (
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
    <NextLink className={classes} href={href} {...props}>
      {children}
    </NextLink>
  )
}
