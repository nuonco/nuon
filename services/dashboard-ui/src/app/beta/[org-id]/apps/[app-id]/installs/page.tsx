import { DataTable } from '@/components'
import { getAppInstalls } from '@/lib'

export default async function AppInstalls({ params }) {
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
  const installs = await getAppInstalls({ appId, orgId })

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
    <section className="px-6 py-8">
      <DataTable
        headers={['ID', 'Name', 'Platform', 'Sandbox', 'Runner']}
        initData={tableData}
      />
    </section>
  )
}
