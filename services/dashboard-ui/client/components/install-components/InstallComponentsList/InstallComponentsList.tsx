import { useEffect, type ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import {
  COLLECTION_VIEW_MODES,
  COLLECTION_VIEW_STORAGE_KEY,
  CollectionViewToggle,
  type TCollectionView,
} from '@/components/common/CollectionViewToggle'
import { EmptyState } from '@/components/common/EmptyState'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Pagination, type IPagination } from '@/components/common/Pagination'
import { Skeleton } from '@/components/common/Skeleton'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { ComponentType } from '@/components/components/ComponentType'
import { InstallConfigBehindBadge } from '@/components/installs/InstallResourceConfig'
import { usePagination } from '@/hooks/use-pagination'
import { useStoredViewMode } from '@/hooks/use-stored-view-mode'
import { PaginationProvider } from '@/providers/pagination-provider'
import type { TComponentType } from '@/types'
import { cn } from '@/utils/classnames'

export type TInstallComponentListItem = {
  actions?: ReactNode
  behind?: boolean
  enabled?: boolean | null
  id: string
  href?: string
  latestDeploy: ReactNode
  name: string
  status?: string
  type?: TComponentType
}

export interface IInstallComponentsList {
  actions?: ReactNode
  components: TInstallComponentListItem[]
  filterActions?: ReactNode
  filtered?: boolean
  loading?: boolean
  onViewChange?: (view: TCollectionView) => void
  pagination?: Omit<IPagination, 'position'>
  search?: ReactNode
  view?: TCollectionView
}

const InstallComponentsListBase = ({
  actions,
  components,
  filterActions,
  filtered = false,
  loading = false,
  onViewChange,
  pagination,
  search,
  view: viewProp,
}: IInstallComponentsList) => {
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
  }, [components, setIsPaginating])

  const isGrid = view === 'grid'

  return (
    <div className="flex flex-col gap-4 md:gap-6">
      <div className="flex flex-row flex-wrap items-center justify-between gap-4">
        <div className="flex flex-wrap items-center gap-4 w-full md:w-fit">
          {search}
          {filterActions}
        </div>
        <div className="flex items-center gap-3 ml-auto">
          <CollectionViewToggle value={view} onChange={setView} />
          {actions}
        </div>
      </div>

      {loading && !components.length ? (
        <div
          className={cn(
            'flex flex-col gap-4',
            isGrid && 'md:grid md:grid-cols-2'
          )}
        >
          <Skeleton height="14rem" width="100%" />
          <Skeleton height="14rem" width="100%" />
          {isGrid ? (
            <div className="hidden md:contents">
              <Skeleton height="14rem" width="100%" />
              <Skeleton height="14rem" width="100%" />
            </div>
          ) : null}
        </div>
      ) : components.length ? (
        <div
          className={cn(
            'flex flex-col gap-4',
            isGrid && 'md:grid md:grid-cols-2'
          )}
        >
          {components.map((component) => {
            const disabled = component.enabled === false

            return (
              <Card
                key={component.id}
                className={cn(
                  '!p-4 !gap-4',
                  disabled && 'opacity-55',
                  isGrid && 'md:h-full'
                )}
              >
                <div className="flex items-start justify-between gap-3 flex-wrap">
                  <div className="flex flex-col gap-1.5 min-w-0">
                    <div className="flex items-center gap-2 flex-wrap min-w-0">
                      {component.type ? (
                        <ComponentType
                          type={component.type}
                          displayVariant="icon-only"
                          iconSize="16"
                          colorVariant={disabled ? 'mono' : 'color'}
                        />
                      ) : null}
                      {component.href ? (
                        <Link href={component.href} className="min-w-0">
                          <Text
                            variant="body"
                            weight="stronger"
                            role="heading"
                            level={3}
                            theme={disabled ? 'neutral' : undefined}
                          >
                            {component.name}
                          </Text>
                        </Link>
                      ) : (
                        <Text
                          variant="body"
                          weight="stronger"
                          role="heading"
                          level={3}
                          theme={disabled ? 'neutral' : undefined}
                        >
                          {component.name}
                        </Text>
                      )}
                      {disabled ? (
                        <Badge size="sm" theme="neutral">
                          Disabled
                        </Badge>
                      ) : (
                        <Status status={component.status} variant="badge" />
                      )}
                      {component.behind ? <InstallConfigBehindBadge /> : null}
                    </div>
                    <ID>{component.id}</ID>
                  </div>
                  {component.actions}
                </div>

                <div className="flex flex-col gap-4 border-t pt-4">
                  {component.latestDeploy}
                </div>
              </Card>
            )
          })}
        </div>
      ) : filtered ? (
        <EmptyState
          variant="table"
          emptyTitle="No components found"
          emptyMessage="No components match the current search and filters."
        />
      ) : (
        <EmptyState
          variant="table"
          emptyTitle="No components configured"
          emptyMessage="Components appear here once they are defined on the app config and synced to this install."
        />
      )}

      {pagination && (pagination.hasNext || (pagination.offset ?? 0) !== 0) ? (
        <Pagination {...pagination} />
      ) : null}
    </div>
  )
}

export const InstallComponentsList = (props: IInstallComponentsList) => (
  <PaginationProvider>
    <InstallComponentsListBase {...props} />
  </PaginationProvider>
)
