import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { InstallComponentsTable } from '@/components/install-components/InstallComponentsTable'
import { ManageAllDropdown } from '@/components/install-components/management/ManageAllDropdown'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'

export const Components = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  return (
    <PageSection>
      <PageTitle title={`Components | ${install?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/components`,
            text: 'Components',
          },
        ]}
      />
      <div className="flex items-start justify-between gap-4">
        <HeadingGroup>
          <Text variant="base" weight="strong">
            Install components
          </Text>
          <Text variant="subtext" theme="neutral">
            View and manage all components for this install.
          </Text>
        </HeadingGroup>
        <div className="shrink-0">
          <ManageAllDropdown />
        </div>
      </div>

      <InstallComponentsTable shouldPoll />
    </PageSection>
  )
}
