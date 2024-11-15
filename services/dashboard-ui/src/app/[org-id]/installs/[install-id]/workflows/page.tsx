import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  InstallPageSubNav,
  InstallWorkflowsTable,
  DashboardContent,
  Text,
} from '@/components'
import { getInstall, getOrg } from '@/lib'
import type { TWorkflow } from '@/types'

const workflows: Array<TWorkflow> = [
  {
    id: 'wkf12345678912345',
    name: 'Fetch logs',
    on: 'manual',
    jobs: [{ id: 'j-1' }],
  },
  {
    id: 'wkf09876543210987',
    name: 'Health check',
    on: 'schedule',
    jobs: [{ id: 'j-1' }, { id: 'j-1' }, { id: 'j-1' }, { id: 'j-1' }],
  },
]

export default withPageAuthRequired(async function InstallWorkflowRuns({
  params,
}) {
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string
  const [org, install] = await Promise.all([
    getOrg({ orgId }),
    getInstall({ installId, orgId }),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${org.id}/apps`, text: org.name },
        { href: `/${org.id}/installs`, text: 'Installs' },
        { href: `/${org.id}/installs/${install.id}`, text: install.name },
      ]}
      heading={install.name}
      headingUnderline={install.id}
      meta={<InstallPageSubNav installId={installId} orgId={orgId} />}
    >
      <section className="px-6 py-8">
        {workflows.length ? (
          <InstallWorkflowsTable
            installId={installId}
            orgId={orgId}
            workflows={workflows}
          />
        ) : (
          <Text>No workflow runs</Text>
        )}
      </section>
    </DashboardContent>
  )
})
