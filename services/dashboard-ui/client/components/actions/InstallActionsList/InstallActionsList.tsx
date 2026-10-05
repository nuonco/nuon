import { useEffect, type ReactNode } from 'react'
import { Card } from '@/components/common/Card'
import {
  COLLECTION_VIEW_MODES,
  COLLECTION_VIEW_STORAGE_KEY,
  CollectionViewToggle,
  type TCollectionView,
} from '@/components/common/CollectionViewToggle'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Pagination, type IPagination } from '@/components/common/Pagination'
import { Text } from '@/components/common/Text'
import { LatestActionRunCard } from '@/components/actions/LatestActionRunCard'
import { RemovedFromAppConfigBadge } from '@/components/installs/RemovedFromAppConfig'
import { usePagination } from '@/hooks/use-pagination'
import { useStoredViewMode } from '@/hooks/use-stored-view-mode'
import { PaginationProvider } from '@/providers/pagination-provider'
import { cn } from '@/utils/classnames'

export type TInstallActionListItem = {
  actions?: ReactNode
  href?: string
  id: string
  latestRun: ReactNode
  loading?: boolean
  name: string
  removed?: boolean
}

const loadingItems = (count: number): TInstallActionListItem[] =>
  Array.from({ length: count }, (_, index) => ({
    id: `loading-${index}`,
    name: '',
    loading: true,
    latestRun: <LatestActionRunCard flush isLoading />,
  }))

export interface IInstallActionsList {
  actions?: ReactNode
  banner?: ReactNode
  filterActions?: ReactNode
  filtered?: boolean
  items: TInstallActionListItem[]
  loading?: boolean
  onViewChange?: (view: TCollectionView) => void
  pagination?: Omit<IPagination, 'position'>
  search?: ReactNode
  view?: TCollectionView
}

const InstallActionsListBase = ({
  actions,
  banner,
  filterActions,
  filtered = false,
  items,
  loading = false,
  onViewChange,
  pagination,
  search,
  view: viewProp,
}: IInstallActionsList) => {
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
  const isGrid = view === 'grid'
  const rows = loading && !items.length ? loadingItems(3) : items

  useEffect(() => {
    setIsPaginating(false)
  }, [items, setIsPaginating])

  return (
    <div className="flex flex-col gap-4 md:gap-6">
      {banner}
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

      {rows.length ? (
        <div
          className={cn(
            'flex flex-col gap-4',
            isGrid && 'md:grid md:grid-cols-2'
          )}
        >
          {rows.map((item) => (
            <Card
              key={item.id}
              className={cn(
                '!p-4 !gap-4',
                isGrid && 'md:h-full',
                item.removed && 'opacity-55'
              )}
            >
              <div className="flex items-start justify-between gap-3 flex-wrap">
                <div className="flex flex-col gap-1.5 min-w-0">
                  <div className="flex items-center gap-2 flex-wrap min-w-0">
                    <Icon
                      variant="TerminalWindowIcon"
                      size={16}
                      className="text-cool-grey-400 shrink-0"
                    />
                    <Text
                      variant="body"
                      weight="stronger"
                      role="heading"
                      level={3}
                      loading={item.loading}
                      loadingWidth={18}
                    >
                      {item.href ? (
                        <Link href={item.href} variant="inline">
                          {item.name}
                        </Link>
                      ) : (
                        item.name
                      )}
                    </Text>
                    {item.removed ? (
                      <RemovedFromAppConfigBadge kind="action" />
                    ) : null}
                  </div>
                  <ID loading={item.loading} loadingWidth={24}>
                    {item.id}
                  </ID>
                </div>
                {item.actions}
              </div>

              <div className="flex flex-col gap-4 border-t pt-4">
                <Text variant="subtext" weight="strong" theme="neutral">
                  Latest run
                </Text>
                {item.latestRun}
              </div>
            </Card>
          ))}
        </div>
      ) : filtered ? (
        <EmptyState
          variant="table"
          emptyTitle="No actions found"
          emptyMessage="No actions match the current search and filters."
        />
      ) : (
        <EmptyState
          variant="table"
          emptyTitle="No actions configured"
          emptyMessage="Actions appear here once they are defined on the app config and synced to this install."
        />
      )}

      {pagination && (pagination.hasNext || (pagination.offset ?? 0) !== 0) ? (
        <Pagination {...pagination} />
      ) : null}
    </div>
  )
}

export const InstallActionsList = (props: IInstallActionsList) => (
  <PaginationProvider>
    <InstallActionsListBase {...props} />
  </PaginationProvider>
)
