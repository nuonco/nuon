import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { AccessRolesTable } from '@/components/access-roles/AccessRolesTable'
import { CreateRoleButton } from '@/components/access-roles/RoleForm'

import { useOrg } from '@/hooks/use-org'

export const AccessRoles = () => {
  const { org } = useOrg()

  return (
    <>
      <PageTitle title={`Roles | ${org?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${org.id}`,
            text: org?.name,
          },
          {
            path: `/${org.id}/settings`,
            text: 'Settings',
          },
          {
            path: `/${org.id}/settings/roles`,
            text: 'Roles',
          },
        ]}
      />
      <PageHeader className="flex items-center justify-between">
        <HeadingGroup>
          <Text variant="h3" weight="stronger" level={1}>
            Roles
          </Text>
          <Text theme="neutral">
            Control what members, service accounts, and tokens can do in this
            org.
          </Text>
        </HeadingGroup>
        <CreateRoleButton />
      </PageHeader>
      <PageContent>
        <PageSection>
          <AccessRolesTable />
        </PageSection>
      </PageContent>
    </>
  )
}
