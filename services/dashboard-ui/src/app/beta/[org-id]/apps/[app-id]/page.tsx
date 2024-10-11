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
import type {
  TAppInputConfig,
  TAppRunnerConfig,
  TAppSandboxConfig,
} from '@/types'

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
      getAppInputLatestConfig({ appId, orgId }).catch(console.error),
      getAppRunnerLatestConfig({ appId, orgId }).catch(console.error),
      getAppSandboxLatestConfig({ appId, orgId }).catch(console.error),
    ])

    return (
      <DashboardContent
        breadcrumb={[
          { href: `/beta/${org.id}/apps`, text: org.name },
          { href: `/beta/${org.id}/apps`, text: 'Apps' },
          { href: `/beta/${org.id}/apps/${app.id}`, text: app.name },
        ]}
        heading={app.name}
        headingUnderline={app.id}
        meta={<SubNav links={subNavLinks} />}
      >
        <div className="flex flex-col md:flex-row flex-auto">
          <section className="flex flex-col gap-4 px-6 py-8 border-r w-full">
            <Heading variant="subheading">Inputs</Heading>
            <AppInputConfig inputConfig={inputCfg as TAppInputConfig} />
          </section>

          <div className="flex flex-col lg:min-w-[510px]">
            <section className="flex flex-col gap-4 px-6 py-8 border-b">
              <Heading variant="subheading">Sandbox</Heading>
              <AppSandboxConfig
                sandboxConfig={sandboxCfg as TAppSandboxConfig}
              />
              <AppSandboxVariables
                variables={(sandboxCfg as TAppSandboxConfig)?.variables}
              />
            </section>

            <section className="flex flex-col gap-4 px-6 py-8">
              <Heading variant="subheading">Runner</Heading>

              <AppRunnerConfig runnerConfig={runnerCfg as TAppRunnerConfig} />
            </section>
          </div>
        </div>
      </DashboardContent>
    )
  },
  { returnTo: '/dashboard' }
)
