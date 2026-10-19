import { useEffect, type ReactNode } from 'react'
import { Card } from '@/components/common/Card'
import {
  COLLECTION_VIEW_MODES,
  COLLECTION_VIEW_STORAGE_KEY,
  CollectionViewToggle,
  type TCollectionView,
} from '@/components/common/CollectionViewToggle'
import { EmptyState } from '@/components/common/EmptyState'
import { ID } from '@/components/common/ID'
import { Pagination, type IPagination } from '@/components/common/Pagination'
import { Skeleton } from '@/components/common/Skeleton'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { ComponentType } from '@/components/components/ComponentType'
import { AlignedSections } from '@/components/install-alignment/AlignedSections'
import { usePagination } from '@/hooks/use-pagination'
import { useStoredViewMode } from '@/hooks/use-stored-view-mode'
import { PaginationProvider } from '@/providers/pagination-provider'
import type { TComponentType, TInstallAlignmentCategory } from '@/types'
import { cn } from '@/utils/classnames'

export type TInstallImageListItem = {
  actions?: ReactNode
  alignment?: TInstallAlignmentCategory
  consumers?: ReactNode
  id: string
  image: ReactNode
  name: string
  status?: string
  type?: TComponentType
}

export interface IInstallImagesList {
  actions?: ReactNode
  aligned?: boolean
  drifted?: (view: TCollectionView) => ReactNode
  filtered?: boolean
  images: TInstallImageListItem[]
  loading?: boolean
  onViewChange?: (view: TCollectionView) => void
  pagination?: Omit<IPagination, 'position'>
  search?: ReactNode
  view?: TCollectionView
}

const InstallImagesListBase = ({
  actions,
  aligned = false,
  drifted,
  filtered = false,
  images,
  loading = false,
  onViewChange,
  pagination,
  search,
  view: viewProp,
}: IInstallImagesList) => {
  const { setIsPaginating } = usePagination()
  const [storedView, setStoredView] = useStoredViewMode<TCollectionView>(
    COLLECTION_VIEW_STORAGE_KEY,
    COLLECTION_VIEW_MODES,
    'list'
  )
  const view = viewProp ?? storedView
  const setView = (next: TCollectionView) => {
    onViewChange?.(next)
    if (viewProp === undefined) setStoredView(next)
  }

  useEffect(() => {
    setIsPaginating(false)
  }, [images, setIsPaginating])

  const isGrid = view === 'grid'

  const renderCards = (items: TInstallImageListItem[]) => (
    <div
      className={cn('flex flex-col gap-4', isGrid && 'md:grid md:grid-cols-2')}
    >
      {items.map((image) => (
        <Card
          key={image.id}
          className={cn('!p-4 !gap-4', isGrid && 'md:h-full')}
        >
          <div className="flex items-start justify-between gap-3 flex-wrap">
            <div className="flex flex-col gap-1.5 min-w-0">
              <div className="flex items-center gap-2 flex-wrap min-w-0">
                {image.type ? (
                  <ComponentType
                    type={image.type}
                    displayVariant="icon-only"
                    iconSize="16"
                    colorVariant="color"
                  />
                ) : null}
                <Text variant="body" weight="stronger" role="heading" level={3}>
                  {image.name}
                </Text>
                <Status
                  status={
                    aligned
                      ? image.alignment === 'in-sync'
                        ? 'success'
                        : 'unknown'
                      : image.status
                  }
                  variant="badge"
                >
                  {aligned
                    ? image.alignment === 'in-sync'
                      ? 'Synced'
                      : image.alignment === 'disabled'
                        ? 'Disabled'
                        : 'Sync status unavailable'
                    : undefined}
                </Status>
              </div>
              {aligned &&
              (!image.alignment || image.alignment === 'unknown') ? (
                <Text variant="subtext" theme="neutral">
                  Delivery of the configured version can’t be confirmed from the
                  available records.
                </Text>
              ) : null}
              <ID>{image.id}</ID>
            </div>
            {image.actions}
          </div>

          {image.consumers}

          <div className="flex flex-col gap-4 border-t pt-4">{image.image}</div>
        </Card>
      ))}
    </div>
  )

  return (
    <div className="flex flex-col gap-4 md:gap-6">
      <div className="flex flex-row flex-wrap items-center justify-between gap-4">
        <div className="flex flex-wrap items-center gap-4 w-full md:w-fit">
          {search}
        </div>
        <div className="flex items-center gap-3 ml-auto">
          <CollectionViewToggle value={view} onChange={setView} />
          {actions}
        </div>
      </div>

      {drifted?.(view)}

      {loading && !images.length ? (
        <div
          className={cn(
            'flex flex-col gap-4',
            isGrid && 'md:grid md:grid-cols-2'
          )}
        >
          <Skeleton height="12rem" width="100%" />
          <Skeleton height="12rem" width="100%" />
          {isGrid ? (
            <div className="hidden md:contents">
              <Skeleton height="12rem" width="100%" />
              <Skeleton height="12rem" width="100%" />
            </div>
          ) : null}
        </div>
      ) : images.length ? (
        aligned ? (
          <AlignedSections items={images} noun="images" render={renderCards} />
        ) : (
          renderCards(images)
        )
      ) : aligned ? (
        <EmptyState
          variant="table"
          emptyTitle="No other images"
          emptyMessage="Check Syncing and Pending sync above, or adjust the search."
        />
      ) : filtered ? (
        <EmptyState
          variant="table"
          emptyTitle="No images found"
          emptyMessage="No images match the current search."
        />
      ) : (
        <EmptyState
          variant="table"
          emptyTitle="No images configured"
          emptyMessage="Images appear here once the app config defines container image components for this install."
        />
      )}

      {pagination && (pagination.hasNext || (pagination.offset ?? 0) !== 0) ? (
        <Pagination {...pagination} />
      ) : null}
    </div>
  )
}

export const InstallImagesList = (props: IInstallImagesList) => (
  <PaginationProvider>
    <InstallImagesListBase {...props} />
  </PaginationProvider>
)
