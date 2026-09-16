import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { InstallBranchesSection } from '@/components/installs/InstallBranches'
import { ChangeAppBranchButton } from '@/components/installs/management/ChangeAppBranch'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { PageSection } from '@/components/layout/PageSection'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'

export const InstallBranches = () => {
  const { install, refresh } = useInstall()
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

          {install && (
            <ChangeAppBranchButton install={install} onSuccess={refresh} />
          )}
        </div>

        {!branchId && (
          <Text variant="subtext" theme="neutral">
            This install uses the most recent config from{' '}
            <code>nuon apps sync</code>. Branch-based deployments and install
            groups are not available until you explicitly move it to a branch.
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
