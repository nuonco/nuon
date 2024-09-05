import {
  DashboardContent,
  DataTable,
  InstallStatus,
  SubNav,
  type TLink,
} from '@/components'
import { InstallProvider } from '@/context'
import { getInstall, getOrg } from '@/lib'

export default async function InstallComponents({ params }) {
  const orgId = params?.['org-id'] as string
  const installId = params?.['install-id'] as string
  const subNavLinks: Array<TLink> = [
    { href: `/beta/${orgId}/installs/${installId}`, text: 'Status' },
    {
      href: `/beta/${orgId}/installs/${installId}/components`,
      text: 'Components',
    },
  ]

  const [install, org] = await Promise.all([
    getInstall({ orgId, installId }),
    getOrg({ orgId }),
  ])

  const tableData =
    install?.install_components?.reduce((acc, installComponent) => {
      acc.push([
        installComponent.id,
        installComponent.component.name,
        installComponent.install_deploys?.[0]?.status,
        installComponent.install_deploys?.[0]?.status,
        installComponent.component.config_versions || 0,
        `/beta/${orgId}/installs/${installId}/components/${installComponent.id}`,
      ])

      return acc
    }, []) || []

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/beta/${org.id}`, text: org.name },
        { href: `/beta/${org.id}/installs`, text: 'Installs' },
        { href: `/beta/${org.id}/installs/${install.id}`, text: install.name },
      ]}
      heading={install.name}
      headingUnderline={install.id}
      statues={
        <div>
          <InstallProvider initInstall={install}>
            <InstallStatus />
          </InstallProvider>
        </div>
      }
      meta={<SubNav links={subNavLinks} />}
    >
      <section className="px-6 py-8">
        <DataTable
          headers={['ID', 'Name', 'Deployment', 'Build', 'Config']}
          initData={tableData}
        />
      </section>
    </DashboardContent>
  )
}
