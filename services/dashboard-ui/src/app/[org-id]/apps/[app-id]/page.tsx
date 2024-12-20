import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  AppInputConfig,
  AppPageSubNav,
  AppRunnerConfig,
  AppSandboxConfig,
  AppSandboxVariables,
  DashboardContent,
  Section,
  Markdown,
} from '@/components'
import {
  getApp,
  getAppLatestConfig,
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

export default withPageAuthRequired(async function App({ params }) {
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
  const [org, app, appConfig, inputCfg, runnerCfg, sandboxCfg] =
    await Promise.all([
      getOrg({ orgId }),
      getApp({ appId, orgId }),
      getAppLatestConfig({ appId, orgId }),
      getAppInputLatestConfig({ appId, orgId }).catch(console.error),
      getAppRunnerLatestConfig({ appId, orgId }).catch(console.error),
      getAppSandboxLatestConfig({ appId, orgId }).catch(console.error),
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
      <div className="flex flex-col md:flex-row flex-auto">
        <div className="divide-y flex flex-col flex-grow">
          <Section className="border-r" heading="README">
            <Markdown content={appConfig.readme} />
          </Section>

          <Section className="border-r" heading="Inputs">
            <AppInputConfig inputConfig={inputCfg as TAppInputConfig} />
          </Section>
        </div>

        <div className="divide-y flex flex-col lg:min-w-[450px] lg:max-w-[450px]">
          <Section className="flex-initial" heading="Sandbox">
            <div className="flex flex-col gap-8">
              <AppSandboxConfig
                sandboxConfig={sandboxCfg as TAppSandboxConfig}
              />
              <AppSandboxVariables
                variables={(sandboxCfg as TAppSandboxConfig)?.variables}
              />
            </div>
          </Section>

          <Section heading="Runner">
            <AppRunnerConfig runnerConfig={runnerCfg as TAppRunnerConfig} />
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
})
