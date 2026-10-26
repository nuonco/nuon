import { useQuery } from '@tanstack/react-query'
import { useLocation } from 'react-router'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { BranchRunCommit } from '@/components/branches/BranchRunCommit'
import { InstallStatusSummary } from '@/components/installs/InstallStatusSummary'
import { useOpenInstallSettings } from '@/components/installs/InstallSettingsPanel'
import { installAppBranchGroup } from '@/components/branches/active-app-branch-connection'
import { ChangeAppBranchButton } from '@/components/installs/management/ChangeAppBranch'
import { ChangeInstallGroupButton } from '@/components/installs/management/ChangeInstallGroup'
import { branchRunHref } from '@/components/install-alignment/alignment-utils'
import { useInstallHref } from '@/hooks/use-install-path'
import { useInstall } from '@/hooks/use-install'
import { useInstallOverview } from '@/hooks/use-install-overview'
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
  const { data: overview, isLoading: isLoadingOverview } = useInstallOverview()
  const openSettings = useOpenInstallSettings()

  const branchId = install?.app_branch?.id
  const tracking = overview?.branch_tracking
  const selectedCommit = tracking?.selected_commit?.sha
    ? tracking.selected_commit
    : undefined

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
      !isLoadingOverview &&
      !selectedCommit,
  })

  if (!install) return null

  const latestWorkflow = branchRuns?.data?.[0]
  const branchRun = latestWorkflow?.app_branch_runs?.at(0)
  const commit = branchRun?.vcs_connection_commit
  const repo =
    tracking?.repo ||
    vcsRepo(branchRun?.app_branch_config) ||
    (branch ? vcsRepo(latestBranchConfig(branch)) : undefined)
  const latestRunHref =
    org?.id && install.app_id && branchId && latestWorkflow?.id
      ? `/${org.id}/apps/${install.app_id}/branches/${branchId}/runs/${latestWorkflow.id}`
      : undefined

  return (
    <NewInstallHeader
      install={install}
      labelColors={labelColors}
      latestCommit={
        selectedCommit ? (
          <BranchRunCommit
            displayVariant="inline"
            status={selectedCommit.run_status}
            href={branchRunHref({
              orgId: org?.id,
              appId: install.app_id,
              commit: selectedCommit,
            })}
            message={selectedCommit.message?.split('\n')[0]}
            author={selectedCommit.author}
            sha={selectedCommit.sha}
            repo={repo}
            createdAt={selectedCommit.created_at}
            showStatus={false}
          />
        ) : branchRun ? (
          <BranchRunCommit
            displayVariant="inline"
            status={branchRun.status}
            href={latestRunHref}
            message={commit?.message?.split('\n')[0]}
            author={commit?.author_name}
            avatarUrl={commit?.author_avatar_url}
            sha={commit?.sha ?? branchRun.head_sha}
            repo={repo}
            createdAt={branchRun.created_at}
            showStatus={false}
          />
        ) : undefined
      }
      latestCommitLabel={
        selectedCommit ? 'Selected commit' : 'Latest branch run'
      }
      latestCommitLoading={isLoadingOverview || isLoadingBranchRuns}
      installPath={installHref({
        installId: install.id,
        appId: install.app_id,
      })}
      orgId={org?.id}
      branchAction={
        <ChangeAppBranchButton iconOnly install={install} onSuccess={refresh} />
      }
      groupName={
        install.app_branch ? installAppBranchGroup(install) : undefined
      }
      groupAction={
        install.app_branch ? (
          <ChangeInstallGroupButton
            iconOnly
            install={install}
            onSuccess={refresh}
          />
        ) : undefined
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
