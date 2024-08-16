import { DataTable, Heading, Link } from '@/components'
import { getOrg, getInstalls } from '@/lib'

export default async function Installs({ params }) {
  const orgId = params?.['org-id'] as string
  const org = await getOrg({ orgId })
  const installs = await getInstalls({ orgId })

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
    <>
      <header>
        <Heading>{org.name} / Installs</Heading>
      </header>
      <section>
        <DataTable
          headers={['ID', 'Name', 'App', 'Platform', 'Sandbox']}
          initData={tableData}
        />
      </section>
    </>
  )
}
