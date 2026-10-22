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
import { Expand } from '@/components/common/Expand'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
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

export type TInstallComponentListItem = {
  actions?: ReactNode
  alignment?: TInstallAlignmentCategory
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
  aligned?: boolean
  components: TInstallComponentListItem[]
  drifted?: (view: TCollectionView) => ReactNode
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
  aligned = false,
  components,
  drifted,
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
  const tornDown = components.filter((item) => item.alignment === 'torn-down')
  const remaining = components.filter((item) => item.alignment !== 'torn-down')

  const renderCards = (items: TInstallComponentListItem[]) => (
    <div
      className={cn('flex flex-col gap-4', isGrid && 'md:grid md:grid-cols-2')}
    >
      {items.map((component) => {
        const disabled = component.enabled === false

        return (
          <Card
            key={component.id}
            className={cn(
              disabled ? '!p-3 !gap-3 self-start w-full' : '!p-4 !gap-4',
              isGrid && !disabled && 'md:h-full'
            )}
          >
            <div
              className={cn(
                'flex justify-between gap-3 flex-wrap',
                disabled ? 'items-center' : 'items-start'
              )}
            >
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
                </div>
                {!disabled ? <ID>{component.id}</ID> : null}
              </div>
              {component.actions}
            </div>

            {!disabled ? (
              <div className="flex flex-col gap-4 border-t pt-4">
                {component.latestDeploy}
              </div>
            ) : null}
          </Card>
        )
      })}
    </div>
  )

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

      {drifted?.(view)}

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
      ) : remaining.length ? (
        aligned ? (
          <AlignedSections
            items={remaining}
            noun="components"
            render={renderCards}
          />
        ) : (
          renderCards(remaining)
        )
      ) : tornDown.length ? null : aligned ? (
        <EmptyState
          variant="table"
          emptyTitle="No other components"
          emptyMessage="Check Ongoing deployments and Drifted above, or adjust the filters."
        />
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

      {tornDown.length ? (
        <section
          aria-label="Torn down components"
          data-testid="torn-down-section"
        >
          <Expand
            id="torn-down-components"
            className="border-t"
            heading={`Torn down components · ${tornDown.length}`}
          >
            <ul className="divide-y max-h-64 overflow-y-auto">
              {tornDown.map((component) => (
                <li
                  key={component.id}
                  className="flex items-center gap-3 px-2 py-2"
                >
                  {component.type ? (
                    <ComponentType
                      type={component.type}
                      displayVariant="icon-only"
                      iconSize="16"
                      colorVariant="mono"
                    />
                  ) : null}
                  {component.href ? (
                    <Link href={component.href} className="!font-mono">
                      {component.name}
                    </Link>
                  ) : (
                    <Text family="mono" variant="subtext">
                      {component.name}
                    </Text>
                  )}
                </li>
              ))}
            </ul>
          </Expand>
        </section>
      ) : null}

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
