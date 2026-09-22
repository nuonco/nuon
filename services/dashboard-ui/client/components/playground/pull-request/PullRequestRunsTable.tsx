import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Duration } from '@/components/common/Duration'
import { Icon } from '@/components/common/Icon'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { BranchRunCommit } from '@/components/branches/BranchRunCommit'
import { previewModeDisplayLabel } from '@/components/branches/shared/preview-mode'
import { Panel } from '@/components/surfaces/Panel'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TAppBranchRunPreviewMode } from '@/types'
import type {
  TPlaygroundPullRequest,
  TPlaygroundPullRequestRun,
} from './fixtures'

const MODE_THEME: Record<TAppBranchRunPreviewMode, 'brand' | 'info' | 'neutral'> =
  {
    apply: 'brand',
    'plan-only': 'info',
    'build-only': 'neutral',
  }

const ImpactCell = ({ run }: { run: TPlaygroundPullRequestRun }) => {
  if (run.mode === 'build-only') {
    return (
      <Text variant="subtext" theme="neutral">
        {run.components_built} built
      </Text>
    )
  }

  if (run.status !== 'success') {
    return (
      <Text variant="subtext" theme="neutral">
        —
      </Text>
    )
  }

  return (
    <span className="flex items-center gap-1 whitespace-nowrap">
      <Badge size="xs" theme="success">
        +{run.added}
      </Badge>
      <Badge size="xs" theme="warn">
        ~{run.changed}
      </Badge>
      <Badge size="xs" theme="error">
        -{run.removed}
      </Badge>
    </span>
  )
}

// A factory, not a component: SurfacesProvider clones whatever it is handed to
// inject isVisible, so the Panel has to be the outermost element.
const runDetailPanel = (
  pullRequest: TPlaygroundPullRequest,
  run: TPlaygroundPullRequestRun
) => (
  <Panel heading={`Preview run #${run.run_number}`} size="half">
    <div className="flex flex-col gap-6">
      <div className="grid grid-cols-2 gap-4">
        <div className="flex flex-col gap-1">
          <Text variant="subtext" theme="neutral">
            Status
          </Text>
          <Status variant="default" status={run.status} />
        </div>
        <div className="flex flex-col gap-1">
          <Text variant="subtext" theme="neutral">
            Mode
          </Text>
          <Badge size="sm" theme={MODE_THEME[run.mode]}>
            {previewModeDisplayLabel(run.mode)}
          </Badge>
        </div>
        <div className="flex flex-col gap-1">
          <Text variant="subtext" theme="neutral">
            Install
          </Text>
          <Text variant="subtext" family={run.install_name ? 'mono' : 'sans'}>
            {run.install_name ?? 'Not used'}
          </Text>
        </div>
        <div className="flex flex-col gap-1">
          <Text variant="subtext" theme="neutral">
            Duration
          </Text>
          <Duration
            variant="subtext"
            beginTime={run.started_at}
            endTime={run.completed_at}
          />
        </div>
      </div>

      <div className="flex flex-col gap-2">
        <Text variant="subtext" theme="neutral">
          Commit
        </Text>
        <BranchRunCommit
          showStatus={false}
          sha={run.head_sha}
          message={run.commit_message}
          author={run.commit_author}
          createdAt={run.started_at}
          href={`${pullRequest.html_url}/commits/${run.head_sha}`}
          isExternal
        />
      </div>

      {run.error_message ? (
        <div className="flex flex-col gap-2">
          <Text variant="subtext" theme="neutral">
            Error
          </Text>
          <pre className="whitespace-pre-wrap rounded-md border border-red-300 bg-red-50 p-3 font-mono text-xs text-red-800 dark:border-red-500/40 dark:bg-red-950 dark:text-red-500">
            {run.error_message}
          </pre>
        </div>
      ) : null}
    </div>
  </Panel>
)

export interface IPullRequestRunsTable {
  pullRequest: TPlaygroundPullRequest
  runs: TPlaygroundPullRequestRun[]
  isLoading?: boolean
}

export const PullRequestRunsTable = ({
  pullRequest,
  runs,
  isLoading = false,
}: IPullRequestRunsTable) => {
  const { addPanel } = useSurfaces()

  const columns = useMemo<ColumnDef<TPlaygroundPullRequestRun>[]>(
    () => [
      {
        header: 'Run',
        accessorKey: 'run_number',
        cell: ({ row }) => (
          <Button
            variant="ghost"
            size="sm"
            className="!px-1.5"
            onClick={() => addPanel(runDetailPanel(pullRequest, row.original))}
          >
            <span className="flex items-center gap-1.5 whitespace-nowrap">
              <Status variant="default" status={row.original.status} />
              <Text variant="subtext" family="mono">
                #{row.original.run_number}
              </Text>
            </span>
          </Button>
        ),
      },
      {
        header: 'Commit',
        accessorKey: 'head_sha',
        cell: ({ row }) => (
          <BranchRunCommit
            showStatus={false}
            className="max-w-96"
            sha={row.original.head_sha}
            message={row.original.commit_message}
            author={row.original.commit_author}
            createdAt={row.original.started_at}
            href={`${pullRequest.html_url}/commits/${row.original.head_sha}`}
            isExternal
          />
        ),
      },
      {
        header: 'Mode',
        accessorKey: 'mode',
        cell: ({ row }) => (
          <Badge size="sm" theme={MODE_THEME[row.original.mode]}>
            {previewModeDisplayLabel(row.original.mode)}
          </Badge>
        ),
      },
      {
        header: 'Install',
        accessorKey: 'install_name',
        cell: ({ row }) =>
          row.original.install_name ? (
            <Text variant="subtext" family="mono">
              {row.original.install_name}
            </Text>
          ) : (
            <Text variant="subtext" theme="neutral">
              —
            </Text>
          ),
      },
      {
        header: 'Impact',
        id: 'impact',
        enableSorting: false,
        cell: ({ row }) => <ImpactCell run={row.original} />,
      },
      {
        header: 'Started',
        accessorKey: 'started_at',
        cell: ({ row }) => (
          <Time
            variant="subtext"
            theme="neutral"
            time={row.original.started_at}
            format="relative"
          />
        ),
      },
      {
        header: 'Duration',
        id: 'duration',
        enableSorting: false,
        cell: ({ row }) =>
          row.original.completed_at ? (
            <Duration
              variant="subtext"
              theme="neutral"
              beginTime={row.original.started_at}
              endTime={row.original.completed_at}
            />
          ) : (
            <Text variant="subtext" theme="neutral">
              in progress
            </Text>
          ),
      },
    ],
    [addPanel, pullRequest]
  )

  return (
    <Table<TPlaygroundPullRequestRun>
      columns={columns}
      data={runs}
      isLoading={isLoading}
      enableSearch={false}
      initialSorting={[{ id: 'run_number', desc: true }]}
      emptyStateProps={{
        emptyTitle: 'No preview runs yet',
        emptyMessage:
          'The first run fires when this pull request is opened or pushed to.',
        action: (
          <Text
            variant="subtext"
            theme="neutral"
            className="flex items-center gap-1"
          >
            <Icon variant="GitCommitIcon" size={12} theme="neutral" />
            Waiting on a commit
          </Text>
        ),
      }}
    />
  )
}
