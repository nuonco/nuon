import { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  AppCreateInstallButton,
  AppInputConfig,
  AppPageSubNav,
  AppRunnerConfig,
  AppSandboxConfig,
  AppSandboxVariables,
  DashboardContent,
  ErrorFallback,
  Loading,
  Section,
  Markdown,
} from '@/components'
import {
  getApp,
  getAppLatestConfig,
  getAppLatestInputConfig,
  getAppLatestRunnerConfig,
  getAppLatestSandboxConfig,
  getOrg,
  type IGetApp,
} from '@/lib'

export default withPageAuthRequired(async function App({ params }) {
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
  const [org, app, appConfig, inputCfg] = await Promise.all([
    getOrg({ orgId }),
    getApp({ appId, orgId }),
    getAppLatestConfig({ appId, orgId }).catch(console.error),
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
      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto">
        <div className="divide-y flex flex-col md:col-span-7">
          {appConfig ? (
            <Section className="border-r" heading="README">
              <Markdown content={appConfig.readme} />
            </Section>
          ) : null}

          {inputCfg ? (
            <Section className="border-r" heading="Inputs">
              <AppInputConfig inputConfig={inputCfg} />
            </Section>
          ) : null}
        </div>

        <div className="divide-y flex flex-col md:col-span-5">
          <Section className="flex-initial" heading="Sandbox">
            <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={
                  <Loading loadingText="Loading latest sandbox config..." />
                }
              >
                <LoadAppSandboxConfig appId={appId} orgId={orgId} />
              </Suspense>
            </ErrorBoundary>
          </Section>

          <Section heading="Runner">
            <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={
                  <Loading loadingText="Loading latest runner config..." />
                }
              >
                <LoadAppRunnerConfig appId={appId} orgId={orgId} />
              </Suspense>
            </ErrorBoundary>
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
})

const LoadAppSandboxConfig: FC<IGetApp> = async (props) => {
  const sandboxConfig = await getAppLatestSandboxConfig(props)
  return (
    <div className="flex flex-col gap-8">
      <AppSandboxConfig sandboxConfig={sandboxConfig} />
      <AppSandboxVariables variables={sandboxConfig?.variables} />
    </div>
  )
}

const LoadAppRunnerConfig: FC<{ appId: string; orgId: string }> = async ({
  appId,
  orgId,
}) => {
  const runnerConfig = await getAppLatestRunnerConfig({ appId, orgId })
  return <AppRunnerConfig runnerConfig={runnerConfig} />
}
