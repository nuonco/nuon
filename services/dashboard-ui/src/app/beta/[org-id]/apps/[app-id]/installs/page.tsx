import { DashboardContent, DataTable, SubNav, type TLink } from '@/components'
import { getApp, getAppInstalls, getOrg } from '@/lib'

export default async function AppInstalls({ params }) {
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
  const subNavLinks: Array<TLink> = [
    { href: `/beta/${orgId}/apps/${appId}`, text: 'Config' },
    { href: `/beta/${orgId}/apps/${appId}/components`, text: 'Components' },
    { href: `/beta/${orgId}/apps/${appId}/installs`, text: 'Installs' },
  ]

  const app = await getApp({ appId, orgId })
  const installs = await getAppInstalls({ appId, orgId })
  const org = await getOrg({ orgId })

  const tableData = installs.reduce((acc, install) => {
    acc.push([
      install.id,
      install.name,
      install.azure_account?.location ? 'AWS' : 'Azure',
      install.sandbox_status,
      install.runner_status,
      `/beta/${orgId}/installs/${install.id}`,
    ])

    return acc
  }, [])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/beta/${org.id}`, text: org.name },
        { href: `/beta/${org.id}/apps`, text: 'Apps' },
        { href: `/beta/${org.id}/apps/${app.id}`, text: app.name },
      ]}
      heading={app.name}
      headingUnderline={app.id}
      meta={<SubNav links={subNavLinks} />}
    >
      <section className="px-6 py-8">
        <DataTable
          headers={['ID', 'Name', 'Platform', 'Sandbox', 'Runner']}
          initData={tableData}
        />
      </section>
    </DashboardContent>
  )
}
