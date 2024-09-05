import { DashboardContent, DataTable } from '@/components'
import { getApps, getOrg } from '@/lib'

export default async function Apps({ params }) {
  const orgId = params?.['org-id'] as string
  const [apps, org] = await Promise.all([getApps({ orgId }), getOrg({ orgId })])

  const tableData = apps.reduce((acc, app) => {
    acc.push([
      app.id,
      app.name,
      app.cloud_platform,
      app?.runner_config?.app_runner_type,
      `/beta/${orgId}/apps/${app.id}`,
    ])

    return acc
  }, [])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/beta/${org.id}`, text: org.name },
        { href: `/beta/${org.id}/apps`, text: 'Apps' },
      ]}
    >
      <section className="px-6 py-8">
        <DataTable
          headers={['ID', 'Name', 'Platform', 'Runner']}
          initData={tableData}
        />
      </section>
    </DashboardContent>
  )
}
