import { FaAws } from 'react-icons/fa'
import { VscAzure } from 'react-icons/vsc'
import {
  DashboardContent,
  DataTable,
  Heading,
  Status,
  SubNav,
  Text,
  type TLink,
} from '@/components'
import { getApp, getAppInstalls, getOrg } from '@/lib'

export default async function AppInstalls({ params }) {
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
  const subNavLinks: Array<TLink> = [
    { href: `/beta/${orgId}/apps/${appId}`, text: 'Config' },
    { href: `/beta/${orgId}/apps/${appId}/components`, text: 'Components' },
    { href: `/beta/${orgId}/apps/${appId}/installs`, text: 'Installs' },
  ]

  const app = await getApp({ appId, orgId })
  const installs = await getAppInstalls({ appId, orgId })
  const org = await getOrg({ orgId })

  const tableData = installs.reduce((acc, install) => {
    /* eslint react/jsx-key: 0 */
    acc.push([
      <div className="flex flex-col gap-2">
        <Heading variant="subheading">{install.name}</Heading>
        <Text variant="overline">{install.id}</Text>
      </div>,
      <div className="flex flex-col gap-2">
        <Status
          status={install.sandbox_status}
          label="Sandbox"
          isLabelStatusText
        />
        <Status
          status={install.runner_status}
          label="Runner"
          isLabelStatusText
        />
        <Status
          status={install.composite_component_status}
          label="Components"
          isLabelStatusText
        />
      </div>,
      <Text variant="caption">{app?.name}</Text>,
      <Text className="flex items-center gap-2" variant="caption">
        {install.app_sandbox_config.cloud_platform === 'azure' ? (
          <>
            <VscAzure className="text-md" /> {'Azure'}
          </>
        ) : (
          <>
            <FaAws className="text-xl mb-[-4px]" /> {'Amazon'}
          </>
        )}
      </Text>,
      `/beta/${orgId}/installs/${install.id}`,
    ])
    /* eslint react/jsx-key: 1 */

    return acc
  }, [])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/beta/${org.id}`, text: org.name },
        { href: `/beta/${org.id}/apps`, text: 'Apps' },
        { href: `/beta/${org.id}/apps/${app.id}`, text: app.name },
      ]}
      heading={app.name}
      headingUnderline={app.id}
      meta={<SubNav links={subNavLinks} />}
    >
      <section className="px-6 py-8">
        <DataTable
          headers={['Name', 'Statues', 'App', 'Platform']}
          initData={tableData}
        />
      </section>
    </DashboardContent>
  )
}
