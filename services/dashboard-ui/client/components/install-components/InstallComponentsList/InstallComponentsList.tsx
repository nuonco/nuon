import { useEffect, type ReactNode } from 'react'
import { useLocation } from 'react-router'
import { Badge } from '@/components/common/Badge'
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
import { scrollElementIntoView } from '@/utils/scroll'

export type TInstallComponentListItem = {
  actions?: ReactNode
  enabled?: boolean | null
  id: string
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
  pagination?: Omit<IPagination, 'position'>
  search?: ReactNode
}

const InstallComponentsListBase = ({
  actions,
  components,
  filterActions,
  filtered = false,
  loading = false,
  pagination,
  search,
}: IInstallComponentsList) => {
  const { setIsPaginating } = usePagination()
  const { hash } = useLocation()

  useEffect(() => {
    setIsPaginating(false)
  }, [components, setIsPaginating])

  useEffect(() => {
    const id = hash.replace('#', '')
    if (!id) return
    scrollElementIntoView(document.getElementById(id), { block: 'start' })
  }, [components, hash])

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

      {loading && !components.length ? (
        <div className="flex flex-col gap-4">
          <Skeleton height="14rem" width="100%" />
          <Skeleton height="14rem" width="100%" />
        </div>
      ) : components.length ? (
        <div className="flex flex-col gap-4">
          {components.map((component) => {
            const disabled = component.enabled === false

            return (
              <Card
                key={component.id}
                id={component.id}
                className={`!p-4 !gap-4 scroll-mt-4 ${disabled ? 'opacity-55' : ''}`}
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
                      <Text
                        variant="body"
                        weight="stronger"
                        role="heading"
                        level={3}
                        theme={disabled ? 'neutral' : undefined}
                      >
                        {component.name}
                      </Text>
                      {disabled ? (
                        <Badge size="sm" theme="neutral">
                          Disabled
                        </Badge>
                      ) : (
                        <Status status={component.status} variant="badge" />
                      )}
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
