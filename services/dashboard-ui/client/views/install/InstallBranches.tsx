import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { InstallBranchesSection } from '@/components/installs/InstallBranches'
import { installAppBranchGroup } from '@/components/branches/active-app-branch-connection'
import { ChangeAppBranchButton } from '@/components/installs/management/ChangeAppBranch'
import { ChangeInstallGroupButton } from '@/components/installs/management/ChangeInstallGroup'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { PageSection } from '@/components/layout/PageSection'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrgFeatureFlag } from '@/hooks/use-org-feature-flag'

export const InstallBranches = () => {
  const { install, refresh } = useInstall()
  const appSyncDisabled = useOrgFeatureFlag('disable-app-sync')
  const branchId = install?.app_branch_id
  const branchName = install?.app_branch?.name ?? branchId

  return (
    <PageSection>
      <PageTitle segments={['App branches', install?.name]} />

      <SectionHeader
        title="Branch configuration"
        description="The app branch determines which app config and deployment plan this install uses."
      />

      <Card className="!p-4 !gap-4">
        <div className="flex items-start justify-between gap-4 flex-wrap">
          <div className="flex items-start gap-8 flex-wrap">
            <LabeledValue label="App branch">
              {branchName ? (
                <div className="flex items-center gap-1.5">
                  <Icon
                    variant="GitBranchIcon"
                    size={14}
                    className="text-cool-grey-400"
                  />
                  <Text variant="body" family="mono">
                    {branchName}
                  </Text>
                </div>
              ) : (
                <Text variant="subtext" theme="neutral">
                  None
                </Text>
              )}
            </LabeledValue>

            {branchId && (
              <LabeledValue label="Install group">
                <Text
                  variant="body"
                  family="mono"
                  theme={
                    install && installAppBranchGroup(install)
                      ? undefined
                      : 'neutral'
                  }
                >
                  {(install && installAppBranchGroup(install)) || 'None'}
                </Text>
              </LabeledValue>
            )}
          </div>

          {install && (
            <div className="flex items-center gap-2">
              <ChangeInstallGroupButton install={install} onSuccess={refresh} />
              <ChangeAppBranchButton install={install} onSuccess={refresh} />
            </div>
          )}
        </div>

        {!branchId && (
          <Text variant="subtext" theme="neutral">
            {appSyncDisabled
              ? 'This install has no app branch. App sync is disabled for this org, so move it to an app branch to get config changes.'
              : 'This install uses the latest sync config. Branch-based deployments and install groups are not available until you move it to an app branch.'}
          </Text>
        )}
      </Card>

      <SectionHeader
        title="Connected branches"
        description="Branches this install is connected to and their latest run status."
      />

      <InstallBranchesSection install={install} />
    </PageSection>
  )
}
