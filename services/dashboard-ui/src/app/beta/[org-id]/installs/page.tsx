import { FaAws } from 'react-icons/fa'
import { VscAzure } from 'react-icons/vsc'
import {
  DashboardContent,
  DataTable,
  Heading,
  Status,
  Text,
} from '@/components'
import { getOrg, getInstalls } from '@/lib'

export default async function Installs({ params }) {
  const orgId = params?.['org-id'] as string
  const [installs, org] = await Promise.all([
    getInstalls({ orgId }),
    getOrg({ orgId }),
  ])

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
      <Text variant="caption">{install?.app?.name}</Text>,
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
        { href: `/beta/${org.id}/installs`, text: 'Installs' },
      ]}
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
