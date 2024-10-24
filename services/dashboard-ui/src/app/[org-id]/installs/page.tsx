import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { DashboardContent, NoInstalls, OrgInstallsTable } from '@/components'
import { getOrg, getInstalls } from '@/lib'

export default withPageAuthRequired(
  async function Installs({ params }) {
    const orgId = params?.['org-id'] as string
    const [installs, org] = await Promise.all([
      getInstalls({ orgId }),
      getOrg({ orgId }),
    ])

    return (
      <DashboardContent
        breadcrumb={[
          { href: `/${org.id}/apps`, text: org.name },
          { href: `/${org.id}/installs`, text: 'Installs' },
        ]}
      >
        <section className="px-6 py-8">
          {installs.length ? <OrgInstallsTable orgId={orgId} installs={installs} /> : <NoInstalls />}
        </section>
      </DashboardContent>
    )
  },
  { returnTo: '/' }
)
