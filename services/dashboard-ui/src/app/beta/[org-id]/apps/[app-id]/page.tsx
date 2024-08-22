import {
  AppInputConfig,
  AppRunnerConfig,
  AppSandboxConfig,
  AppSandboxVariables,
  Heading,
} from '@/components'
import {
  getApp,
  getAppInputLatestConfig,
  getAppRunnerLatestConfig,
  getAppSandboxLatestConfig,
  getOrg,
} from '@/lib'

export default async function App({ params }) {
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
  const [org, app, inputCfg, runnerCfg, sandboxCfg] = await Promise.all([
    getOrg({ orgId }),
    getApp({ appId, orgId }),
    getAppInputLatestConfig({ appId, orgId }),
    getAppRunnerLatestConfig({ appId, orgId }),
    getAppSandboxLatestConfig({ appId, orgId }),
  ])

  return (
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
  )
}
