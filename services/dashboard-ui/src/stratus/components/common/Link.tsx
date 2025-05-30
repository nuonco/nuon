import classNames from 'classnames'
import React, { type FC } from 'react'
import NextLink, { type LinkProps as NextLinkProps } from 'next/link'
import './Link.css'

export type TLinkVariant = 'default' | 'ghost' | 'nav' | 'breadcrumb'

export interface ILink
  extends Omit<React.AnchorHTMLAttributes<HTMLAnchorElement>, 'href'>,
    Partial<NextLinkProps> {
  isActive?: boolean
  isExternal?: boolean
  variant?: TLinkVariant
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
  const classes = classNames(`link ${variant}`, {
    active: isActive,
    inactive: !isActive,
    [`${className}`]: Boolean(className),
  })

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
