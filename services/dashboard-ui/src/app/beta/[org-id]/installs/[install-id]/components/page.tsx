import {
  ComponentConfigType,
  DashboardContent,
  DataTable,
  InstallStatus,
  Heading,
  Status,
  SubNav,
  Text,
  Time,
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
      /* eslint react/jsx-key: 0 */
      acc.push([
        <div className="flex flex-col gap-2">
          <Heading variant="subheading">
            {installComponent?.component?.name}
          </Heading>
          <Text variant="caption">{installComponent.id}</Text>
        </div>,

        <Text variant="caption">
          <ComponentConfigType
            componentId={installComponent?.component_id}
            orgId={orgId}
          />
        </Text>,

        <Time
          time={installComponent.install_deploys?.[0].updated_at}
          format="relative"
          variant="caption"
        />,

        <Text variant="caption">TKTK</Text>,

        <Status status={installComponent.install_deploys?.[0]?.status} />,

        <Text variant="caption">
          {installComponent.install_deploys?.[0]?.component_config_version || 0}
        </Text>,
        `/beta/${orgId}/installs/${installId}/components/${installComponent.id}`,
      ])
      /* eslint react/jsx-key: 1 */
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
          headers={[
            'Name',
            'Type',
            'Deployment',
            'Dependencies',
            'Build',
            'Config',
          ]}
          initData={tableData}
        />
      </section>
    </DashboardContent>
  )
}
