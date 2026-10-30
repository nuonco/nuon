import type { CSSProperties, ReactNode } from 'react'
import { Badge } from '@/components/common/Badge'
import '@/components/common/LabelBadge.css'
import { Button } from '@/components/common/Button'
import { CommitLink } from '@/components/common/GitReferenceLink'
import { CompositeError } from '@/components/common/CompositeError'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { DetailPage } from '@/components/layout/DetailPage'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { OverviewLoadingTrack } from '@/components/branches/BranchOverview/OverviewLoadingTrack'
import {
  buildComponentsStep,
  buildOverviewLoadingStages,
  preRolloutCompositeError,
  type TPreviewProgress,
} from '@/components/branches/BranchOverview/overview-loading'
import {
  getStepDisplayStatus,
  stepStatusCategory,
} from '@/components/branches/shared/step-status'
import { previewModeDisplayLabel } from '@/components/branches/shared/preview-mode'
import { previewConfigStatus } from '@/components/pull-requests/map-pull-request'
import {
  PreviewRunProgress,
  type TPreviewRunBuild,
} from '@/components/pull-requests/PreviewRunProgress'
import type {
  TAppBranchPreviewConfig,
  TAppBranchRun,
  TAppBranchRunPreviewMode,
  TCompositeError,
  TInstallWorkflow,
  TInstallWorkflowStep,
  TPreviewSourcePR,
} from '@/types'

const githubLabelColor = (color?: string) =>
  color && /^[0-9a-fA-F]{6}$/.test(color) ? `#${color}` : undefined

const sameSha = (left?: string, right?: string) => {
  if (!left || !right) return true
  const a = left.toLowerCase()
  const b = right.toLowerCase()
  return a === b || a.startsWith(b) || b.startsWith(a)
}

const runMode = (run?: TAppBranchRun): TAppBranchRunPreviewMode | undefined =>
  run?.preview?.mode ?? run?.preview?.resolved_preview_config?.mode

const runInstall = (run?: TAppBranchRun) =>
  run?.preview?.install_name ||
  run?.preview?.resolved_preview_config?.install_name

const runSha = (run?: TAppBranchRun) =>
  run?.vcs_connection_commit?.sha || run?.head_sha || run?.metadata?.head_sha || ''

const trackMode = (mode?: TAppBranchRunPreviewMode): TPreviewProgress =>
  mode === 'apply' ? 'apply' : mode === 'plan-only' ? 'plan-only' : 'build-only'

// Real preview-run failures (unlike rollout failures) rarely populate a step's
// composite_error; fall back to the human description so the failure is still
// visible.
const previewRunFailure = (
  steps: TInstallWorkflowStep[],
  runError?: TCompositeError
): TCompositeError | undefined => {
  const fromStages = preRolloutCompositeError(steps, runError)
  if (fromStages) return fromStages
  const failed = steps.find(
    (step) => stepStatusCategory(getStepDisplayStatus(step)) === 'error'
  )
  if (!failed) return undefined
  return (
    failed.status?.composite_error ??
    (failed.status?.status_human_description
      ? { message: failed.status.status_human_description, severity: 'error' }
      : undefined)
  )
}

const runContextSentence = ({
  mode,
  install,
  needsInput,
  oneOff,
}: {
  mode?: TAppBranchRunPreviewMode
  install?: string
  needsInput: boolean
  oneOff: boolean
}) => {
  const label = previewModeDisplayLabel(mode ?? 'plan-only')
  if (needsInput) return `${label} — no install selected yet.`
  const base =
    mode === 'build-only' || mode === 'none'
      ? label
      : install
        ? `${label} against ${install}`
        : `${label}, no install selected`
  return oneOff
    ? `${base}. One-off for this run — later pushes use the branch default.`
    : `${base}. This is the branch default.`
}

export interface IPullRequestPreview {
  source?: TPreviewSourcePR
  workflows?: TInstallWorkflow[]
  repo?: string
  baseRef?: string
  branchConfig?: TAppBranchPreviewConfig
  basePath: string
  orgId?: string
  appId?: string
  backLink?: ReactNode
  changes?: ReactNode
  isLoading?: boolean
  onEditConfig?: () => void
  onRerun?: () => void
}

export const PullRequestPreview = ({
  source,
  workflows = [],
  repo,
  baseRef,
  branchConfig,
  basePath,
  orgId,
  appId,
  backLink,
  changes,
  isLoading = false,
  onEditConfig,
  onRerun,
}: IPullRequestPreview) => {
  const latest = workflows[0]
  const latestRun = latest?.app_branch_runs?.[0]
  const steps = latest?.steps ?? []
  const mode = runMode(latestRun)
  const install = runInstall(latestRun)
  const sha = runSha(latestRun)
  const stale = !!source?.head_sha && !!sha && !sameSha(source.head_sha, sha)
  const needsInput = previewConfigStatus(branchConfig, latestRun) === 'needs-input'
  const oneOff = !!(
    latestRun?.preview?.override_preview_config?.mode ||
    latestRun?.preview?.override_preview_config?.install_id
  )
  const stages = latest
    ? buildOverviewLoadingStages({ steps, sha, previewMode: trackMode(mode) })
    : []
  const failure = latest
    ? previewRunFailure(steps, latestRun?.composite_error)
    : undefined
  const running = !!latest && !latest.finished && !failure
  const diffReady =
    stages.find((stage) => stage.id === 'build-components')?.status ===
    'success'
  const buildMetadata = buildComponentsStep(steps)?.status?.metadata as
    | { builds?: TPreviewRunBuild[]; sandbox_build_id?: string }
    | undefined

  return (
    <DetailPage
      className="flex-1"
      header={
        <>
          {backLink}
          <DetailHeader
            backLink={false}
            loading={isLoading && !source}
            title={
              source
                ? `#${source.pr_number} ${source.title || ''}`.trim()
                : 'Pull request'
            }
            status={
              source ? (
                <span className="flex flex-wrap items-center gap-2">
                  <Badge size="sm" theme="success">
                    open
                  </Badge>
                  {source.is_draft ? (
                    <Badge size="sm">Draft</Badge>
                  ) : null}
                  {(source.labels ?? []).map((label) => {
                    const color = githubLabelColor(label.color)
                    return (
                      <Badge
                        key={label.name}
                        size="sm"
                        theme={color ? 'none' : 'default'}
                        className={color ? 'label-badge-value' : undefined}
                        style={
                          color
                            ? ({ '--label-color': color } as CSSProperties)
                            : undefined
                        }
                      >
                        {label.name}
                      </Badge>
                    )
                  })}
                </span>
              ) : null
            }
            description={
              source ? (
                <span className="flex flex-wrap items-center gap-x-2 gap-y-1">
                  {source.url ? (
                    <Link href={source.url} isExternal variant="inline">
                      <Icon variant="GitHub" size={12} />
                      {repo || 'GitHub'}
                    </Link>
                  ) : null}
                  {source.head_ref ? (
                    <Text variant="subtext" theme="neutral" family="mono" flex>
                      <Icon variant="GitBranchIcon" />
                      {source.head_ref}
                      {baseRef ? ` → ${baseRef}` : ''}
                    </Text>
                  ) : null}
                  {source.head_sha ? (
                    <CommitLink sha={source.head_sha} repo={repo} />
                  ) : null}
                  {source.author ? (
                    <Text variant="subtext" theme="neutral">
                      {source.author}
                    </Text>
                  ) : null}
                  {source.created_at ? (
                    <span className="inline-flex items-center gap-1">
                      <Text variant="subtext" theme="neutral">
                        opened
                      </Text>
                      <Time
                        time={source.created_at}
                        format="relative"
                        variant="subtext"
                        theme="neutral"
                      />
                    </span>
                  ) : null}
                </span>
              ) : null
            }
            actions={
              source ? (
                needsInput ? (
                  <Button variant="primary" onClick={onEditConfig}>
                    Choose an install
                  </Button>
                ) : (
                  <>
                    <Button variant="primary" onClick={onRerun}>
                      Re-run preview
                    </Button>
                    <Button onClick={onEditConfig}>
                      Run with different settings
                    </Button>
                  </>
                )
              ) : null
            }
          />
        </>
      }
    >
      <div className="flex flex-1 flex-col gap-4">
        {isLoading || latest ? (
          <div className="flex flex-col gap-3">
            <div className="flex flex-wrap items-center justify-between gap-3">
              {isLoading ? (
                <Text
                  variant="subtext"
                  theme="neutral"
                  loading
                  loadingWidth={220}
                >
                  Preview progress
                </Text>
              ) : (
                <OverviewLoadingTrack stages={stages} label="Preview progress" />
              )}
              {latest?.id ? (
                <Link href={`${basePath}/runs/${latest.id}`}>View run</Link>
              ) : null}
            </div>

            {failure ? <CompositeError error={failure} /> : null}

            {stale ? (
              <Text variant="subtext" theme="neutral">
                This run is for an older commit than the pull request.
              </Text>
            ) : null}

            <Text variant="subtext" theme="neutral" loading={isLoading}>
              {runContextSentence({ mode, install, needsInput, oneOff })}
            </Text>

            {running && diffReady ? (
              <PreviewRunProgress
                variant="inline"
                stages={stages}
                description={latest?.status?.status_human_description}
                sha={sha}
                baseRef={baseRef}
                install={install}
              />
            ) : null}
          </div>
        ) : (
          <div className="flex flex-1 flex-col items-center justify-center">
            <EmptyState
              variant="history"
              emptyTitle="No preview runs yet"
              emptyMessage="A run starts when this pull request is opened or pushed to."
              action={
                needsInput ? (
                  <Button variant="secondary" onClick={onEditConfig}>
                    Choose an install
                  </Button>
                ) : undefined
              }
            />
          </div>
        )}

        {running && !diffReady ? (
          <div className="flex flex-1 flex-col items-center justify-center py-12">
            <PreviewRunProgress
              stages={stages}
              description={latest?.status?.status_human_description}
              builds={buildMetadata?.builds}
              sandboxBuildId={buildMetadata?.sandbox_build_id}
              sha={sha}
              baseRef={baseRef}
              install={install}
              orgId={orgId}
              appId={appId}
            />
          </div>
        ) : null}

        {!running || diffReady ? changes : null}

        {workflows.length > 1 ? (
          <div className="flex flex-col gap-2">
            <SectionHeader
              title="Earlier runs"
              description="Newest first. The latest run is above."
            />
            <div className="flex flex-col divide-y border rounded-md">
              {workflows.slice(1).map((workflow) => {
                const run = workflow.app_branch_runs?.[0]
                const href = workflow.id
                  ? `${basePath}/runs/${workflow.id}`
                  : undefined
                const row = (
                  <>
                    <Status
                      status={workflow.status?.status || run?.status || 'pending'}
                    />
                    <Text variant="subtext">
                      {runContextSentence({
                        mode: runMode(run),
                        install: runInstall(run),
                        needsInput: false,
                        oneOff: !!(
                          run?.preview?.override_preview_config?.mode ||
                          run?.preview?.override_preview_config?.install_id
                        ),
                      })}
                    </Text>
                    <span className="ml-auto">
                      <Time
                        time={workflow.created_at || run?.created_at}
                        format="relative"
                        variant="subtext"
                        theme="neutral"
                      />
                    </span>
                  </>
                )
                return href ? (
                  <Link
                    key={workflow.id}
                    href={href}
                    className="!flex !w-full items-center gap-3 px-3 py-2 hover:bg-black/5 dark:hover:bg-white/5"
                  >
                    {row}
                  </Link>
                ) : (
                  <div
                    key={run?.id}
                    className="flex items-center gap-3 px-3 py-2"
                  >
                    {row}
                  </div>
                )
              })}
            </div>
          </div>
        ) : null}
      </div>
    </DetailPage>
  )
}
