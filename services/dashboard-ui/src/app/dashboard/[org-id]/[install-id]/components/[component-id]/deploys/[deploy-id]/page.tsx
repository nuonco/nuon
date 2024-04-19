import { DateTime } from 'luxon'
import React, { Suspense, type FC } from 'react'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  Build,
  Card,
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
  getDeploy,
  getDeployLogs,
  getDeployPlan,
  type IGetComponentConfig,
  type IGetDeployLogs,
  type IGetDeployPlan,
} from '@/lib'
import type {
  TComponentConfig,
  TInstallDeployLogs,
  TInstallDeployPlan,
} from '@/types'

const DeployLogs: FC<IGetDeployLogs> = async (params) => {
  let content = <>Loading...</>
  let logs: TInstallDeployLogs
  try {
    logs = await getDeployLogs(params)
    content = logs.length ? (
      <Logs logs={logs} />
    ) : (
      <Text variant="label">No logs to show</Text>
    )
  } catch (error) {
    content = <Text variant="label">Can not find deploy logs</Text>
  }

  return (
    <Card className="flex-initial">
      <Heading>Logs</Heading>
      {content}
    </Card>
  )
}

const DeployPlan: FC<IGetDeployPlan> = async (params) => {
  let content = <>Loading...</>
  let plan: TInstallDeployPlan
  try {
    plan = await getDeployPlan(params)
    content = <Plan plan={plan} />
  } catch (error) {
    content = <Text variant="label">No plan to show</Text>
  }
  return (
    <Card className="flex-1">
      <Heading>Deploy plan</Heading>
      {content}
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
  async function InstallDeployDashboard({ params }) {
    const orgId = params?.['org-id'] as string
    const installId = params?.['install-id'] as string
    const installComponentId = params?.['component-id'] as string
    const deployId = params?.['deploy-id'] as string

    const deploy = await getDeploy({ orgId, installId, deployId })
    const buildId = deploy?.build_id
    const componentId = deploy?.component_id

    const component = await getComponent({ componentId, orgId }).catch(
      () => null
    )

    return (
      <Page
        header={
          <PageHeader
            info={
              <>
                <Status status={deploy?.status} />
                <div className="flex flex-col flex-auto gap-1">
                  <Text variant="caption">
                    <b>Install ID:</b> {deploy?.install_id}
                  </Text>
                  <Text variant="caption">
                    <b>Build ID:</b> {deploy?.build_id}
                  </Text>
                  <Text variant="caption">
                    <b>Component ID:</b> {componentId}
                  </Text>
                </div>
              </>
            }
            title={
              <span className="flex flex-col gap-2">
                <Text variant="overline">{deploy?.id}</Text>
                <Heading
                  level={1}
                  variant="title"
                  className="flex gap-1 items-center"
                >
                  {component?.name} deploy
                </Heading>
              </span>
            }
            summary={
              <Text variant="caption">
                {DateTime.fromISO(deploy?.created_at).toRelative()}
              </Text>
            }
          />
        }
        links={[
          { href: orgId },
          { href: installId },
          {
            href: 'components/' + installComponentId,
            text: installComponentId,
          },
          { href: 'deploys/' + deployId, text: deployId },
        ]}
      >
        <Grid variant="3-cols">
          <div className="flex flex-col gap-6 overflow-hidden">
            <Heading variant="subtitle">Component details</Heading>

            <Suspense fallback="Loading...">
              <Build {...{ buildId, componentId, orgId }} />
            </Suspense>

            <Suspense fallback="Loading...">
              <LoadComponentConfig {...{ componentId, orgId }} />
            </Suspense>
          </div>

          <div className="flex flex-col gap-6 lg:col-span-2 overflow-hidden">
            <Heading variant="subtitle">Deploy details</Heading>
            <Suspense fallback="Loading...">
              <DeployLogs {...{ deployId, installId, orgId }} />
            </Suspense>

            <Suspense fallback="Loading...">
              <DeployPlan {...{ deployId, installId, orgId }} />
            </Suspense>
          </div>
        </Grid>
      </Page>
    )
  },
  { returnTo: '/dashboard' }
)
