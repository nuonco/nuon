import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  AppCreateInstallButton,
  AppInstallsTable,
  AppPageSubNav,
  DashboardContent,
  NoInstalls,
} from '@/components'
import { getApp, getAppInstalls, getOrg, getAppLatestInputConfig } from '@/lib'

export default withPageAuthRequired(async function AppInstalls({ params }) {
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
  const [app, installs, org, inputCfg] = await Promise.all([
    getApp({ appId, orgId }),
    getAppInstalls({ appId, orgId }),
    getOrg({ orgId }),
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
