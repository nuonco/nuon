import { Outlet } from 'react-router'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { TabNav } from '@/components/navigation/TabNav'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'

export const PoliciesLayout = () => {
  const { org } = useOrg()
  const { app } = useApp()

  const basePath = `/${org?.id}/apps/${app?.id}/policies`

  return (
    <PageSection>
      <PageTitle title={`Policies | ${app?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/policies`, text: 'Policies' },
        ]}
      />
      <HeadingGroup>
        <Text variant="base" weight="strong">
          App policies
        </Text>
        <Text variant="subtext" theme="neutral">
          Define validation rules that run against builds and deploys.
        </Text>
      </HeadingGroup>
      <TabNav
        basePath={basePath}
        tabs={[
          { path: '/', text: 'Definitions' },
          { path: '/analytics', text: 'Analytics' },
        ]}
      />
      <Outlet />
    </PageSection>
  )
}
