import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState'
import { Expand } from '@/components/common/Expand'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Card } from '@/components/common/Card'
import { ChangeCountSummary } from '@/components/approvals/plan-diffs/ChangeCountSummary'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { Toast } from '@/components/surfaces/Toast'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import {
  getInstallConfigVersions,
  getInstallConfigVersionDiff,
  syncInstallConfig,
} from '@/lib'
import type {
  TInstallConfigVersion,
  TConfigDiffNode,
  TConfigDiffKey,
} from '@/types'

export const InstallConfigs = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const { data: versions, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-config-versions', org?.id, install?.id],
    queryFn: () =>
      getInstallConfigVersions({
        orgId: org!.id,
        installId: install!.id,
      }),
    enabled: !!org?.id && !!install?.id,
  })

  const { mutate: syncNow, isPending: isSyncing } = useMutation({
    mutationFn: () =>
      syncInstallConfig({
        installId: install!.id,
        orgId: org!.id,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['install-config-versions', org?.id, install?.id],
      })
      addToast(
        <Toast heading="Syncing config" theme="info">
          <Text>Syncing configs for {install?.name}.</Text>
        </Toast>
      )
    },
    onError: (err) => {
      addToast(
        <Toast heading="Config sync failed" theme="error">
          <Text>{err?.error || 'Unable to trigger config sync.'}</Text>
        </Toast>
      )
    },
  })

  const canSync = !!install?.id

  return (
    <PageSection>
      <PageTitle title={`Configs | ${install?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/configs`,
            text: 'Configs',
          },
        ]}
      />

      <div className="flex items-start justify-between gap-4">
        <HeadingGroup>
          <Text variant="base" weight="strong">
            Config history
          </Text>
          <Text variant="subtext" theme="neutral">
            Install config versions synced from git.
          </Text>
        </HeadingGroup>
        {canSync && (
          <div className="shrink-0">
            <Button
              variant="secondary"
              disabled={isSyncing}
              onClick={() => syncNow()}
            >
              {isSyncing ? 'Syncing...' : 'Sync now'}
            </Button>
          </div>
        )}
      </div>

      {isLoading && (
        <Card>
          <Text variant="subtext" theme="neutral">
            Loading...
          </Text>
        </Card>
      )}

      {!isLoading && (!versions || versions.length === 0) && (
        <Card>
          <EmptyState
            emptyTitle="No configs yet"
            emptyMessage="Config versions will appear here after the first sync from git."
            variant="table"
          />
        </Card>
      )}

      {versions && versions.length > 0 && (
        <div className="flex flex-col gap-3">
          {versions.map((version) => (
            <ConfigVersionRow key={version.id} version={version} />
          ))}
        </div>
      )}
    </PageSection>
  )
}

const ConfigVersionRow = ({ version }: { version: TInstallConfigVersion }) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const sync = version.install_config_sync
  const triggeredBy = sync?.triggered_by ?? 'unknown'
  const commit = sync?.vcs_connection_commit

  const { data: diff, isLoading: isDiffLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: [
      'install-config-version-diff',
      org?.id,
      install?.id,
      version.id,
    ],
    queryFn: () =>
      getInstallConfigVersionDiff({
        orgId: org!.id,
        installId: install!.id,
        versionId: version.id,
      }),
    enabled: !!org?.id && !!install?.id && !!version.id,
    retry: 1,
  })

  const diffSummary = diff?.key ? getDiffSummary(diff) : null

  return (
    <Expand
      id={`config-version-${version.id}`}
      className="border rounded-xl bg-white dark:bg-dark-grey-900 shadow-sm overflow-hidden"
      headerClassName="px-5 py-4"
      heading={
        <div className="flex items-center gap-3 w-full">
          <Status
            status={version.status?.status ?? 'unknown'}
            variant="badge"
          />
          <Badge size="sm" theme="neutral">
            {triggeredBy}
          </Badge>
          <Badge size="sm" theme={version.created ? 'success' : 'info'}>
            {version.created ? 'Created' : 'Updated'}
          </Badge>
          {version.file_path && (
            <Text
              variant="subtext"
              theme="neutral"
              className="font-mono text-xs"
            >
              {version.file_path}
            </Text>
          )}
          {version.created_at && (
            <Time
              variant="subtext"
              theme="neutral"
              time={version.created_at}
              format="relative"
              className="ml-auto"
            />
          )}
          {diffSummary && (
            <ChangeCountSummary
              added={diffSummary.added}
              updated={diffSummary.changed}
              removed={diffSummary.removed}
            />
          )}
        </div>
      }
    >
      <div className="p-5 border-t flex flex-col gap-4">
        {commit && <CommitCard commit={commit} />}

        {isDiffLoading && !diff ? (
          <Skeleton lines={3} height="1rem" />
        ) : diff ? (
          <DiffTree node={diff} />
        ) : (
          <Text variant="subtext" theme="neutral">
            No diff available
          </Text>
        )}
      </div>
    </Expand>
  )
}

const CommitCard = ({
  commit,
}: {
  commit: NonNullable<
    NonNullable<TInstallConfigVersion['install_config_sync']>['vcs_connection_commit']
  >
}) => {
  return (
    <div className="flex items-start gap-3">
      <Icon
        variant="GitCommitIcon"
        size={16}
        className="mt-0.5 shrink-0 text-cool-grey-500 dark:text-cool-grey-400"
      />
      <div className="flex flex-col gap-1 min-w-0">
        {commit.message && (
          <Text variant="body" weight="strong" className="truncate">
            {commit.message.split('\n')[0]?.trim()}
          </Text>
        )}
        <div className="flex items-center gap-2 flex-wrap">
          {commit.sha && (
            <Badge size="sm" variant="code">
              {commit.sha.slice(0, 8)}
            </Badge>
          )}
          {commit.author_name && (
            <Text variant="subtext" theme="neutral">
              {commit.author_name}
            </Text>
          )}
        </div>
      </div>
    </div>
  )
}

const OP_THEME = {
  add: 'success',
  remove: 'error',
  change: 'warn',
  noop: 'neutral',
  '': 'neutral',
} as const

const OP_PREFIX = {
  add: '+',
  remove: '-',
  change: '~',
  noop: ' ',
  '': ' ',
} as const

const DiffTree = ({ node }: { node: TConfigDiffNode }) => {
  const changes = collectChangedLeaves(node)

  if (changes.length === 0) {
    return (
      <Text variant="subtext" theme="neutral">
        No changes detected
      </Text>
    )
  }

  return (
    <div className="flex flex-col border rounded-md divide-y overflow-hidden">
      {changes.map((change) => (
        <DiffRow key={change.path} change={change} />
      ))}
    </div>
  )
}

interface IDiffChange {
  path: string
  op: TConfigDiffKey['op']
  diff: string
}

const DiffRow = ({ change }: { change: IDiffChange }) => {
  const op = change.op || ''
  return (
    <div className="flex items-center gap-3 px-4 py-2.5">
      <Text
        variant="subtext"
        theme={OP_THEME[op]}
        weight="strong"
        family="mono"
        className="shrink-0 w-4 text-center"
      >
        {OP_PREFIX[op]}
      </Text>
      <Text variant="subtext" weight="strong" className="font-mono">
        {change.path}
      </Text>
      <Text variant="subtext" theme="neutral" className="ml-auto font-mono text-xs truncate max-w-[50%]">
        {change.diff}
      </Text>
    </div>
  )
}

function collectChangedLeaves(
  node: TConfigDiffNode | null | undefined,
  prefix = ''
): IDiffChange[] {
  if (!node) return []

  const path = prefix ? `${prefix}.${node.key}` : node.key

  if (node.diff) {
    if (node.diff.op === 'noop' || node.diff.op === '') return []
    return [{ path, op: node.diff.op, diff: node.diff.diff }]
  }

  if (node.children) {
    return node.children
      .filter(Boolean)
      .flatMap((child) => collectChangedLeaves(child, path))
  }

  return []
}

function getDiffSummary(node: TConfigDiffNode): {
  added: number
  changed: number
  removed: number
} {
  const leaves = collectChangedLeaves(node)
  return {
    added: leaves.filter((l) => l.op === 'add').length,
    changed: leaves.filter((l) => l.op === 'change').length,
    removed: leaves.filter((l) => l.op === 'remove').length,
  }
}
