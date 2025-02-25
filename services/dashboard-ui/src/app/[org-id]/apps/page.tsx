import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { DashboardContent, NoApps, OrgAppsTable } from '@/components'
import { getApps, getOrg } from '@/lib'
// TODO(nnnat): move segment init script to org dashboard
import { SegmentAnalyticsSetOrg } from '@/utils'

export default withPageAuthRequired(async function Apps({ params }) {
  const orgId = params?.['org-id'] as string
  const [apps, org] = await Promise.all([getApps({ orgId }), getOrg({ orgId })])

  return (
    <>
      {process.env.SEGMENT_WRITE_KEY && <SegmentAnalyticsSetOrg org={org} />}
      <DashboardContent
        breadcrumb={[
          { href: `/${orgId}/apps`, text: 'Apps' },
        ]}
      >
        <section className="px-6 py-8">
          {apps.length ? (
            <OrgAppsTable apps={apps} orgId={orgId} />
          ) : (
            <NoApps />
          )}
        </section>
      </DashboardContent>
    </>
  )
})
