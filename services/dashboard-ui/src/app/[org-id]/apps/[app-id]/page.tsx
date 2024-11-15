import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  AppInputConfig,
  AppPageSubNav,
  AppRunnerConfig,
  AppSandboxConfig,
  AppSandboxVariables,
  DashboardContent,
  Section,
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

export default withPageAuthRequired(async function App({ params }) {
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
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
        { href: `/${org.id}/apps`, text: org.name },
        { href: `/${org.id}/apps`, text: 'Apps' },
        { href: `/${org.id}/apps/${app.id}`, text: app.name },
      ]}
      heading={app.name}
      headingUnderline={app.id}
      meta={<AppPageSubNav appId={appId} orgId={orgId} />}
    >
      <div className="flex flex-col md:flex-row flex-auto">
        <Section className="border-r" heading="Inputs">
          <AppInputConfig inputConfig={inputCfg as TAppInputConfig} />
        </Section>

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
