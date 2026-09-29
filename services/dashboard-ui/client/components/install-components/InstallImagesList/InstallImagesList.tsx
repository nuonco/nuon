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
import { usePagination } from '@/hooks/use-pagination'
import { useStoredViewMode } from '@/hooks/use-stored-view-mode'
import { PaginationProvider } from '@/providers/pagination-provider'
import type { TComponentType } from '@/types'
import { cn } from '@/utils/classnames'

export type TInstallImageListItem = {
  actions?: ReactNode
  id: string
  image: ReactNode
  name: string
  status?: string
  type?: TComponentType
}

export interface IInstallImagesList {
  actions?: ReactNode
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
    'list',
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

      {loading && !images.length ? (
        <div
          className={cn(
            'flex flex-col gap-4',
            isGrid && 'md:grid md:grid-cols-2',
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
        <div
          className={cn(
            'flex flex-col gap-4',
            isGrid && 'md:grid md:grid-cols-2',
          )}
        >
          {images.map((image) => (
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
                    <Text
                      variant="body"
                      weight="stronger"
                      role="heading"
                      level={3}
                    >
                      {image.name}
                    </Text>
                    <Status status={image.status} variant="badge" />
                  </div>
                  <ID>{image.id}</ID>
                </div>
                {image.actions}
              </div>

              <div className="flex flex-col gap-4 border-t pt-4">
                {image.image}
              </div>
            </Card>
          ))}
        </div>
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
