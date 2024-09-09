import { FaAws } from 'react-icons/fa'
import { VscAzure } from 'react-icons/vsc'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { DashboardContent, DataTable, Heading, Text } from '@/components'
import { getApps, getOrg } from '@/lib'

export default withPageAuthRequired(
  async function Apps({ params }) {
    const orgId = params?.['org-id'] as string
    const [apps, org] = await Promise.all([
      getApps({ orgId }),
      getOrg({ orgId }),
    ])

    const tableData = apps.reduce((acc, app) => {
      /* eslint react/jsx-key: 0 */
      acc.push([
        <div className="flex flex-col gap-2">
          <Heading variant="subheading">{app.name}</Heading>
          <Text variant="overline">{app.id}</Text>
        </div>,
        <Text className="flex items-center gap-2" variant="caption">
          {app.cloud_platform === 'azure' ? (
            <>
              <VscAzure className="text-md" /> {'Azure'}
            </>
          ) : (
            <>
              <FaAws className="text-xl mb-[-4px]" /> {'Amazon'}
            </>
          )}
        </Text>,
        <Text variant="caption">{app.sandbox_config?.terraform_version}</Text>,
        <Text variant="caption">{app?.runner_config?.app_runner_type}</Text>,
        `/beta/${orgId}/apps/${app.id}`,
      ])
      /* eslint react/jsx-key: 1 */
      return acc
    }, [])

    return (
      <DashboardContent
        breadcrumb={[
          { href: `/beta/${org.id}`, text: org.name },
          { href: `/beta/${org.id}/apps`, text: 'Apps' },
        ]}
      >
        <section className="px-6 py-8">
          <DataTable
            headers={['Name', 'Platform', 'Sandbox', 'Runner']}
            initData={tableData}
          />
        </section>
      </DashboardContent>
    )
  },
  { returnTo: '/dashboard' }
)
