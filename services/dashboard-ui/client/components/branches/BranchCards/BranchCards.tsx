import { useState, type ReactNode } from 'react'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { Pagination, type IPagination } from '@/components/common/Pagination'
import { Skeleton } from '@/components/common/Skeleton'
import { BranchCard, type TBranchCardData } from './BranchCard'

export type TBranchCardsLayout = 'grid' | 'stacked'

export interface IBranchCards {
  cards: TBranchCardData[]
  isLoading?: boolean
  emptyAction?: ReactNode
  pagination?: Omit<IPagination, 'position'>
  initialExpanded?: boolean
  visibleCount?: number
  layout?: TBranchCardsLayout
}

const LAYOUT_CLASSES: Record<TBranchCardsLayout, string> = {
  grid: 'grid grid-cols-1 lg:grid-cols-2 gap-4',
  stacked: 'flex flex-col gap-4',
}

export const BranchCards = ({
  cards,
  isLoading = false,
  emptyAction,
  pagination,
  initialExpanded = false,
  visibleCount = 4,
  layout = 'grid',
}: IBranchCards) => {
  const [isExpanded, setIsExpanded] = useState(initialExpanded)

  if (isLoading) {
    return (
      <div className={LAYOUT_CLASSES[layout]}>
        {Array.from({ length: 4 }).map((_, i) => (
          <Card key={i} className="gap-3 p-4">
            <Skeleton lines={3} width={['40%', '70%', '55%']} />
          </Card>
        ))}
      </div>
    )
  }

  if (cards.length === 0) {
    return (
      <Card>
        <EmptyState
          variant="diagram"
          emptyTitle="No branches yet"
          emptyMessage="Create a branch and connect a repository to start deploying this app from git."
          action={emptyAction}
        />
      </Card>
    )
  }

  return (
    <div className="flex flex-col gap-4">
      <div className={LAYOUT_CLASSES[layout]}>
        {(isExpanded ? cards : cards.slice(0, visibleCount)).map((card) => (
          <BranchCard key={card.branchId} card={card} />
        ))}
      </div>
      {cards.length > visibleCount ? (
        <div className="flex justify-center">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setIsExpanded((expanded) => !expanded)}
          >
            {isExpanded
              ? 'Show less'
              : `View more (${cards.length - visibleCount})`}
          </Button>
        </div>
      ) : null}
      {pagination && (pagination.hasNext || (pagination.offset ?? 0) > 0) ? (
        <Pagination {...pagination} />
      ) : null}
    </div>
  )
}
