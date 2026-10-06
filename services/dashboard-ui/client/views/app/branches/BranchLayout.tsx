import { useMemo } from 'react'
import { Outlet, useMatch, useParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { PageContent } from '@/components/layout/PageContent'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { SubNav } from '@/components/navigation/SubNav'
import { useApp } from '@/hooks/use-app'
import { useBranch } from '@/hooks/use-branch'
import { useBranchNavCounts } from '@/hooks/use-branch-nav-counts'
import { useNewAppIA } from '@/hooks/use-new-app-ia'
import { useOrg } from '@/hooks/use-org'
import { BranchProvider } from '@/providers/branch-provider'
import { AppBranchSwitcher } from '@/components/branches/AppBranchSwitcher'
import { BranchDetailActions } from '@/components/branches/BranchDetailActions'
import { BranchHeaderMeta } from '@/components/branches/BranchHeaderMeta'
import { BranchPendingApprovals } from '@/components/branches/BranchRunApproval'
import { WorkflowRunPanelHost } from '@/components/branches/WorkflowRunPanel'
import { getRunTitle } from '@/components/branches/shared/run-title'
import { getBranchWorkflowRun, getBranchWorkflowRuns } from '@/lib'
import { latestBranchConfig } from '@/utils/branch-utils'
import type { TAppBranchConfig, TNavItem } from '@/types'

const triggerLabel = (config?: TAppBranchConfig) => {
  const mode =
    !config?.run_config?.mode || config.run_config.mode === 'all'
      ? 'push'
      : config.run_config.mode

  switch (mode) {
    case 'on_tag':
    case 'on_tag_prefix':
      return `Tags matching ${config?.run_config?.tag_prefix ?? 'the configured prefix'}`
    case 'on_github_label':
      return `Merged pull requests labeled ${config?.run_config?.github_label ?? 'with the configured label'}`
    case 'manual_only':
      return 'Manual runs'
    default:
      return 'Every push'
  }
}

const BranchTemplate = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { branch } = useBranch()
  const params = useParams()
  const detailMatch = useMatch(
    '/:orgId/apps/:appId/branches/:branchId/:section/:detail/*'
  )
  const runRolloutMatch = useMatch(
    '/:orgId/apps/:appId/branches/:branchId/runs/:runId/rollout'
  )
  const isDetailRoute = !!detailMatch && !params.runId
  const branchId = params.branchId as string
  const orgId = org.id!
  const appId = app.id!
  const basePath = `/${orgId}/apps/${appId}/branches/${branchId}`

  const currentConfig = useMemo(() => latestBranchConfig(branch), [branch])
  const vcs =
    currentConfig?.connected_github_vcs_config ??
    currentConfig?.public_git_vcs_config
  const { data: latestRunsResult, isLoading: isLoadingLatestRun } = useQuery({
    queryKey: ['branch-latest-run', orgId, appId, branchId],
    queryFn: () =>
      getBranchWorkflowRuns({ orgId, appId, branchId, limit: 1, offset: 0 }),
    enabled: !!orgId && !!appId && !!branchId,
    refetchInterval: 5000,
    placeholderData: keepPreviousData,
  })

  const latestRun = latestRunsResult?.data?.[0]
  const { data: pinnedWorkflow } = useQuery({
    queryKey: ['branch-run', orgId, appId, branchId, params.runId],
    queryFn: () =>
      getBranchWorkflowRun({
        orgId,
        appId,
        branchId,
        runId: params.runId!,
      }),
    enabled: !!orgId && !!appId && !!branchId && !!params.runId,
  })
  const hasDeploymentPlan = (currentConfig?.install_groups?.length ?? 0) > 0
  const showTriggerNudge =
    hasDeploymentPlan && !isLoadingLatestRun && !latestRun
  const hasInstallSyncing = !!org?.features?.['app-install-syncing']
  const navCounts = useBranchNavCounts({ orgId, appId, branchId })

  const navLinks: TNavItem[] = [
    { path: `/`, iconVariant: 'GraphIcon', text: 'Overview' },
    { path: `/installs`, iconVariant: 'CubeIcon', text: 'Installs' },
    { path: `/rollout`, iconVariant: 'StackIcon', text: 'Rollout' },
    { path: `/runs`, iconVariant: 'ListIcon', text: 'Previous runs' },
    ...(hasInstallSyncing
      ? [
          {
            path: `/install-configs`,
            iconVariant: 'ArrowsClockwiseIcon' as const,
            text: 'Install configs',
          },
        ]
      : []),
    { path: `/settings`, iconVariant: 'GearIcon', text: 'Settings' },
    {
      type: 'section',
      label: 'App template',
      defaultOpen: true,
      collapsible: false,
    },
    {
      path: `/inputs`,
      iconVariant: 'ListChecksIcon',
      text: 'Inputs',
      count: navCounts.inputs,
    },
    {
      path: `/components`,
      iconVariant: 'CardsIcon',
      text: 'Components',
      count: currentConfig?.component_ids?.length,
    },
    {
      path: `/actions`,
      iconVariant: 'TerminalWindowIcon',
      text: 'Actions',
      count: currentConfig?.action_ids?.length,
    },
    {
      path: `/runbooks`,
      iconVariant: 'BookIcon',
      text: 'Runbooks',
      count: currentConfig?.runbook_ids?.length,
    },
    {
      path: `/sandbox`,
      iconVariant: 'ShippingContainerIcon',
      text: 'Sandboxes',
    },
    {
      path: `/policies`,
      iconVariant: 'ShieldCheckIcon',
      text: 'Policies',
      count: navCounts.policies,
    },
    {
      path: `/roles`,
      iconVariant: 'FileLockIcon',
      text: 'Roles',
      count: navCounts.roles,
    },
    {
      path: `/labels`,
      iconVariant: 'TagIcon',
      text: 'Labels',
      count: navCounts.labels,
    },
    { path: `/readme`, iconVariant: 'BookOpenIcon', text: 'README' },
  ]
  // Run detail renders its own BranchRunApproval; layout banner is for other routes.
  const approvalRun = params.runId ? null : latestRun

  return (
    <>
      {!isDetailRoute ? (
        <Breadcrumbs
          breadcrumbs={[
            { path: `/${orgId}`, text: org.name },
            { path: `/${orgId}/apps`, text: 'Apps' },
            { path: `/${orgId}/apps/${appId}`, text: app.name },
            { path: basePath, text: branch.name },
            ...(params.runId
              ? [
                  { path: `${basePath}/runs`, text: 'Previous runs' },
                  {
                    path: `${basePath}/runs/${params.runId}`,
                    text: pinnedWorkflow ? getRunTitle(pinnedWorkflow) : 'Run',
                  },
                ]
              : []),
            ...(runRolloutMatch
              ? [
                  {
                    path: `${basePath}/runs/${params.runId}/rollout`,
                    text: 'Rollout',
                  },
                ]
              : []),
          ]}
        />
      ) : null}
      <DetailHeader
        variant="page"
        backLink={false}
        title={app.name}
        identity={
          <BranchHeaderMeta
            configuration={<AppBranchSwitcher />}
            repo={vcs?.repo}
            gitBranch={vcs?.branch}
            directory={vcs?.directory}
            trigger={triggerLabel(currentConfig)}
          />
        }
        actions={
          <BranchDetailActions
            branch={branch}
            currentConfig={currentConfig}
            appId={appId}
            orgId={orgId}
            showTriggerNudge={showTriggerNudge}
          />
        }
      />
      <PageContent className="border-t" variant="row">
        <SubNav
          basePath={basePath}
          links={navLinks}
          storageKey="subnav:branch"
          pinLastGroup
        />
        <div className="flex flex-col flex-1 min-w-0">
          {approvalRun ? (
            <BranchPendingApprovals
              run={approvalRun}
              className="px-4 md:px-6 pt-4 md:pt-6"
            />
          ) : null}
          <Outlet />
        </div>
      </PageContent>
      <WorkflowRunPanelHost />
    </>
  )
}

export const BranchLayout = () => {
  const hasNewAppIA = useNewAppIA()
  const params = useParams()
  const branchId = params.branchId as string

  if (!hasNewAppIA) return <Outlet />

  return (
    <BranchProvider branchId={branchId} shouldPoll>
      <BranchTemplate />
    </BranchProvider>
  )
}
