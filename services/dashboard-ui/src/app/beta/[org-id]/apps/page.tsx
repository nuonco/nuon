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
      app.sandbox_config?.aws_region_type,
      app?.runner_config?.app_runner_type,
      `/beta/${orgId}/apps/${app.id}`,
    ])

    return acc
  }, [])

  return (
    <DashboardContent breadcrumb={[org.name, 'Apps']}>
      <section className="px-6 py-8">
        <DataTable
          headers={['ID', 'Name', 'Platform', 'Sandbox', 'Runner']}
          initData={tableData}
        />
      </section>
    </DashboardContent>
  )
}
