import { useEffect, useState, type ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import type { TCollectionView } from '@/components/common/CollectionViewToggle'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { SectionHeader } from '@/components/layout/SectionHeader'
import type { TInstallAlignmentResource } from '@/types'
import { cn } from '@/utils/classnames'

export const OngoingDeployments = ({
  resources,
  renderCard,
  view,
  pageSize,
  search,
}: {
  resources: TInstallAlignmentResource[]
  renderCard: (resource: TInstallAlignmentResource) => ReactNode
  view: TCollectionView
  pageSize: number
  search?: string
}) => {
  const [page, setPage] = useState(0)
  useEffect(() => setPage(0), [search])
  if (!resources.length) return null

  const lastPage = Math.ceil(resources.length / pageSize) - 1
  const currentPage = Math.min(page, lastPage)
  const start = currentPage * pageSize
  const imagesOnly = resources.every((resource) => resource.kind === 'image')
  const title = imagesOnly ? 'Syncing' : 'Ongoing deployments'

  return (
    <section aria-label={title} data-testid="ongoing-section">
      <Card className="!p-3 !gap-3">
        <SectionHeader
          title={
            <span className="flex items-center gap-2">
              <Icon variant="ArrowsClockwiseIcon" size={16} />
              {title}
            </span>
          }
          status={
            <Badge size="sm" theme="info">
              {resources.length}
            </Badge>
          }
          description={
            imagesOnly
              ? 'Image syncs queued, running, or waiting for action.'
              : 'Deployments queued, running, or waiting for action.'
          }
        />
        <div
          className={cn(
            view === 'grid'
              ? 'flex flex-col gap-2 md:grid md:grid-cols-2 md:items-start'
              : 'flex flex-col divide-y -mx-3'
          )}
          data-testid="ongoing-cards"
          data-view={view}
        >
          {resources.slice(start, start + pageSize).map((resource) => (
            <div
              key={resource.id}
              className={cn('min-w-0', view === 'list' && 'px-3')}
            >
              {renderCard(resource)}
            </div>
          ))}
        </div>
        {resources.length > pageSize ? (
          <div className="flex flex-wrap gap-3 items-center justify-between">
            <Text variant="subtext" aria-live="polite">
              Showing {start + 1}–{Math.min(start + pageSize, resources.length)}{' '}
              of {resources.length}
            </Text>
            <div className="flex gap-2">
              <Button
                disabled={currentPage === 0}
                onClick={() => setPage(currentPage - 1)}
                aria-label="Previous ongoing page"
              >
                <Icon variant="CaretLeftIcon" size={14} />
                Previous
              </Button>
              <Button
                disabled={currentPage >= lastPage}
                onClick={() => setPage(currentPage + 1)}
                aria-label="Next ongoing page"
              >
                Next
                <Icon variant="CaretRightIcon" size={14} />
              </Button>
            </div>
          </div>
        ) : null}
      </Card>
    </section>
  )
}
