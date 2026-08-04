import type { ReactNode } from 'react'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { Pagination, type IPagination } from '@/components/common/Pagination'
import { Skeleton } from '@/components/common/Skeleton'
import { BranchCard, type TBranchCardData } from './BranchCard'

export interface IBranchCards {
  cards: TBranchCardData[]
  isLoading?: boolean
  emptyAction?: ReactNode
  pagination?: Omit<IPagination, 'position'>
}

export const BranchCards = ({
  cards,
  isLoading = false,
  emptyAction,
  pagination,
}: IBranchCards) => {
  if (isLoading) {
    return (
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
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
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {cards.map((card) => (
          <BranchCard key={card.branchId} card={card} />
        ))}
      </div>
      {pagination && (pagination.hasNext || (pagination.offset ?? 0) > 0) ? (
        <Pagination {...pagination} />
      ) : null}
    </div>
  )
}
