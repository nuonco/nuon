import { DataTable, Link } from '@/components'
import { getInstall } from '@/lib'

export default async function InstallComponents({ params }) {
  const orgId = params?.['org-id'] as string
  const installId = params?.['install-id'] as string
  const install = await getInstall({ orgId, installId })

  const tableData = install.install_components.reduce((acc, installComponent) => {
    acc.push([
      installComponent.id,
      installComponent.component.name,
      installComponent.install_deploys?.[0].status,
      installComponent.install_deploys?.[0].status,
      installComponent.component.config_versions || 0,
      `/beta/${orgId}/installs/${installId}/components/${installComponent.id}`,
    ])

    return acc
  }, [])

  return (
    <section className="px-6 py-8">
       <DataTable
        headers={['ID', 'Name', 'Deployment', 'Build', 'Config']}
        initData={tableData}
      />
    </section>
  )
}
