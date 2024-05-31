import React, { Suspense } from 'react'
import { GoClock, GoPackage } from 'react-icons/go'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  BuildCommit,
  BuildLogsCard,
  BuildPlanCard,
  Card,
  Code,
  ComponentBuildStatus,
  ComponentConfig,
  Duration,
  Grid,
  Heading,
  Page,
  PageHeader,
  PageSummary,
  PageTitle,
  Text,
  Time,
} from '@/components'
import { BuildProvider } from '@/context'
import { getBuild, getBuildLogs } from '@/lib'
import type { TComponentBuildLogs } from '@/types'

export default withPageAuthRequired(
  async function ({ params }) {
    const orgId = params?.['org-id'] as string
    const buildId = params?.['build-id'] as string

    const build = await getBuild({ orgId, buildId })
    const buildLogs = await getBuildLogs({
      orgId,
      componentId: build.component_id,
      buildId,
    }).catch(console.error)

    return (
      <BuildProvider initBuild={build} shouldPoll>
        <Page
          header={
            <PageHeader
              info={
                <>
                  <ComponentBuildStatus />
                  <div className="flex flex-col flex-auto gap-1">
                    <Text variant="caption">
                      <b>Build ID:</b> {build.id}
                    </Text>
                    <Text variant="caption">
                      <b>Component ID:</b> {build.component_id}
                    </Text>
                    {build?.vcs_connection_commit && <BuildCommit />}
                  </div>
                </>
              }
              title={
                <PageTitle
                  overline={build.id}
                  title={`${build.component_name} build`}
                />
              }
              summary={
                <PageSummary>
                  <Text variant="caption">
                    <GoPackage />
                    <Time time={build?.updated_at} />
                  </Text>
                  <Text variant="caption">
                    <GoClock />
                    <Duration
                      unitDisplay="short"
                      listStyle="long"
                      variant="caption"
                      beginTime={build?.created_at}
                      endTime={build?.updated_at}
                    />
                  </Text>
                </PageSummary>
              }
            />
          }
          links={[{ href: orgId }, { href: buildId }]}
        >
          <Grid variant="3-cols">
            <div className="flex flex-col gap-6">
              <Heading variant="subtitle">Configuration</Heading>
              <Card>
                <Suspense fallback="Loading component config...">
                  <ComponentConfig
                    orgId={orgId}
                    componentId={build?.component_id}
                    componentConfigId={build.component_config_connection_id}
                    version={0}
                  />
                </Suspense>
              </Card>
            </div>

            <div className="flex flex-col gap-6 lg:col-span-2">
              <Heading variant="subtitle">Build details</Heading>
              {build?.status === 'failed' ||
                (build?.status === 'error' && (
                  <Card>
                    <Heading>Build {build?.status}</Heading>
                    <Code>{build?.status_description}</Code>
                  </Card>
                ))}

              <BuildLogsCard
                initLogs={buildLogs as TComponentBuildLogs}
                shouldPoll
              />

              <BuildPlanCard
                orgId={orgId}
                componentId={build.component_id}
                buildId={build.id}
              />
            </div>
          </Grid>
        </Page>
      </BuildProvider>
    )
  },
  { returnTo: '/dashboard' }
)
