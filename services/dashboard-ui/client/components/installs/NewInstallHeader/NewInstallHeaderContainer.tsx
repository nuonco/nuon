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
import { getAppBranch } from '@/lib'
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
  const desiredCommit = tracking?.selected_commit?.workflow_id
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

  if (!install) return null

  const repo =
    tracking?.repo ||
    (branch ? vcsRepo(latestBranchConfig(branch)) : undefined)

  return (
    <NewInstallHeader
      install={install}
      labelColors={labelColors}
      latestCommit={
        desiredCommit ? (
          <BranchRunCommit
            displayVariant="inline"
            status={desiredCommit.run_status}
            href={branchRunHref({
              orgId: org?.id,
              appId: install.app_id,
              commit: desiredCommit,
            })}
            message={desiredCommit.message?.split('\n')[0]}
            author={desiredCommit.author}
            sha={desiredCommit.sha}
            repo={repo}
            createdAt={desiredCommit.created_at}
            showStatus={false}
          />
        ) : undefined
      }
      latestCommitLabel="Desired commit"
      latestCommitLoading={isLoadingOverview}
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
