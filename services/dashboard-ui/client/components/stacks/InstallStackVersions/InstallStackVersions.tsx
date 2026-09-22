import type { ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Status } from '@/components/common/Status'
import { Time } from '@/components/common/Time'
import { StackVersionDetails } from '@/components/stacks/StackVersionDetails'
import type { TInstallStack } from '@/types'

export type TInstallStackVersion = TInstallStack['versions'][number]

export interface IInstallStackVersions {
  latestAction?: ReactNode
  loading?: boolean
  versions: TInstallStackVersion[]
}

export const InstallStackVersions = ({
  latestAction,
  loading = false,
  versions,
}: IInstallStackVersions) => {
  if (!loading && !versions.length) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="No stack versions yet"
        emptyMessage="Versions appear here each time this install applies a new stack config."
      />
    )
  }

  if (loading && !versions.length) {
    return (
      <div className="flex flex-col gap-2">
        <Skeleton height="4rem" width="100%" />
        <Skeleton height="4rem" width="100%" />
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-2">
      {versions.map((version, index) => (
        <Card key={version.id} className="!p-4 !gap-3">
          <div className="flex items-center justify-between gap-3 flex-wrap">
            <div className="flex items-center gap-2 min-w-0">
              <Icon
                variant="StackIcon"
                size={14}
                className="text-cool-grey-400 shrink-0"
              />
              <Badge size="sm" variant="code" theme="neutral">
                {version.id}
              </Badge>
            </div>
            <Status status={version.composite_status?.status} variant="badge" />
          </div>
          <div className="flex items-center justify-between gap-3 flex-wrap">
            <Time
              time={version.created_at}
              format="relative"
              variant="subtext"
              theme="neutral"
            />
            <div className="flex items-center gap-2 flex-wrap justify-end">
              <StackVersionDetails
                version={version}
                panelKey={`stack-version-${version.id}`}
                triggerButton={{
                  variant: 'secondary',
                  size: 'sm',
                  children: 'View details',
                }}
              />
              {index === 0 ? latestAction : null}
            </div>
          </div>
        </Card>
      ))}
    </div>
  )
}
