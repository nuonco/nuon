import type { Metadata } from 'next'
import { AsyncBoundary } from '@/components/common/AsyncBoundary'
import { BackToTop } from '@/components/common/BackToTop'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { getApp, getAppConfigs, getOrg } from '@/lib'
import type { TPageProps } from '@/types'
import { AppInputs, AppInputsError, AppInputsSkeleton } from './inputs-config'
import { AppRunner, AppRunnerError, AppRunnerSkeleton } from './runner-config'
import {
  AppSandbox,
  AppSandboxError,
  AppSandboxSkeleton,
} from './sandbox-config'
import { AppStack, AppStackError, AppStackSkeleton } from './stack-config'

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
  return (
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

        <div className="@container">
          <div className="grid grid-cols-1 @lg:grid-cols-5 gap-6">
            <div className="@lg:col-span-2">
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
            <div className="@lg:col-span-3">
              <AsyncBoundary
                errorFallback={<AppStackError />}
                loadingFallback={<AppStackSkeleton />}
              >
                <AppStack
                  appConfigId={configs?.at(0)?.id}
                  appId={appId}
                  orgId={orgId}
                />
              </AsyncBoundary>
            </div>
          </div>
        </div>

        {/* old page stuff */}
        <BackToTop containerId={containerId} />
      </PageSection>
    </>
  )
}
