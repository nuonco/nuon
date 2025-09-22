import type { Metadata } from 'next'
import { Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import {
  AppCreateInstallButton,
  AppPageSubNav,
  DashboardContent,
  ErrorFallback,
  Loading,
  Section,
} from '@/components'
import { getAppById, getAppConfigs } from '@/lib'
import { InputsConfig } from './inputs-config'
import { ReadmeConfig } from './readme-config'
import { RunnerConfig } from './runner-config'
import { SandboxConfig } from './sandbox-config'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const { data: app } = await getAppById({ appId, orgId })

  return {
    title: `Configuration | ${app.name} | Nuon`,
  }
}

export default async function AppConfigPage({ params }) {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const [{ data: app }, { data: configs }] = await Promise.all([
    getAppById({ appId, orgId }),
    getAppConfigs({ appId, orgId, limit: 1 }),
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
        configs?.length ? (
          <AppCreateInstallButton
            platform={app?.runner_config.app_runner_type}
          />
        ) : null
      }
      meta={<AppPageSubNav appId={appId} orgId={orgId} />}
    >
      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto">
        <div className="divide-y flex flex-col md:col-span-7">
          <ErrorBoundary fallbackRender={ErrorFallback}>
            <Suspense
              fallback={
                <Section className="border-r" heading="README">
                  <Loading
                    loadingText="Loading latest README config..."
                    variant="stack"
                  />
                </Section>
              }
            >
              <ReadmeConfig
                appConfigId={configs?.at(0)?.id}
                appId={appId}
                orgId={orgId}
              />
            </Suspense>
          </ErrorBoundary>
        </div>

        <div className="divide-y flex flex-col md:col-span-5">
          <ErrorBoundary fallbackRender={ErrorFallback}>
            <Suspense
              fallback={
                <Section className="flex-initial" heading="Inputs">
                  <Loading loadingText="Loading latest input config..." />
                </Section>
              }
            >
              <InputsConfig
                appConfigId={configs?.at(0)?.id}
                appId={appId}
                appName={app?.name}
                orgId={orgId}
              />
            </Suspense>
          </ErrorBoundary>

          <Section className="flex-initial" heading="Sandbox">
            <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={
                  <Loading loadingText="Loading latest sandbox config..." />
                }
              >
                <SandboxConfig
                  appConfigId={configs?.at(0)?.id}
                  appId={appId}
                  orgId={orgId}
                />
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
                <RunnerConfig
                  appConfigId={configs?.at(0)?.id}
                  appId={appId}
                  orgId={orgId}
                />
              </Suspense>
            </ErrorBoundary>
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
}
