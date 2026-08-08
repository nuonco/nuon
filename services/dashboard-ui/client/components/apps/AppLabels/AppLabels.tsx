import type { ReactNode } from 'react'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import type { TAppLabelKey } from '@/lib/ctl-api/apps/get-app-labels'
import { LabelRow } from './LabelRow'

interface IAppLabels {
  labels: TAppLabelKey[]
  defaultLabels?: Record<string, string>
  isLoading?: boolean
  isPending?: boolean
  resetAction?: ReactNode
  onOverride: (key: string, color: string) => void
  onRemoveOverride: (key: string) => void
}

export const AppLabels = ({
  labels,
  defaultLabels = {},
  isLoading,
  isPending,
  resetAction,
  onOverride,
  onRemoveOverride,
}: IAppLabels) => {
  const defaultEntries = Object.entries(defaultLabels).sort(([a], [b]) =>
    a.localeCompare(b),
  )

  return (
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

    {defaultEntries.length > 0 && (
      <>
        <HeadingGroup className="gap-1.5">
          <Text variant="base" weight="stronger" level={2}>Default labels</Text>
          <Text variant="subtext" theme="neutral">
            Applied to every install of this app. Edit them via default_labels in the app config
            and sync.
          </Text>
        </HeadingGroup>
        <Card>
          <div className="flex flex-col divide-y">
            {defaultEntries.map(([key, value]) => (
              <div
                key={key}
                className="flex items-center justify-between gap-4 py-4 first:pt-0 last:pb-0"
              >
                <LabelBadge
                  labelKey={key}
                  labelValue={value}
                  customColor={labels.find((l) => l.key === key)?.color}
                  size="sm"
                />
                <Tooltip
                  tipContent="Defined in the app config's default_labels"
                  position="left"
                  tipContentClassName="!whitespace-normal !w-auto max-w-[200px] text-xs"
                >
                  <span className="flex">
                    <Icon variant="LockIcon" size="16" />
                  </span>
                </Tooltip>
              </div>
            ))}
          </div>
        </Card>
      </>
    )}

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
}
