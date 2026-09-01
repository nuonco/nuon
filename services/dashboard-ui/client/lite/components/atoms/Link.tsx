import type { AnchorHTMLAttributes } from 'react'
import { Link as RouterLink } from 'react-router'
import { cn } from '@/utils/classnames'
import { Icon } from './Icon'
import { Text, type TTextVariant } from './Text'

export interface ILink extends AnchorHTMLAttributes<HTMLAnchorElement> {
  href: string
  variant?: TTextVariant
  external?: boolean
  reloadDocument?: boolean
  disabled?: boolean
  loading?: boolean
  loadingWidth?: number
}

const VARIANT_CLASSES: Record<TTextVariant, string> = {
  display: 'text-display',
  title: 'text-title',
  heading: 'text-heading',
  body: 'text-body',
  caption: 'text-caption',
  label: 'text-label',
}

const EXTERNAL_HREF = /^([a-z][a-z0-9+.-]*:|\/\/)/i

export const isExternalHref = (href: string) => EXTERNAL_HREF.test(href)

const LINK_CLASSES =
  'text-link underline decoration-link/35 underline-offset-2 outline-none transition-colors ' +
  'hover:text-link-hover hover:decoration-link-hover/60 ' +
  'active:text-link-active ' +
  'focus-visible:rounded-xs focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-focus-ring'

const DISABLED_CLASSES = 'text-tertiary no-underline cursor-not-allowed'

export const Link = ({
  href,
  variant,
  external,
  reloadDocument,
  disabled = false,
  loading = false,
  loadingWidth,
  className,
  children,
  ...props
}: ILink) => {
  if (loading) {
    return (
      <Text variant={variant} loading loadingWidth={loadingWidth} className={className} />
    )
  }

  const sizing = variant ? VARIANT_CLASSES[variant] : undefined

  if (disabled) {
    return (
      <span aria-disabled className={cn(sizing, DISABLED_CLASSES, className)} >
        {children}
      </span>
    )
  }

  const isExternal = external ?? isExternalHref(href)

  if (isExternal) {
    return (
      <a
        href={href}
        target="_blank"
        rel="noopener noreferrer"
        className={cn('inline-flex items-center gap-1', sizing, LINK_CLASSES, className)}
        {...props}
      >
        {children}
        <Icon variant="ArrowSquareOutIcon" size="1em" aria-hidden />
        <span className="sr-only">(opens in a new tab)</span>
      </a>
    )
  }

  return (
    <RouterLink
      to={href}
      reloadDocument={reloadDocument}
      className={cn(sizing, LINK_CLASSES, className)}
      {...props}
    >
      {children}
    </RouterLink>
  )
}
