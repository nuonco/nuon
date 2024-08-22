import { DashboardContent, DataTable } from '@/components'
import { getOrg, getInstalls } from '@/lib'

export default async function Installs({ params }) {
  const orgId = params?.['org-id'] as string
  const [installs, org] = await Promise.all([
    getInstalls({ orgId }),
    getOrg({ orgId }),
  ])

  const tableData = installs.reduce((acc, install) => {
    acc.push([
      install.id,
      install.name,
      install.app_sandbox_config?.cloud_platform,
      install.sandbox_status,
      install.runner_status,
      `/beta/${orgId}/installs/${install.id}`,
    ])

    return acc
  }, [])

  return (
    <DashboardContent breadcrumb={[org.name, 'Installs']}>
      <section className="px-6 py-8">
        <DataTable
          headers={['ID', 'Name', 'App', 'Platform', 'Sandbox']}
          initData={tableData}
        />
      </section>
    </DashboardContent>
  )
}
