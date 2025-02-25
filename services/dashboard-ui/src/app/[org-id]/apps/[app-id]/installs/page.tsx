import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  AppCreateInstallButton,
  AppInstallsTable,
  AppPageSubNav,
  DashboardContent,
  NoInstalls,
} from '@/components'
import { getApp, getAppInstalls, getAppLatestInputConfig } from '@/lib'

export default withPageAuthRequired(async function AppInstalls({ params }) {
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
  const [app, installs, inputCfg] = await Promise.all([
    getApp({ appId, orgId }),
    getAppInstalls({ appId, orgId }),
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
        {installs.length ? (
          <AppInstallsTable
            installs={installs.map((install) => ({ ...install, app }))}
            orgId={orgId}
          />
        ) : (
          <NoInstalls />
        )}
      </section>
    </DashboardContent>
  )
})
