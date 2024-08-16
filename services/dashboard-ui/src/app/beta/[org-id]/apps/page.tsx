import { Card, DataTable, Heading, Text, Link } from '@/components'
import { getApps, getOrg } from '@/lib'

export default async function Apps({ params }) {
  const orgId = params?.['org-id'] as string
  const org = await getOrg({ orgId })
  const apps = await getApps({ orgId })

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
    <>
      <header>
        <Heading>{org.name} / Apps</Heading>
      </header>
      <section>
        <DataTable
          headers={['ID', 'Name', 'Platform', 'Sandbox', 'Runner']}
          initData={tableData}
        />
      </section>
    </>
  )
}
