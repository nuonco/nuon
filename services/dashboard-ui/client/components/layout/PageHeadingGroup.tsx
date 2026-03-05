import React from 'react'
import { BackLink } from '@/components/common/BackLink'
import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'

interface IPageHeadingGroup {
  title: React.ReactNode
  subtitle?: React.ReactNode
  showBackLink?: boolean
  headingLevel?: 1 | 2 | 3 | 4 | 5 | 6
  titleProps?: React.ComponentProps<typeof Text>
  subtitleProps?: React.ComponentProps<typeof Text>
  parentClassName?: string
  className?: string
}

export const PageHeadingGroup: React.FC<IPageHeadingGroup> = ({
  title,
  subtitle,
  showBackLink = false,
  headingLevel = 1,
  titleProps,
  subtitleProps,
  parentClassName = '',
  className = '',
}) => (
  <div className={cn('flex flex-col gap-4', parentClassName)}>
    {showBackLink && <BackLink />}
    <hgroup className={cn('flex flex-col gap-0.5', className)}>
      <Text variant="h3" weight="stronger" level={headingLevel} {...titleProps}>
        {title}
      </Text>
      {subtitle && (
        <Text variant="subtext" theme="neutral" {...subtitleProps}>
          {subtitle}
        </Text>
      )}
    </hgroup>
  </div>
)
