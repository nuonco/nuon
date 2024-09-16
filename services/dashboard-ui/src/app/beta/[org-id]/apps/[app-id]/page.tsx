import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  AppInputConfig,
  AppRunnerConfig,
  AppSandboxConfig,
  AppSandboxVariables,
  DashboardContent,
  Heading,
  SubNav,
  type TLink,
} from '@/components'
import {
  getApp,
  getAppInputLatestConfig,
  getAppRunnerLatestConfig,
  getAppSandboxLatestConfig,
  getOrg,
} from '@/lib'

export default withPageAuthRequired(
  async function App({ params }) {
    const appId = params?.['app-id'] as string
    const orgId = params?.['org-id'] as string
    const subNavLinks: Array<TLink> = [
      { href: `/beta/${orgId}/apps/${appId}`, text: 'Config' },
      { href: `/beta/${orgId}/apps/${appId}/components`, text: 'Components' },
      { href: `/beta/${orgId}/apps/${appId}/installs`, text: 'Installs' },
    ]

    const [org, app, inputCfg, runnerCfg, sandboxCfg] = await Promise.all([
      getOrg({ orgId }),
      getApp({ appId, orgId }),
      getAppInputLatestConfig({ appId, orgId }),
      getAppRunnerLatestConfig({ appId, orgId }),
      getAppSandboxLatestConfig({ appId, orgId }),
    ])

    return (
      <DashboardContent
        breadcrumb={[
          { href: `/beta/${org.id}`, text: org.name },
          { href: `/beta/${org.id}/apps`, text: 'Apps' },
          { href: `/beta/${org.id}/apps/${app.id}`, text: app.name },
        ]}
        heading={app.name}
        headingUnderline={app.id}
        meta={<SubNav links={subNavLinks} />}
      >
        <div className="flex flex-col md:flex-row flex-auto">
          <section className="flex flex-col gap-4 px-6 py-8 border-r">
            <Heading variant="subheading">Inputs</Heading>
            <AppInputConfig inputConfig={inputCfg} />
          </section>

          <div className="flex flex-col lg:min-w-[510px]">
            <section className="flex flex-col gap-4 px-6 py-8 border-b">
              <Heading variant="subheading">Sandbox</Heading>
              <AppSandboxConfig sandboxConfig={sandboxCfg} />
              <AppSandboxVariables variables={sandboxCfg.variables} />
            </section>

            <section className="flex flex-col gap-4 px-6 py-8">
              <Heading variant="subheading">Runner</Heading>

              <AppRunnerConfig runnerConfig={runnerCfg} />
            </section>
          </div>
        </div>
      </DashboardContent>
    )
  },
  { returnTo: '/dashboard' }
)
