import type { ReactNode } from 'react'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import type { TAppLabelKey } from '@/lib/ctl-api/apps/get-app-labels'
import { LabelRow } from './LabelRow'

interface IAppLabels {
  labels: TAppLabelKey[]
  isLoading?: boolean
  isPending?: boolean
  resetAction?: ReactNode
  onOverride: (key: string, color: string) => void
  onRemoveOverride: (key: string) => void
}

export const AppLabels = ({
  labels,
  isLoading,
  isPending,
  resetAction,
  onOverride,
  onRemoveOverride,
}: IAppLabels) => (
  <>
    <HeadingGroup className="gap-1.5">
      <div className="flex items-center justify-between gap-4">
        <Text variant="h3" weight="stronger" level={1}>Labels</Text>
        {resetAction}
      </div>
      <Text variant="subtext" theme="neutral">
        Every label key used across components, actions, runbooks, and installs. Each key gets a
        color automatically — override any you want to customize.
      </Text>
    </HeadingGroup>

    {isLoading ? (
      <Card>
        <div className="flex flex-col gap-4">
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} height="48px" width="100%" />
          ))}
        </div>
      </Card>
    ) : labels.length === 0 ? (
      <EmptyState
        variant="diagram"
        emptyTitle="No labels yet"
        emptyMessage="Add labels to your components, actions, runbooks, or installs to see them here."
      />
    ) : (
      <Card>
        <div className="flex flex-col divide-y">
          {labels.map((label) => (
            <LabelRow
              key={label.key}
              label={label}
              disabled={isPending}
              onOverride={onOverride}
              onRemoveOverride={onRemoveOverride}
            />
          ))}
        </div>
      </Card>
    )}
  </>
)
