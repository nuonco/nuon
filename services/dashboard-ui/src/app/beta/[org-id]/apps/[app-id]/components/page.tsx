import { DataTable } from '@/components'
import { getAppComponents } from '@/lib'

export default async function AppComponents({ params }) {
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
  const components = await getAppComponents({ appId, orgId })

  const tableData = components.reduce((acc, component) => {
    acc.push([
      component.id,
      component.name,
      component.dependencies?.length || 0,
      component?.status,
      component.config_versions,
      `/beta/${orgId}/apps/${appId}/components/${component.id}`,
    ])

    return acc
  }, [])

  return (
    <section className="px-6 py-8">
      <DataTable
        headers={['ID', 'Name', 'Dependencies', 'Build', 'Config']}
        initData={tableData}
      />
    </section>
  )
}
