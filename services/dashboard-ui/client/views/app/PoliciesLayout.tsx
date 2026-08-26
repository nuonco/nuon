import { Outlet, useParams } from 'react-router'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { TabNav } from '@/components/navigation/TabNav'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'

export const PoliciesLayout = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { branchId } = useParams()

  const appBase = branchId
    ? `/${org?.id}/apps/${app?.id}/branches/${branchId}`
    : `/${org?.id}/apps/${app?.id}`
  const basePath = `${appBase}/policies`

  return (
    <PageSection>
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: basePath, text: 'Policies' },
        ]}
      />
      <SectionHeader
        title="App policies"
        description="Define validation rules that run against builds and deploys."
      />
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
