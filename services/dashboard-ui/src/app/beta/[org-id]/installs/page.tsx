import { DashboardContent, OrgInstallsTable } from '@/components'
import { getOrg, getInstalls } from '@/lib'

export default async function Installs({ params }) {
  const orgId = params?.['org-id'] as string
  const [installs, org] = await Promise.all([
    getInstalls({ orgId }),
    getOrg({ orgId }),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/beta/${org.id}`, text: org.name },
        { href: `/beta/${org.id}/installs`, text: 'Installs' },
      ]}
    >
      <section className="px-6 py-8">
        <OrgInstallsTable orgId={orgId} installs={installs} />
      </section>
    </DashboardContent>
  )
}
