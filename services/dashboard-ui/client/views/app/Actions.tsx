import { ActionsTable } from '@/components/actions/ActionsTable'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'

export const Actions = () => {
  const { org } = useOrg()
  const { app } = useApp()

  return (
    <PageSection isScrollable>
      <PageTitle title={`Actions | ${app?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/actions`, text: 'Actions' },
        ]}
      />
      <HeadingGroup>
        <Text variant="base" weight="strong">
          App actions
        </Text>
      </HeadingGroup>
      <ActionsTable />
    </PageSection>
  )
}
