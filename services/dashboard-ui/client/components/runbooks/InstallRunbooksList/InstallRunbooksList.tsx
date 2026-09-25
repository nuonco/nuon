import { useEffect, type ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Pagination, type IPagination } from '@/components/common/Pagination'
import { Text } from '@/components/common/Text'
import { LatestRunbookRunCard } from '@/components/runbooks/LatestRunbookRunCard'
import { RemovedFromAppConfigBadge } from '@/components/installs/RemovedFromAppConfig'
import { usePagination } from '@/hooks/use-pagination'
import { PaginationProvider } from '@/providers/pagination-provider'

export type TInstallRunbookListItem = {
  actions?: ReactNode
  description?: string
  href?: string
  id: string
  latestRun: ReactNode
  loading?: boolean
  name: string
  removed?: boolean
  stepCount?: number
}

const loadingItems = (count: number): TInstallRunbookListItem[] =>
  Array.from({ length: count }, (_, index) => ({
    id: `loading-${index}`,
    name: '',
    loading: true,
    latestRun: <LatestRunbookRunCard flush isLoading />,
  }))

export interface IInstallRunbooksList {
  actions?: ReactNode
  filterActions?: ReactNode
  filtered?: boolean
  items: TInstallRunbookListItem[]
  loading?: boolean
  pagination?: Omit<IPagination, 'position'>
  search?: ReactNode
}

const InstallRunbooksListBase = ({
  actions,
  filterActions,
  filtered = false,
  items,
  loading = false,
  pagination,
  search,
}: IInstallRunbooksList) => {
  const { setIsPaginating } = usePagination()
  const rows = loading && !items.length ? loadingItems(3) : items

  useEffect(() => {
    setIsPaginating(false)
  }, [items, setIsPaginating])

  return (
    <div className="flex flex-col gap-4 md:gap-6">
      {search || filterActions || actions ? (
        <div className="flex flex-row flex-wrap items-center justify-between gap-4">
          <div className="flex flex-wrap items-center gap-4 w-full md:w-fit">
            {search}
            {filterActions}
          </div>
          {actions}
        </div>
      ) : null}

      {rows.length ? (
        <div className="flex flex-col gap-4">
          {rows.map((item) => (
            <Card
              key={item.id}
              className={`!p-4 !gap-4 ${item.removed ? 'opacity-55' : ''}`}
            >
              <div className="flex items-start justify-between gap-3 flex-wrap">
                <div className="flex flex-col gap-1.5 min-w-0">
                  <div className="flex items-center gap-2 flex-wrap min-w-0">
                    <Icon
                      variant="BookIcon"
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
                    {typeof item.stepCount === 'number' ? (
                      <Badge size="sm" theme="neutral">
                        {item.stepCount}{' '}
                        {item.stepCount === 1 ? 'step' : 'steps'}
                      </Badge>
                    ) : null}
                    {item.removed ? (
                      <RemovedFromAppConfigBadge kind="runbook" />
                    ) : null}
                  </div>
                  <ID loading={item.loading} loadingWidth={24}>
                    {item.id}
                  </ID>
                  {item.loading ? (
                    <Text variant="subtext" loading loadingWidth={40} />
                  ) : item.description ? (
                    <Text variant="subtext" theme="neutral">
                      {item.description}
                    </Text>
                  ) : null}
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
          emptyTitle="No runbooks found"
          emptyMessage="No runbooks match the current search and filters."
        />
      ) : (
        <EmptyState
          variant="table"
          emptyTitle="No runbooks configured"
          emptyMessage="Runbooks appear here once they are defined on the app config and synced to this install."
        />
      )}

      {pagination && (pagination.hasNext || (pagination.offset ?? 0) !== 0) ? (
        <Pagination {...pagination} />
      ) : null}
    </div>
  )
}

export const InstallRunbooksList = (props: IInstallRunbooksList) => (
  <PaginationProvider>
    <InstallRunbooksListBase {...props} />
  </PaginationProvider>
)
