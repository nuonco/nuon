import { useQuery } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { CommitLink } from '@/components/common/GitReferenceLink'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { getInstallWorkflows, getWorkflowSteps } from '@/lib'
import { Time } from '@/components/common/Time'
import type { TAppBranchRun, TWorkflow } from '@/types'
import type { TTrackInstall, TTrackStatus } from './RolloutTrack'

type TRunSummary = Pick<
  TAppBranchRun,
  'id' | 'head_sha' | 'vcs_connection_commit' | 'created_at'
>

export const previousBranchRun = (
  workflows: TWorkflow[] | undefined,
  currentRunId?: string
): TRunSummary | undefined =>
  (workflows ?? [])
    .flatMap((workflow) => workflow.app_branch_runs ?? [])
    .find((run) => !!run.id && run.id !== currentRunId)

const RunLine = ({
  label,
  run,
  repo,
  emptyText,
}: {
  label: string
  run?: TRunSummary
  repo?: string
  emptyText?: string
}) => {
  const sha = run?.vcs_connection_commit?.sha ?? run?.head_sha
  const message = run?.vcs_connection_commit?.message?.split('\n')[0]
  const createdAt = run?.vcs_connection_commit?.created_at ?? run?.created_at
  return (
    <div className="flex min-w-0 items-start gap-3">
      <Text variant="label" theme="neutral" className="w-24 shrink-0 pt-0.5">
        {label}
      </Text>
      {run ? (
        <div className="flex min-w-0 flex-col gap-0.5">
          <Text variant="subtext" className="truncate">
            {message || 'No commit message'}
          </Text>
          <span className="flex items-center gap-2">
            {sha ? (
              <CommitLink sha={sha} repo={repo} textVariant="label" />
            ) : null}
            {createdAt ? (
              <Time
                time={createdAt}
                format="relative"
                variant="label"
                theme="neutral"
              />
            ) : null}
          </span>
        </div>
      ) : (
        <Text variant="subtext" theme="neutral">
          {emptyText}
        </Text>
      )}
    </div>
  )
}

const StatusLine = ({
  label,
  value,
}: {
  label: string
  value?: TTrackStatus
}) =>
  value ? (
    <span className="flex min-w-0 items-center gap-3">
      <Text variant="label" theme="neutral" className="w-24 shrink-0">
        {label}
      </Text>
      <Status status={value.status || 'unknown'} />
      {value.detail ? (
        <Text variant="subtext" theme="neutral" className="truncate">
          {value.detail}
        </Text>
      ) : null}
    </span>
  ) : null

export const InstallRolloutPanel = ({
  install,
  approval,
  orgId,
  repo,
  branchRun,
  ...props
}: IPanel & {
  install: TTrackInstall
  approval?: string
  orgId?: string
  repo?: string
  branchRun?: TAppBranchRun
}) => {
  const { data: steps, isLoading } = useQuery({
    queryKey: ['workflow-steps', orgId, install.workflowId],
    queryFn: () =>
      getWorkflowSteps({ orgId: orgId!, workflowId: install.workflowId! }),
    enabled: !!orgId && !!install.workflowId,
  })
  const { data: workflows, isLoading: isLoadingHistory } = useQuery({
    queryKey: ['install-rollout-workflows', orgId, install.id],
    queryFn: () =>
      getInstallWorkflows({
        orgId: orgId!,
        installId: install.id,
        limit: 20,
        planonly: false,
      }),
    enabled: !!orgId && !!install.id,
  })
  const visibleSteps = (steps ?? []).filter(
    (step) => step.execution_type !== 'hidden' && step.name !== 'internal'
  )
  const previousRun = previousBranchRun(workflows?.data, branchRun?.id)

  return (
    <Panel {...props} size="half" heading={install.name}>
      <div className="flex flex-col gap-4">
        <div className="flex flex-col gap-3 rounded-xl border bg-white p-4 shadow-sm dark:bg-dark-grey-900">
          <span className="flex items-start justify-between gap-3">
            <Text variant="base" weight="strong" family="mono">
              {install.name}
            </Text>
            <Status status={install.status} />
          </span>
          {approval ? (
            <Text variant="subtext" theme="neutral">
              {approval}
            </Text>
          ) : null}
          <div className="flex flex-col gap-2 border-t pt-3">
            <RunLine label="Rolling out" run={branchRun} repo={repo} />
            <RunLine
              label="Previous run"
              run={previousRun}
              repo={repo}
              emptyText={
                isLoadingHistory ? 'Loading…' : 'No earlier branch run'
              }
            />
          </div>
          {install.overviewHref ? (
            <Button href={install.overviewHref} variant="secondary" size="sm">
              View install
            </Button>
          ) : null}
        </div>
        <StatusLine label="Resources" value={install.resources} />
        <StatusLine label="Deployment" value={install.deployment} />
        <StatusLine label="Health" value={install.health} />
        {install.workflowId ? (
          <div className="flex flex-col gap-2">
            <span className="flex items-center justify-between gap-3">
              <Text variant="body" weight="strong">
                Workflow
              </Text>
              {install.workflowHref ? (
                <Link href={install.workflowHref}>View workflow</Link>
              ) : null}
            </span>
            {isLoading ? (
              <Text variant="subtext" theme="neutral">
                Loading…
              </Text>
            ) : visibleSteps.length === 0 ? (
              <Text variant="subtext" theme="neutral">
                This workflow has no steps yet.
              </Text>
            ) : (
              <ul className="flex flex-col gap-2">
                {visibleSteps.map((step) => (
                  <li
                    key={step.id}
                    className="flex items-center gap-3 rounded-xl border bg-white p-3 shadow-sm dark:bg-dark-grey-900"
                  >
                    <Status status={step.status?.status || 'pending'} />
                    <Text variant="subtext">{step.name}</Text>
                  </li>
                ))}
              </ul>
            )}
          </div>
        ) : null}
      </div>
    </Panel>
  )
}
