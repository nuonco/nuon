import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { DashboardContent, OrgAppsTable } from '@/components'
import { getApps, getOrg } from '@/lib'
import { SegmentAnalyticsSetOrg } from '@/utils'

export default withPageAuthRequired(
  async function Apps({ params }) {
    const orgId = params?.['org-id'] as string
    const [apps, org] = await Promise.all([
      getApps({ orgId }),
      getOrg({ orgId }),
    ])

    return (
      <>
        {process.env.SEGMENT_WRITE_KEY && <SegmentAnalyticsSetOrg org={org} />}
        <DashboardContent
          breadcrumb={[
            { href: `/${org.id}/apps`, text: org.name },
            { href: `/${org.id}/apps`, text: 'Apps' },
          ]}
        >
          <section className="px-6 py-8">
            <OrgAppsTable apps={apps} orgId={orgId} />
          </section>
        </DashboardContent>
      </>
    )
  },
  { returnTo: '/' }
)
