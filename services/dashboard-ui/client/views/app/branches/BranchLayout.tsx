import { useMemo } from 'react'
import { Outlet, useMatch, useParams, useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { PageContent } from '@/components/layout/PageContent'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { SubNav } from '@/components/navigation/SubNav'
import { useApp } from '@/hooks/use-app'
import { useBranch } from '@/hooks/use-branch'
import { useNewAppIA } from '@/hooks/use-new-app-ia'
import { useOrg } from '@/hooks/use-org'
import { BranchProvider } from '@/providers/branch-provider'
import { AppBranchSwitcher } from '@/components/branches/AppBranchSwitcher'
import { BranchVcsBadges } from '@/components/branches/BranchVcsBadges'
import { BranchDetailActions } from '@/components/branches/BranchDetailActions'
import { BranchPendingApprovals } from '@/components/branches/BranchRunApproval'
import {
  BranchSettingsPanel,
  useOpenBranchSettings,
  BRANCH_SETTINGS_PANEL_KEY,
} from '@/components/branches/BranchSettingsPanel'
import { getBranchWorkflowRuns } from '@/lib'
import { latestBranchConfig } from '@/utils/branch-utils'
import type { TNavItem } from '@/types/dashboard.types'

const BranchTemplate = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { branch } = useBranch()
  const params = useParams()
  const isDetailRoute = !!useMatch(
    '/:orgId/apps/:appId/branches/:branchId/:section/:detail/*'
  )
  const openSettings = useOpenBranchSettings()
  const [searchParams] = useSearchParams()
  const isSettingsOpen =
    searchParams.get('panel') === BRANCH_SETTINGS_PANEL_KEY
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
  const hasDeploymentPlan = (currentConfig?.install_groups?.length ?? 0) > 0
  const showTriggerNudge =
    hasDeploymentPlan && !isLoadingLatestRun && !latestRun

  const hasInstallSyncing = !!org?.features?.['app-install-syncing']

  const navLinks: TNavItem[] = [
    { path: `/`, iconVariant: 'HouseSimpleIcon', text: 'Overview' },
    { path: `/runs`, iconVariant: 'PlayIcon', text: 'Updates' },
    { type: 'section', label: 'Installs', defaultOpen: false },
    { path: `/runs`, iconVariant: 'PlayIcon', text: 'Runs' },
    { path: `/plan`, iconVariant: 'TreeStructureIcon', text: 'Deployment plan' },
    { path: `/preview`, iconVariant: 'EyeIcon', text: 'Preview' },
    { type: 'section', label: 'Definition' },
    { path: `/components`, iconVariant: 'CardsIcon', text: 'Components' },
    { path: `/actions`, iconVariant: 'TerminalWindowIcon', text: 'Actions' },
    { path: `/runbooks`, iconVariant: 'BookIcon', text: 'Runbooks' },
    { path: `/sandbox`, iconVariant: 'ShippingContainerIcon', text: 'Sandbox builds' },
    { type: 'section', label: 'Distribution' },
    { path: `/installs`, iconVariant: 'CubeIcon', text: 'Installs' },
    { path: `/plan`, iconVariant: 'TreeStructureIcon', text: 'Install groups' },
    ...(hasInstallSyncing
      ? [
        {
          path: `/install-configs`,
          iconVariant: 'ArrowsClockwiseIcon' as const,
          text: 'Install configs',
        },
      ]
      : []),
    { type: 'section', label: 'Template' },
    { path: `/components`, iconVariant: 'CardsIcon', text: 'Components' },
    { path: `/actions`, iconVariant: 'TerminalWindowIcon', text: 'Actions' },
    { path: `/runbooks`, iconVariant: 'BookIcon', text: 'Runbooks' },
    { path: `/sandbox`, iconVariant: 'ShippingContainerIcon', text: 'Sandboxes' },
    { path: `/policies`, iconVariant: 'ShieldCheckIcon', text: 'Policies' },
    { path: `/roles`, iconVariant: 'FileLockIcon', text: 'Roles' },
    { type: 'section', label: 'Configuration' },
    { path: `/labels`, iconVariant: 'TagIcon', text: 'Labels' },
    { path: `/readme`, iconVariant: 'BookOpenIcon', text: 'README' },
    {
      type: 'action',
      key: 'settings',
      iconVariant: 'GearIcon',
      text: 'Settings',
      onClick: openSettings,
      isActive: isSettingsOpen,
    },
  ]

  return (
    <>
      {!isDetailRoute ? (
        <>
          <Breadcrumbs
            breadcrumbs={[
              { path: `/${orgId}`, text: org.name },
              { path: `/${orgId}/apps`, text: 'Apps' },
              { path: `/${orgId}/apps/${appId}`, text: app.name },
              { path: basePath, text: branch.name },
            ]}
          />
          <DetailHeader
            variant="page"
            backLink={false}
            title={app.name}
            status={<AppBranchSwitcher />}
            identity={
              <>
                <BranchVcsBadges repo={vcs?.repo} branch={vcs?.branch} />
                <Text variant="subtext" theme="info">
                  Last updated{' '}
                  <Time
                    variant="subtext"
                    time={branch.updated_at}
                    format="relative"
                  />
                </Text>
              </>
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
        </>
      ) : null}
      <BranchSettingsPanel />
      <PageContent className="border-t" variant="row">
        <SubNav basePath={basePath} links={navLinks} storageKey="subnav:branch" />
        <div className="flex flex-col flex-1 min-w-0">
          {latestRun && params.runId !== latestRun.id ? (
            <BranchPendingApprovals
              run={latestRun}
              runHref={`${basePath}/runs/${latestRun.id}`}
              className="px-4 md:px-6 pt-4 md:pt-6"
            />
          ) : null}
          <Outlet />
        </div>
      </PageContent>
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
