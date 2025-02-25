import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  AppCreateInstallButton,
  AppPageSubNav,
  AppWorkflowsTable,
  DashboardContent,
  NoActions,
} from '@/components'
import { getApp, getAppActionWorkflows, getAppLatestInputConfig } from '@/lib'

export default withPageAuthRequired(async function AppWorkflows({ params }) {
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
  const [app, workflows, inputCfg] = await Promise.all([
    getApp({ appId, orgId }),
    getAppActionWorkflows({ appId, orgId }).catch(console.error),
    getAppLatestInputConfig({ appId, orgId }).catch(console.error),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/apps`, text: 'Apps' },
        { href: `/${orgId}/apps/${app.id}`, text: app.name },
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
        {workflows && workflows?.length ? (
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
