import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { DashboardContent, OrgAppsTable } from '@/components'
import { getApps, getOrg } from '@/lib'

export default withPageAuthRequired(
  async function Apps({ params }) {
    const orgId = params?.['org-id'] as string
    const [apps, org] = await Promise.all([
      getApps({ orgId }),
      getOrg({ orgId }),
    ])

    return (
      <DashboardContent
        breadcrumb={[
          { href: `/beta/${org.id}`, text: org.name },
          { href: `/beta/${org.id}/apps`, text: 'Apps' },
        ]}
      >
        <section className="px-6 py-8">
          <OrgAppsTable apps={apps} orgId={orgId} />
        </section>
      </DashboardContent>
    )
  },
  { returnTo: '/dashboard' }
)
