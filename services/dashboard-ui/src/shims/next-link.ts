import React from 'react'
import { Link as RouterLink } from 'react-router-dom'

interface NextLinkProps {
  href: any
  children?: React.ReactNode
  className?: string
  prefetch?: boolean
  scroll?: boolean
  replace?: boolean
  target?: string
  rel?: string
  [key: string]: any
}

const Link = React.forwardRef<HTMLAnchorElement, NextLinkProps>(
  ({ href, prefetch, scroll, replace, ...props }, ref) => {
    const to = typeof href === 'string' ? href : href?.pathname || '/'

    if (to.startsWith('http://') || to.startsWith('https://')) {
      return React.createElement('a', { ref, href: to, ...props })
    }

    return React.createElement(RouterLink, {
      ref,
      to,
      replace,
      ...props,
    })
  }
)

Link.displayName = 'Link'

export default Link
export type { NextLinkProps as LinkProps }
