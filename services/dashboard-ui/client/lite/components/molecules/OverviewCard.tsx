import type { ReactNode } from 'react'
import { cn } from '@/utils/classnames'
import { Card, type ICard } from '../atoms/Card'
import { Text } from '../atoms/Text'

export interface IOverviewCard extends ICard {
  title: string
  footer?: ReactNode
}

export const OverviewCard = ({
  title,
  footer,
  children,
  className,
  ...props
}: IOverviewCard) => (
  <Card
    as="section"
    className={cn('flex min-h-36 flex-col gap-4', className)}
    {...props}
  >
    <Text as="h2" variant="caption" color="secondary" weight="medium">
      {title}
    </Text>
    <div className="flex flex-1 flex-col gap-2">{children}</div>
    {footer ? <div className="flex items-center gap-2">{footer}</div> : null}
  </Card>
)

export const OverviewCardGrid = ({
  children,
  columns = 4,
}: {
  children: ReactNode
  columns?: 3 | 4
}) => (
  <div className="@container">
    <div
      className={cn(
        'grid grid-cols-1 gap-4 @xl:grid-cols-2',
        columns === 3 ? '@4xl:grid-cols-3' : '@4xl:grid-cols-4'
      )}
    >
      {children}
    </div>
  </div>
)
