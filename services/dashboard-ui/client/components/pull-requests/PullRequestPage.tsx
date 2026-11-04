import { useMemo } from 'react'
import { useNavigate, useParams } from 'react-router'
import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import { BackLink } from '@/components/common/BackLink'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { ProviderError } from '@/components/layout/ProviderError'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { Toast } from '@/components/surfaces/Toast'
import { useApp } from '@/hooks/use-app'
import { useBranch } from '@/hooks/use-branch'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { getBranchWorkflowRuns, getPreviewSources, triggerBranchRun } from '@/lib'
import type { TAPIError, TAppBranchPreviewConfig } from '@/types'
import { latestBranchConfig } from '@/utils/branch-utils'
import { EditPreviewConfigModal } from './EditPreviewConfigModal'
import { PullRequestConfigChanges } from './PullRequestConfigChanges'
import { pullRequestRuns, toPlaygroundPullRequest } from './map-pull-request'
import { PullRequestPreview } from './PullRequestPreview'
import { RerunPreviewModal } from './RerunPreviewModal'

const RUN_LIMIT = 100

export const PullRequestPage = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { branch } = useBranch()
  const { addToast } = useToast()
  const { addModal } = useSurfaces()
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const params = useParams()
  const orgId = org?.id ?? ''
  const appId = app?.id ?? ''
  const branchId = params.branchId ?? ''
  const prNumber = Number(params.number)
  const basePath = `/${orgId}/apps/${appId}/branches/${branchId}`
  const currentConfig = latestBranchConfig(branch)
  const vcs =
    currentConfig?.connected_github_vcs_config ??
    currentConfig?.public_git_vcs_config

  const sources = useQuery({
    queryKey: ['preview-sources', orgId, appId, branchId],
    queryFn: () => getPreviewSources({ orgId, appId, branchId }),
    enabled: !!orgId && !!appId && !!branchId,
  })

  const runs = useQuery({
    queryKey: ['branch-preview-runs', orgId, appId, branchId, prNumber],
    queryFn: () =>
      getBranchWorkflowRuns({
        orgId,
        appId,
        branchId,
        limit: RUN_LIMIT,
        offset: 0,
      }),
    enabled: !!orgId && !!appId && !!branchId && Number.isFinite(prNumber),
    refetchInterval: 5000,
    placeholderData: keepPreviousData,
  })

  const source = sources.data?.pull_requests?.find(
    (pullRequest) => pullRequest.pr_number === prNumber
  )
  const playgroundRuns = pullRequestRuns(runs.data?.data ?? [], prNumber)
  const previewWorkflows = useMemo(
    () =>
      [...(runs.data?.data ?? [])]
        .filter((workflow) => {
          const branchRun = workflow.app_branch_runs?.[0]
          const number = branchRun?.pr_number ?? branchRun?.metadata?.pr_number
          return number === prNumber
        })
        .sort(
          (a, b) =>
            Date.parse(b.created_at || '') - Date.parse(a.created_at || '')
        ),
    [runs.data?.data, prNumber]
  )
  const latestWorkflow = previewWorkflows[0]
  const latestRun = latestWorkflow?.app_branch_runs?.[0]

  const pullRequest =
    source && Number.isFinite(prNumber)
      ? toPlaygroundPullRequest({
          source,
          orgId,
          appId,
          branchId,
          repo: vcs?.repo,
          baseRef: vcs?.branch,
          branchConfig: currentConfig?.preview_config,
          runs: playgroundRuns,
          latest: latestRun,
        })
      : undefined

  const refresh = () => {
    queryClient.invalidateQueries({
      queryKey: ['branch-preview-runs', orgId, appId, branchId, prNumber],
    })
  }

  const trigger = useMutation({
    mutationFn: (override?: TAppBranchPreviewConfig) => {
      const resolved = {
        ...pullRequest?.resolved_preview_config,
        ...override,
      }
      const mode = resolved.mode
      return triggerBranchRun({
        orgId,
        appId,
        branchId,
        request: {
          preview_run: {
            source: 'pr',
            pr_number: prNumber,
            head_sha: pullRequest?.head_sha || undefined,
            git_ref: pullRequest?.head_ref || undefined,
            mode: mode === 'none' ? undefined : mode,
            install_id:
              mode === 'build-only' || mode === 'none'
                ? undefined
                : resolved.install_id,
          },
        },
      })
    },
    onSuccess: () => {
      refresh()
      addToast(
        <Toast heading="Preview run started" theme="info">
          <Text>Running the preview for #{prNumber}.</Text>
        </Toast>
      )
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Preview run failed" theme="error">
          <Text>{err?.error || 'Unable to start the preview run.'}</Text>
        </Toast>
      )
    },
  })

  const title = Number.isFinite(prNumber) ? `#${prNumber}` : 'Pull request'
  const backLink = (
    <BackLink onClick={() => navigate(`${basePath}/pull-requests`)} />
  )

  if (sources.error && !sources.data) {
    return (
      <PageSection>
        <PageTitle segments={[title, app?.name]} />
        {backLink}
        <ProviderError error={sources.error as TAPIError} />
      </PageSection>
    )
  }

  return (
    <>
      <PageTitle segments={[title, app?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${orgId}`, text: org?.name },
          { path: `/${orgId}/apps`, text: 'Apps' },
          { path: `/${orgId}/apps/${appId}`, text: app?.name },
          { path: basePath, text: branch?.name },
          { path: `${basePath}/pull-requests`, text: 'Pull requests' },
          {
            path: `${basePath}/pull-requests/${prNumber}`,
            text: title,
          },
        ]}
      />
      {sources.isPending ? (
        <PullRequestPreview basePath={basePath} isLoading backLink={backLink} />
      ) : !source ? (
        <PageSection>
          {backLink}
          <Text variant="body" theme="neutral">
            Pull request #{params.number} is not open against this branch.
          </Text>
        </PageSection>
      ) : (
        <PullRequestPreview
          source={source}
          workflows={previewWorkflows}
          repo={vcs?.repo}
          baseRef={vcs?.branch}
          branchConfig={currentConfig?.preview_config}
          basePath={basePath}
          orgId={orgId}
          appId={appId}
          backLink={backLink}
          isLoading={runs.isPending}
          changes={
            latestRun?.id ? (
              <PullRequestConfigChanges
                branchId={branchId}
                appBranchRunId={latestRun.id}
              />
            ) : null
          }
          onEditConfig={() =>
            pullRequest
              ? addModal(
                  <EditPreviewConfigModal
                    pullRequest={pullRequest}
                    orgId={orgId}
                    appId={appId}
                    branchId={branchId}
                    configId={currentConfig?.id}
                    onSave={(override) => trigger.mutateAsync(override)}
                  />
                )
              : undefined
          }
          onRerun={() =>
            addModal(
              <RerunPreviewModal
                prNumber={prNumber}
                onConfirm={() => trigger.mutateAsync(undefined)}
              />
            )
          }
        />
      )}
    </>
  )
}
