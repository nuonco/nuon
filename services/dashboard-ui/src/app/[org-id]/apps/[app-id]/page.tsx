import type { Metadata } from 'next'
import { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import {
  AppConfigGraph,
  AppCreateInstallButton,
  AppInputConfig,
  AppInputConfigModal,
  AppPageSubNav,
  AppRunnerConfig,
  AppSandboxConfig,
  AppSandboxVariables,
  DashboardContent,
  EmptyStateGraphic,
  ErrorFallback,
  Link,
  Loading,
  Section,
  Text,
  Markdown,
} from '@/components'
import {
  getApp,
  getAppLatestConfig,
  getAppLatestInputConfig,
  getAppLatestRunnerConfig,
  getAppLatestSandboxConfig,
} from '@/lib'
import type { IGetApp } from '@/lib/ctl-api/shared-interfaces'
import type { TApp } from '@/types'
import { nueQueryData } from '@/utils'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const app = await getApp({ appId, orgId })

  return {
    title: `${app.name} | Config`,
  }
}

export default async function App({ params }) {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const [app, appConfig, inputCfg] = await Promise.all([
    getApp({ appId, orgId }),
    getAppLatestConfig({ appId, orgId }).catch(console.error),
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
            platform={app?.runner_config.app_runner_type}
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
            <Section className="border-r overflow-x-auto" heading="README">
              <Markdown content={appConfig.readme} />
            </Section>
          ) : (
            <Section className="border-r overflow-x-auto" heading="README">
              <div className="m-auto flex flex-col items-center max-w-[200px] my-6">
                <EmptyStateGraphic variant="table" />
                <Text className="mt-6" variant="med-14">
                  No README in app config
                </Text>
                <Text variant="reg-12" className="text-center !inline-block">
                  You can add a README for your app in your app config TOML
                  file.
                </Text>
              </div>
            </Section>
          )}
        </div>

        <div className="divide-y flex flex-col md:col-span-5">
          {inputCfg && inputCfg?.input_groups?.length ? (
            <Section
              className="flex-initial"
              heading="Inputs"
              actions={
                <AppInputConfigModal
                  inputConfig={inputCfg}
                  appName={app?.name}
                />
              }
            >
              <AppInputConfig inputConfig={inputCfg} />
            </Section>
          ) : null}

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
}

const LoadAppSandboxConfig: FC<IGetApp> = async (props) => {
  const sandboxConfig = await getAppLatestSandboxConfig(props).catch(
    console.error
  )
  return sandboxConfig ? (
    <div className="flex flex-col gap-8">
      <AppSandboxConfig sandboxConfig={sandboxConfig} />
      <AppSandboxVariables variables={sandboxConfig?.variables} />
    </div>
  ) : (
    <div className="m-auto flex flex-col items-center max-w-[200px] my-6">
      <EmptyStateGraphic variant="table" />
      <Text className="mt-6" variant="med-14">
        No app sandbox config
      </Text>
      <Text variant="reg-12" className="text-center !inline-block">
        Read more about app sandbox configs{' '}
        <Link
          className="!inline-block"
          href="https://docs.nuon.co/concepts/sandboxes"
          target="_blank"
        >
          here
        </Link>
        .
      </Text>
    </div>
  )
}

const LoadAppRunnerConfig: FC<{ appId: string; orgId: string }> = async ({
  appId,
  orgId,
}) => {
  const runnerConfig = await getAppLatestRunnerConfig({ appId, orgId }).catch(
    console.error
  )
  return runnerConfig ? (
    <AppRunnerConfig runnerConfig={runnerConfig} />
  ) : (
    <div className="m-auto flex flex-col items-center max-w-[200px] my-6">
      <EmptyStateGraphic variant="table" />
      <Text className="mt-6" variant="med-14">
        No app runner config
      </Text>
      <Text variant="reg-12" className="text-center !inline-block">
        Read more about app runner configs{' '}
        <Link
          className="!inline-block"
          href="https://docs.nuon.co/concepts/runners"
          target="_blank"
        >
          here
        </Link>
        .
      </Text>
    </div>
  )
}

const LoadAppConfigGraph: FC<{ app: TApp; configId: string }> = async ({
  app,
  configId,
}) => {
  return <AppConfigGraph appId={app?.id} configId={configId} />
}
