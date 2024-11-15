import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  AppPageSubNav,
  AppWorkflowsTable,
  DashboardContent,
  Text,
} from '@/components'
import { getApp, getOrg } from '@/lib'
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

export default withPageAuthRequired(async function AppWorkflows({ params }) {
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
  const [org, app] = await Promise.all([
    getOrg({ orgId }),
    getApp({ appId, orgId }),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${org.id}/apps`, text: org.name },
        { href: `/${org.id}/apps`, text: 'Apps' },
        { href: `/${org.id}/apps/${app.id}`, text: app.name },
      ]}
      heading={app.name}
      headingUnderline={app.id}
      meta={<AppPageSubNav appId={appId} orgId={orgId} />}
    >
      <section className="px-6 py-8">
        {workflows.length ? (
          <AppWorkflowsTable
            appId={appId}
            orgId={orgId}
            workflows={workflows}
          />
        ) : (
          <Text>No workflows configured</Text>
        )}
      </section>
    </DashboardContent>
  )
})
