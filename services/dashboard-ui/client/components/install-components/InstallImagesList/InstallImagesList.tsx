import { useEffect, type ReactNode } from 'react'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { ID } from '@/components/common/ID'
import { Pagination, type IPagination } from '@/components/common/Pagination'
import { Skeleton } from '@/components/common/Skeleton'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { ComponentType } from '@/components/components/ComponentType'
import { usePagination } from '@/hooks/use-pagination'
import { PaginationProvider } from '@/providers/pagination-provider'
import type { TComponentType } from '@/types'

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
  pagination?: Omit<IPagination, 'position'>
  search?: ReactNode
}

const InstallImagesListBase = ({
  actions,
  filtered = false,
  images,
  loading = false,
  pagination,
  search,
}: IInstallImagesList) => {
  const { setIsPaginating } = usePagination()

  useEffect(() => {
    setIsPaginating(false)
  }, [images, setIsPaginating])

  return (
    <div className="flex flex-col gap-4 md:gap-6">
      {search || actions ? (
        <div className="flex flex-row flex-wrap items-center justify-between gap-4">
          <div className="flex flex-wrap items-center gap-4 w-full md:w-fit">
            {search}
          </div>
          {actions}
        </div>
      ) : null}

      {loading && !images.length ? (
        <div className="flex flex-col gap-4">
          <Skeleton height="12rem" width="100%" />
          <Skeleton height="12rem" width="100%" />
        </div>
      ) : images.length ? (
        <div className="flex flex-col gap-4">
          {images.map((image) => (
            <Card key={image.id} className="!p-4 !gap-4">
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
