import { DateTime } from 'luxon'
import React, { Suspense, type FC } from 'react'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  BuildLogs,
  Card,
  Code,
  ComponentConfig,
  Grid,
  Heading,
  Logs,
  Page,
  PageHeader,
  Plan,
  Status,
  Text,
} from '@/components'
import {
  getComponent,
  getComponentConfig,
  getBuild,
  getBuildLogs,
  getBuildPlan,
  type IGetComponentConfig,
  type IGetBuildLogs,
  type IGetBuildPlan,
} from '@/lib'
import type {
  TComponentConfig,
  TComponentBuildLogs,
  TComponentBuildPlan,
} from '@/types'
import { BuildHeader } from './build-header'

const LoadBuildLogs: FC<IGetBuildLogs> = async (params) => {
  let logs: TComponentBuildLogs
  try {
    logs = await getBuildLogs(params)
  } catch (error) {
    return <Text variant="label">Can not find build logs</Text>
  }

  return (
    <Card className="flex-initial">
      <Heading>Build logs</Heading>
      <BuildLogs ssrLogs={logs} {...{ ...params }} />
    </Card>
  )
}

const BuildPlan: FC<IGetBuildPlan> = async (params) => {
  let plan: TComponentBuildPlan
  try {
    plan = await getBuildPlan(params)
  } catch (error) {
    return <Text variant="label">No build plan to show</Text>
  }
  return (
    <Card className="flex-1">
      <Heading>Build plan</Heading>
      <Plan plan={plan} />
    </Card>
  )
}

const LoadComponentConfig: FC<IGetComponentConfig> = async (params) => {
  let config: TComponentConfig
  try {
    config = await getComponentConfig(params)
  } catch (error) {
    return <>No config to show</>
  }
  return (
    <Card>
      <Heading>Component config</Heading>
      <ComponentConfig config={config} version={1} />
    </Card>
  )
}

export default withPageAuthRequired(
  async function BuildDashboard({ params }) {
    const orgId = params?.['org-id'] as string
    const buildId = params?.['build-id'] as string

    const build = await getBuild({ orgId, buildId })
    const component = await getComponent({
      componentId: build?.component_id as string,
      orgId,
    }).catch(() => null)

    return (
      <Page
        header={
          <BuildHeader
            ssrBuild={build}
            component={component}
            orgId={orgId}
            buildId={buildId}
          />
        }
        links={[{ href: orgId }, { href: buildId }]}
      >
        <Grid variant="3-cols">
          <div className="flex flex-col gap-6">
            <Heading variant="subtitle">Component details</Heading>
            <Suspense fallback="Loading...">
              <LoadComponentConfig
                {...{ componentId: build?.component_id as string, orgId }}
              />
            </Suspense>
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

            <Suspense fallback="Loading...">
              <LoadBuildLogs
                buildId={buildId}
                componentId={component?.id}
                orgId={orgId}
              />
            </Suspense>

            <Suspense fallback="Loading...">
              <BuildPlan {...{ buildId, componentId: component?.id, orgId }} />
            </Suspense>
          </div>
        </Grid>
      </Page>
    )
  },
  { returnTo: '/dashboard' }
)
