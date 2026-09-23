import type { ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { CodeBlock } from '@/components/diffs/CodeBlock'

export interface IInstallConfigFile {
  action?: ReactNode
  content?: string
  filename?: string
  history?: ReactNode
  isLoading?: boolean
  isManagedByConfig: boolean
  latestVersionId?: string
  syncedAt?: string
}

export const InstallConfigFile = ({
  action,
  content,
  filename,
  history,
  isLoading,
  isManagedByConfig,
  latestVersionId,
  syncedAt,
}: IInstallConfigFile) => {
  if (!isManagedByConfig) {
    return (
      <EmptyState
        variant="diagram"
        emptyTitle="No install config file"
        emptyMessage="This install is managed from the dashboard."
      />
    )
  }

  return (
    <div className="flex flex-col gap-4">
      <Card className="!p-4 !gap-4">
        <div className="flex items-center justify-between gap-3 flex-wrap">
          <span className="flex items-center gap-2 min-w-0">
            <Icon
              variant="FileCodeIcon"
              size={14}
              className="text-cool-grey-400 shrink-0"
            />
            <Text variant="body" weight="strong">
              Current config file
            </Text>
          </span>
          <span className="flex items-center gap-3">
            {syncedAt ? (
              <Text variant="subtext" theme="neutral">
                Synced <Time as="span" time={syncedAt} format="relative" />
              </Text>
            ) : null}
            {latestVersionId ? (
              <Badge size="sm" variant="code" theme="neutral">
                {latestVersionId}
              </Badge>
            ) : null}
          </span>
        </div>
        {isLoading ? (
          <Skeleton height="240px" width="100%" />
        ) : content ? (
          <CodeBlock
            value={content}
            language="toml"
            filename={filename ?? 'install.toml'}
            copy
          />
        ) : (
          <Text variant="subtext" theme="neutral">
            The current install config could not be generated.
          </Text>
        )}
      </Card>

      <div className="flex items-center justify-between gap-3">
        <Text variant="body" weight="strong">
          Config history
        </Text>
        {action}
      </div>
      {history}
    </div>
  )
}
