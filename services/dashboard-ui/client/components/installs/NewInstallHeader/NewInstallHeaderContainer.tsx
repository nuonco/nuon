import { useQuery } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { BranchRunCommit } from '@/components/branches/BranchRunCommit'
import { InstallStatusesContainer } from '@/components/installs/InstallStatuses'
import { useOpenInstallSettings } from '@/components/installs/InstallSettingsPanel'
import { ChangeAppBranchButton } from '@/components/installs/management/ChangeAppBranch'
import { useCurrentAppBranchRun } from '@/hooks/use-current-app-branch-run'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getBranchWorkflowRuns } from '@/lib'
import { NewInstallHeader } from './NewInstallHeader'

export const NewInstallHeaderContainer = () => {
  const { org } = useOrg()
  const { install, labelColors, refresh } = useInstall()
  const { run: appliedRun, isLoading: isLoadingRun } = useCurrentAppBranchRun()
  const openSettings = useOpenInstallSettings()

  const branchId = install?.app_branch?.id

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
            createdAt={run.created_at}
            showStatus={false}
          />
        ) : undefined
      }
      latestCommitLabel={appliedRun ? 'Applied commit' : 'Latest branch run'}
      latestCommitLoading={isLoadingRun || isLoadingBranchRuns}
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
      statuses={<InstallStatusesContainer />}
    />
  )
}
