import type { Metadata } from 'next'
import { Suspense } from 'react'
import { AsyncBoundary } from '@/components/common/AsyncBoundary'
import { BackToTop } from '@/components/common/BackToTop'
import { Card } from '@/components/common/Card'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { getApp, getAppConfigs, getOrg } from '@/lib'
import type { TPageProps } from '@/types'
import { AppInputs, AppInputsError, AppInputsSkeleton } from './inputs-config'
import {
  AppRunner,
  AppRunnerError,
  AppRunnerSkeleton,
} from './runner-config'
import {
  AppSandbox,
  AppSandboxError,
  AppSandboxSkeleton,
} from './sandbox-config'

// NOTE: old layout stuff
import { ErrorBoundary } from 'react-error-boundary'
import {
  AppCreateInstallButton,
  AppPageSubNav,
  DashboardContent,
  ErrorFallback,
  Loading,
  Section,
} from '@/components'
import {
  InputsConfig,
  ReadmeConfig,
  RunnerConfig,
  SandboxConfig,
} from './app-configs.old'

type TAppPageProps = TPageProps<'org-id' | 'app-id'>

export async function generateMetadata({
  params,
}: TAppPageProps): Promise<Metadata> {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const { data: app } = await getApp({ appId, orgId })

  return {
    title: `Configuration | ${app.name} | Nuon`,
  }
}

export default async function AppOverviewPage({ params }: TAppPageProps) {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const [{ data: app }, { data: configs }, { data: org }] = await Promise.all([
    getApp({ appId, orgId }),
    getAppConfigs({ appId, orgId, limit: 1 }),
    getOrg({ orgId }),
  ])

  const containerId = 'app-overview-page'
  return org?.features?.['stratus-layout'] ? (
    <>
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${orgId}`,
            text: org?.name,
          },
          {
            path: `/${orgId}/apps`,
            text: 'Apps',
          },
          {
            path: `/${orgId}/apps/${appId}`,
            text: app?.name,
          },
        ]}
      />
      <PageSection id={containerId} isScrollable>
        <AsyncBoundary
          errorFallback={<AppInputsError />}
          loadingFallback={<AppInputsSkeleton />}
        >
          <AppInputs
            appConfigId={configs?.at(0)?.id}
            appId={appId}
            orgId={orgId}
          />
        </AsyncBoundary>
        {/* old page stuff */}

        <div className="flex gap-6">
          <AsyncBoundary
            errorFallback={<AppSandboxError />}
            loadingFallback={<AppSandboxSkeleton />}
          >
            <AppSandbox
              appConfigId={configs?.at(0)?.id}
              appId={appId}
              orgId={orgId}
            />
          </AsyncBoundary>

          <AsyncBoundary
            errorFallback={<AppRunnerError />}
            loadingFallback={<AppRunnerSkeleton />}
          >
            <AppRunner
              appConfigId={configs?.at(0)?.id}
              appId={appId}
              orgId={orgId}
            />
          </AsyncBoundary>
        </div>

        {/* old page stuff */}
        <BackToTop containerId={containerId} />
      </PageSection>
    </>
  ) : (
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
