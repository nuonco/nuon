import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { AsyncBoundary } from '@/components/common/AsyncBoundary'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { RunnerRecentActivitySkeleton } from '@/components/runners/RunnerRecentActivitySkeleton'
import { RunnerDetailsCardSkeleton } from '@/components/runners/RunnerDetailsCardSkeleton'
import { RunnerHealthCardSkeleton } from '@/components/runners/RunnerHealthCardSkeleton'
import { ManagementDropdown } from '@/components/runners/management/ManagementDropdown'
import {
  getInstall,
  getRunner,
  getRunnerSettings,
  getRunnerLatestHeartbeat,
  getOrg,
} from '@/lib'
import { RunnerProvider } from '@/providers/runner-provider'
import { TPageProps } from '@/types'
import { RunnerActivity, RunnerActivityError } from './runner-activity'
import { RunnerDetails, RunnerDetailsError } from './runner-details'
import { RunnerHealth, RunnerHealthError } from './runner-health'

type TInstallPageProps = TPageProps<'org-id' | 'install-id'>

export async function generateMetadata({
  params,
}: TInstallPageProps): Promise<Metadata> {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const { data: install } = await getInstall({ installId, orgId })

  return {
    title: `Runner | ${install?.name} | Nuon`,
  }
}

export default async function Runner({
  params,
  searchParams,
}: TInstallPageProps) {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const sp = await searchParams
  const [{ data: install }, { data: org }] = await Promise.all([
    getInstall({ installId, orgId }),
    getOrg({ orgId }),
  ])
  const [{ data: runner, error }, { data: settings }, { data: heartbeat }] =
    await Promise.all([
      getRunner({
        orgId,
        runnerId: install.runner_id,
      }),
      getRunnerSettings({
        orgId,
        runnerId: install.runner_id,
      }),
      getRunnerLatestHeartbeat({ orgId, runnerId: install.runner_id }),
    ])

  if (error) {
    notFound()
  }

  return (
    <RunnerProvider initRunner={runner} shouldPoll>
      <PageSection className="@container" isScrollable>
        <Breadcrumbs
          breadcrumbs={[
            {
              path: `/${orgId}`,
              text: org?.name,
            },
            {
              path: `/${orgId}/installs`,
              text: 'Installs',
            },
            {
              path: `/${orgId}/installs/${installId}`,
              text: install?.name,
            },
            {
              path: `/${orgId}/installs/${installId}/runner`,
              text: 'Runner',
            },
          ]}
        />
        <div className="flex gap-4 justify-between">
          <hgroup>
            <Text variant="base" weight="strong">
              Install runner
            </Text>
          </hgroup>
          <ManagementDropdown
            settings={settings}
            isInstallRunner
            isManagedRunner={Boolean(heartbeat?.mng)}
          />
        </div>

        <div className="flex flex-col @min-4xl:flex-row gap-6">
          <AsyncBoundary
            errorFallback={<RunnerDetailsError />}
            loadingFallback={
              <RunnerDetailsCardSkeleton className="flex-initial" />
            }
          >
            <RunnerDetails
              orgId={orgId}
              runnerId={install?.runner_id}
              settings={settings}
            />
          </AsyncBoundary>

          <AsyncBoundary
            errorFallback={<RunnerHealthError />}
            loadingFallback={<RunnerHealthCardSkeleton className="flex-auto" />}
          >
            <RunnerHealth orgId={orgId} runnerId={install.runner_id} />
          </AsyncBoundary>
        </div>

        <div className="flex flex-col gap-6">
          <AsyncBoundary
            errorFallback={<RunnerActivityError />}
            loadingFallback={<RunnerRecentActivitySkeleton />}
          >
            <RunnerActivity
              orgId={orgId}
              offset={sp['offset'] || '0'}
              runnerId={install.runner_id}
            />
          </AsyncBoundary>
        </div>
      </PageSection>
    </RunnerProvider>
  )
}
