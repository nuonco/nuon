import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  AppCreateInstallButton,
  AppPageSubNav,
  AppWorkflowsTable,
  DashboardContent,
  NoActions,
} from '@/components'
import {
  getApp,
  getAppActionWorkflows,
  getAppLatestInputConfig,
  getOrg,
} from '@/lib'

export default withPageAuthRequired(async function AppWorkflows({ params }) {
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
  const [org, app, workflows, inputCfg] = await Promise.all([
    getOrg({ orgId }),
    getApp({ appId, orgId }),
    getAppActionWorkflows({ appId, orgId }),
    getAppLatestInputConfig({ appId, orgId }).catch(console.error),
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
      statues={
        inputCfg ? (
          <AppCreateInstallButton
            platform={app?.cloud_platform}
            inputConfig={inputCfg}
            appId={appId}
            orgId={orgId}
          />
        ) : null
      }
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
