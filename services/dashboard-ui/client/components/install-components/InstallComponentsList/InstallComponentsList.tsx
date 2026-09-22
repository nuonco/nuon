import type { ReactNode } from 'react'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { ID } from '@/components/common/ID'
import { Skeleton } from '@/components/common/Skeleton'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { ComponentType } from '@/components/components/ComponentType'
import type { TComponentType } from '@/types'

export type TInstallComponentListItem = {
  deployAction?: ReactNode
  health?: ReactNode
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
  search?: ReactNode
}

export const InstallComponentsList = ({
  actions,
  components,
  filterActions,
  filtered = false,
  loading = false,
  search,
}: IInstallComponentsList) => (
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
      <div className="flex flex-col divide-y">
        {components.map((component) => (
          <div
            key={component.id}
            className="flex flex-col gap-4 py-6 first:pt-0 last:pb-0"
          >
            <div className="flex items-start justify-between gap-3 flex-wrap">
              <div className="flex flex-col gap-1.5 min-w-0">
                <div className="flex items-center gap-2 flex-wrap min-w-0">
                  <Text variant="body" weight="stronger" role="heading" level={3}>
                    {component.name}
                  </Text>
                  {component.type ? (
                    <ComponentType
                      type={component.type}
                      variant="subtext"
                      colorVariant="color"
                    />
                  ) : null}
                  <Status status={component.status} variant="badge" />
                </div>
                <ID>{component.id}</ID>
              </div>
              {component.deployAction}
            </div>

            <div className="flex flex-col gap-2">
              <Text variant="subtext" weight="strong" theme="neutral">
                Latest deploy
              </Text>
              {component.latestDeploy}
            </div>

            {component.health ? (
              <div className="flex flex-col gap-2">
                <Text variant="subtext" weight="strong" theme="neutral">
                  Health
                </Text>
                <Card className="!p-4">{component.health}</Card>
              </div>
            ) : null}
          </div>
        ))}
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
  </div>
)
