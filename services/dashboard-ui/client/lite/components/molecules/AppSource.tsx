import type { HTMLAttributes, ReactNode } from 'react'
import type { TApp } from '@/types'
import { cn } from '@/utils/classnames'
import { Icon } from '../atoms/Icon'
import { Link } from '../atoms/Link'
import { Text, type TTextColor, type TTextVariant } from '../atoms/Text'

export interface IAppSource
  extends Omit<HTMLAttributes<HTMLSpanElement>, 'children'> {
  source?: string | null
  href?: string
  variant?: TTextVariant
  color?: TTextColor
  iconSize?: number
  fallback?: ReactNode
  loading?: boolean
  loadingWidth?: number
}

export const appSourceFromApp = (app?: TApp | null) =>
  app?.sandbox_config?.public_git_vcs_config?.repo ??
  app?.sandbox_config?.connected_github_vcs_config?.repo ??
  app?.config_repo

export const AppSource = ({
  source,
  href,
  variant = 'caption',
  color = 'secondary',
  iconSize = 16,
  fallback = '—',
  loading = false,
  loadingWidth = 18,
  className,
  ...props
}: IAppSource) => {
  if (!loading && !source) {
    return (
      <Text variant={variant} color="tertiary" className={className} {...props}>
        {fallback}
      </Text>
    )
  }

  return (
    <span
      className={cn('inline-flex min-w-0 items-center gap-1.5', className)}
      {...props}
    >
      <Icon variant="GitHub" size={iconSize} className="shrink-0" aria-hidden />
      {href && source ? (
        <Link
          href={href}
          external
          variant={variant}
          className="min-w-0 truncate font-mono"
        >
          {source}
        </Link>
      ) : (
        <Text
          variant={variant}
          family="mono"
          color={color}
          loading={loading}
          loadingWidth={loadingWidth}
          className="truncate"
        >
          {source}
        </Text>
      )}
    </span>
  )
}
