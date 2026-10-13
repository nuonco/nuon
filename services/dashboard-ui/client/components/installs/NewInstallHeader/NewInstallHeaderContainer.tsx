import { useQuery } from '@tanstack/react-query'
import { useLocation } from 'react-router'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { BranchRunCommit } from '@/components/branches/BranchRunCommit'
import { InstallStatusSummary } from '@/components/installs/InstallStatusSummary'
import { useOpenInstallSettings } from '@/components/installs/InstallSettingsPanel'
import { ChangeAppBranchButton } from '@/components/installs/management/ChangeAppBranch'
import { useCurrentAppBranchRun } from '@/hooks/use-current-app-branch-run'
import { useInstallHref } from '@/hooks/use-install-path'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getAppBranch, getBranchWorkflowRuns } from '@/lib'
import { latestBranchConfig } from '@/utils/branch-utils'
import { vcsRepo } from '@/utils/vcs-urls'
import { NewInstallHeader } from './NewInstallHeader'

export const NewInstallHeaderContainer = () => {
  const { org } = useOrg()
  const { install, labelColors, refresh } = useInstall()
  const { pathname } = useLocation()
  const installHref = useInstallHref()
  const { run: appliedRun, isLoading: isLoadingRun } = useCurrentAppBranchRun()
  const openSettings = useOpenInstallSettings()

  const branchId = install?.app_branch?.id

  const { data: branch } = useQuery({
    queryKey: ['install-header-branch', org?.id, install?.app_id, branchId],
    queryFn: () =>
      getAppBranch({
        orgId: org!.id,
        appId: install!.app_id!,
        branchId: branchId!,
        latestConfig: true,
      }),
    enabled: !!org?.id && !!install?.app_id && !!branchId,
  })

  const { data: branchRuns, isLoading: isLoadingBranchRuns } = useQuery({
    queryKey: ['install-header-branch-run', org?.id, install?.app_id, branchId],
    queryFn: () =>
      getBranchWorkflowRuns({
        orgId: org.id,
        appId: install.app_id!,
        branchId: branchId!,
        limit: 1,
      }),
    enabled:
      !!org?.id &&
      !!install?.app_id &&
      !!branchId &&
      !isLoadingRun &&
      !appliedRun,
  })

  if (!install) return null

  const branchRun = branchRuns?.data?.[0]?.app_branch_runs?.at(0)
  const run = appliedRun ?? branchRun
  const commit = run?.vcs_connection_commit
  const repo =
    vcsRepo(run?.app_branch_config) ??
    (branch ? vcsRepo(latestBranchConfig(branch)) : undefined)
  const runBranchId = run?.app_branch?.id ?? branchId
  const runHref =
    org?.id && install.app_id && runBranchId && run?.id
      ? `/${org.id}/apps/${install.app_id}/branches/${runBranchId}/runs/${run.id}`
      : undefined

  return (
    <NewInstallHeader
      install={install}
      labelColors={labelColors}
      latestCommit={
        run ? (
          <BranchRunCommit
            displayVariant="inline"
            status={run.status}
            href={runHref}
            message={commit?.message?.split('\n')[0]}
            author={commit?.author_name}
            avatarUrl={commit?.author_avatar_url}
            sha={commit?.sha ?? run.head_sha}
            repo={repo}
            createdAt={run.created_at}
            showStatus={false}
          />
        ) : undefined
      }
      latestCommitLabel={appliedRun ? 'Applied commit' : 'Latest branch run'}
      latestCommitLoading={isLoadingRun || isLoadingBranchRuns}
      installPath={installHref({
        installId: install.id,
        appId: install.app_id,
      })}
      orgId={org?.id}
      branchAction={
        <ChangeAppBranchButton iconOnly install={install} onSuccess={refresh} />
      }
      settingsAction={
        <Button
          variant="secondary"
          onClick={openSettings}
          aria-label="Install settings"
        >
          <Icon variant="GearIcon" size={16} />
        </Button>
      }
      statuses={
        /\/health\/?$/.test(pathname) ? undefined : <InstallStatusSummary />
      }
    />
  )
}
