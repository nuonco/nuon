import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  AppPageSubNav,
  AppWorkflowsTable,
  DashboardContent,
  NoActions,
} from '@/components'
import { getApp, getAppWorkflows, getOrg } from '@/lib'

export default withPageAuthRequired(async function AppWorkflows({ params }) {
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
  const [org, app, workflows] = await Promise.all([
    getOrg({ orgId }),
    getApp({ appId, orgId }),
    getAppWorkflows({ appId, orgId }),
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
          <NoActions />
        )}
      </section>
    </DashboardContent>
  )
})
