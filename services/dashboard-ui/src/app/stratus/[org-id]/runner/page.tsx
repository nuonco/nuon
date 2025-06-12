import React, { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { Power } from '@phosphor-icons/react/dist/ssr'
import {
  RunnerMeta,
  RunnerHealthChart,
  RunnerPastJobs,
} from '@/components/Runners'
import {
  Button,
  Header,
  HeaderGroup,
  Page,
  ScrollableDiv,
  Section,
  Skeleton,
  Text,
} from '@/stratus/components'
import type { IPageProps, TOrg, TRunner } from '@/types'
import { nueQueryData } from '@/utils'

const StratusBuildRunner: FC<IPageProps<'org-id'>> = async ({ params }) => {
  const orgId = params?.['org-id']
  const { data: org } = await nueQueryData<TOrg>({
    orgId,
    path: 'orgs/current',
  })
  const { data: runner } = await nueQueryData<TRunner>({
    orgId,
    path: `runners/${org?.runner_group?.runners?.at(0)?.id}`,
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
          <HeaderGroup>
            <Text variant="h3" weight="strong" level={1}>
              Build runner
            </Text>
            <Text theme="muted">
              View your organizations build runner performance and activities.
            </Text>
          </HeaderGroup>
          <Button variant="danger" className="self-center">
            <Power />
            Shutdown runner
          </Button>
        </Header>
        <Section className="gap-12">
          <div className="grid md:grid-cols-12 gap-6">
            <div className="flex flex-col gap-6 p-6 border rounded-md md:col-span-6">
              <Text variant="base" weight="strong">
                Runner details
              </Text>

              <RunnerMeta orgId={orgId} runner={runner} />
            </div>

            <div className="flex flex-col gap-6 p-6 border rounded-md md:col-span-6">
              <Text variant="base" weight="strong">
                Health status
              </Text>
              <div className="align-end">
                <RunnerHealthChart orgId={orgId} runnerId={runner?.id} />
              </div>
            </div>
          </div>
          <div className="flex flex-col gap-6">
            <Text variant="base" weight="strong">
              Recent activity
            </Text>

            <RunnerPastJobs orgId={orgId} runnerId={runner?.id} offset="10" />
          </div>
        </Section>
      </ScrollableDiv>
    </Page>
  )
}

export default StratusBuildRunner
