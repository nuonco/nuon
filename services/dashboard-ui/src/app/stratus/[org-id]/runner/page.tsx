import React, { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { Power } from '@phosphor-icons/react/dist/ssr'
import {
  Button,
  Card,
  Header,
  HeadingGroup,
  Page,
  RunnerDetailsCard,
  RunnerDetailsCardSkeleton,
  RunnerRecentActivity,
  ScrollableDiv,
  Section,
  Text,
  RECENT_ACTIVITY_LIMIT,
  RECENT_ACTIVITY_SEARCH_PARAM,
} from '@/stratus/components'
import type { IPageProps, TOrg, TRunnerHeartbeat, TRunnerJob } from '@/types'
import { nueQueryData } from '@/utils'

const StratusBuildRunner = async ({
  params,
  searchParams,
}: IPageProps<'org-id'>) => {
  const { ['org-id']: orgId } = await params
  const sp = await searchParams
  const { data: org } = await nueQueryData<TOrg>({
    orgId,
    path: 'orgs/current',
  })

  return (
    <Page
      breadcrumb={{
        baseCrumbs: [
          {
            path: `/stratus/${orgId}`,
            text: 'Home',
          },
          {
            path: `/stratus/${orgId}/runner`,
            text: 'Runner',
          },
        ],
      }}
    >
      <ScrollableDiv>
        <Header>
          <HeadingGroup>
            <Text variant="h3" weight="strong" level={1}>
              Build runner
            </Text>
            <Text theme="muted">
              View your organizations build runner performance and activities.
            </Text>
          </HeadingGroup>
          <Button variant="danger" className="self-center">
            <Power />
            Shutdown runner
          </Button>
        </Header>
        <Section className="gap-12">
          <div className="grid md:grid-cols-12 gap-6">
            <ErrorBoundary fallback={<RunnerError />}>
              <Suspense
                fallback={
                  <RunnerDetailsCardSkeleton className="md:col-span-6" />
                }
              >
                <LoadRunnerDetailsCard org={org} />
              </Suspense>
            </ErrorBoundary>

            <div className="flex flex-col gap-6 p-6 border rounded-md md:col-span-6">
              <Text variant="base" weight="strong">
                Health status
              </Text>
            </div>
          </div>
          <div className="flex flex-col gap-6">
            <Text variant="base" weight="strong">
              Recent activity
            </Text>

            <ErrorBoundary fallback={<RunnerError />}>
              <Suspense
                fallback={
                  <RunnerDetailsCardSkeleton className="md:col-span-6" />
                }
              >
                <LoadRunnerRecentActivity
                  org={org}
                  offset={sp[RECENT_ACTIVITY_SEARCH_PARAM] || '0'}
                />
              </Suspense>
            </ErrorBoundary>
          </div>
        </Section>
      </ScrollableDiv>
    </Page>
  )
}

export default StratusBuildRunner

const LoadRunnerDetailsCard: FC<{ org: TOrg }> = async ({ org }) => {
  const runnerGroup = org?.runner_group
  const runner = runnerGroup?.runners?.at(0)
  const { data: runnerHeartbeat, error } = await nueQueryData<TRunnerHeartbeat>(
    {
      orgId: org?.id,
      path: `runners/${runner?.id}/latest-heart-beat`,
    }
  )

  return runnerGroup && runner && !error ? (
    <RunnerDetailsCard
      runner={runner}
      runnerGroup={runnerGroup}
      runnerHeartbeat={runnerHeartbeat}
      className="md:col-span-6"
    />
  ) : (
    <RunnerError />
  )
}

const RunnerError: FC = () => (
  <Card className="md:col-span-6">
    <Text>Unable to load build runner</Text>
  </Card>
)

interface ILoadRunnerRecentActivity {
  offset: string
  org: TOrg
}

const LoadRunnerRecentActivity = async ({
  org,
  offset,
}: ILoadRunnerRecentActivity) => {
  const params = new URLSearchParams({
    offset,
    limit: RECENT_ACTIVITY_LIMIT.toString(),
  }).toString()

  const runnerGroup = org?.runner_group
  const runner = runnerGroup?.runners?.at(0)
  const {
    data: jobs,
    error,
    headers,
  } = await nueQueryData<Array<TRunnerJob>>({
    orgId: org?.id,
    path: `runners/${runner?.id}/jobs${params ? '?' + params : params}`,
    headers: {
      'x-nuon-pagination-enabled': true,
    },
  })

  const pageData = {
    hasNext: headers.get('x-nuon-page-next') || 'false',
    offset: headers.get('x-nuon-page-offset') || '0',
  }

  return runnerGroup && runner && jobs && !error ? (
    <RunnerRecentActivity
      jobs={jobs}
      pagination={{
        pageData,
        param: RECENT_ACTIVITY_SEARCH_PARAM,
        limit: RECENT_ACTIVITY_LIMIT,
      }}
    />
  ) : (
    <RunnerError />
  )
}
