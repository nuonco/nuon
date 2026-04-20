import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { AppRolesTable } from '@/components/roles/AppRolesTable'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'

export const Roles = () => {
  const { org } = useOrg()
  const { app } = useApp()

  return (
    <PageSection>
      <PageTitle title={`Roles | ${app?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/roles`, text: 'Roles' },
        ]}
      />
      <HeadingGroup>
        <Text variant="base" weight="strong">
          IAM roles
        </Text>
        <Text variant="subtext" theme="neutral">
          View the IAM roles that your app uses to access customer AWS resources.
        </Text>
      </HeadingGroup>

      <AppRolesTable />
    </PageSection>
  )
}
