import type { ReactNode } from 'react'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'
import { PageHeader } from './PageHeader'

export type TSectionHeaderVariant = 'page' | 'section'

export interface ISectionHeader {
  actions?: ReactNode
  className?: string
  description?: ReactNode
  loading?: boolean
  loadingWidth?: number
  status?: ReactNode
  title: ReactNode
  variant?: TSectionHeaderVariant
}

const TITLE_PROPS = {
  page: { variant: 'h3', weight: 'stronger', level: 1 },
  section: { variant: 'base', weight: 'strong', level: 2 },
} as const

const DESCRIPTION_VARIANT = {
  page: 'body',
  section: 'subtext',
} as const

export const SectionHeader = ({
  actions,
  className,
  description,
  loading,
  loadingWidth,
  status,
  title,
  variant = 'section',
}: ISectionHeader) => {
  const content = (
    <>
      <HeadingGroup className="gap-1.5 min-w-0">
        <div className="flex flex-wrap items-center gap-2 min-w-0">
          <Text
            {...TITLE_PROPS[variant]}
            loading={loading}
            loadingWidth={loadingWidth}
          >
            {title}
          </Text>
          {status}
        </div>
        {description ? (
          <Text variant={DESCRIPTION_VARIANT[variant]} theme="neutral">
            {description}
          </Text>
        ) : null}
      </HeadingGroup>
      {actions ? (
        <div className="flex flex-wrap items-center gap-2 shrink-0">
          {actions}
        </div>
      ) : null}
    </>
  )

  if (variant === 'page') {
    return <PageHeader className={className}>{content}</PageHeader>
  }

  return (
    <div
      className={cn(
        'flex flex-wrap gap-3 w-full items-start justify-between',
        className
      )}
    >
      {content}
    </div>
  )
}
