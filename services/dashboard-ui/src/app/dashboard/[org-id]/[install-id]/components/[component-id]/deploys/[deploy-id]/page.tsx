import { DateTime } from 'luxon'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { Code, Heading, Logs, Page, Status, Text } from '@/components'
import {
  getBuild,
  getComponent,
  getComponentConfig,
  getDeploy,
  getDeployLogs,
  getDeployPlan,
} from '@/lib'

export default withPageAuthRequired(
  async function InstallDeployDashboard({ params }) {
    const orgId = params?.['org-id'] as string
    const installId = params?.['install-id'] as string
    const installComponentId = params?.['component-id'] as string
    const deployId = params?.['deploy-id'] as string

    const deploy = await getDeploy({ orgId, installId, deployId })
    const buildId = deploy?.build_id
    const componentId = deploy?.component_id

    const [build, component, config, logs, plan] = await Promise.all([
      getBuild({ buildId, componentId, orgId }),
      getComponent({ componentId, orgId }),
      getComponentConfig({ componentId, orgId }),
      getDeployLogs({ orgId, installId, deployId }),
      getDeployPlan({ orgId, installId, deployId }),
    ])

    return (
      <Page
        heading={
          <div className="flex flex-wrap items-end">
            <div className="flex flex-col flex-auto gap-2">
              <Text variant="overline">{deploy?.id}</Text>
              <Heading
                level={1}
                variant="title"
                className="flex gap-1 items-center"
              >
                {component?.name} deploy
              </Heading>
              <Text variant="caption">
                {DateTime.fromISO(deploy?.created_at).toRelative()}
              </Text>
            </div>

            <div className="flex flex-col flex-auto">
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
            </div>
          </div>
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
        <div className="h-fit overflow-auto">
          <Heading>Deploy</Heading>

          <Code variant="preformated">{JSON.stringify(deploy, null, 2)}</Code>

          <Heading>Logs</Heading>
          <Logs logs={logs} />

          <Heading>Deploy plan</Heading>

          <Heading variant="subheading">Rendered variables</Heading>
          <Code>
            {plan.actual?.waypoint_plan?.variables?.variables?.map((v, i) => {
              let variable = null
              if (v?.Actual?.TerraformVariable) {
                variable = (
                  <span className="flex" key={i?.toString()}>
                    <b>{v?.Actual?.TerraformVariable?.name}:</b>{' '}
                    {v?.Actual?.TerraformVariable?.value}
                  </span>
                )
              }

              if (v?.Actual?.HelmValue) {
                variable = (
                  <span className="flex" key={i?.toString()}>
                    <b>{v?.Actual?.HelmValue?.name}:</b>{' '}
                    {v?.Actual?.HelmValue?.value}
                  </span>
                )
              }

              return variable
            })}
          </Code>

          <Heading variant="subheading">Intermediate variables</Heading>
          <Code variant="preformated">
            {JSON.stringify(
              plan.actual?.waypoint_plan?.variables?.intermediate_data,
              null,
              2
            )}
          </Code>

          <Heading variant="subheading">Job config</Heading>
          <Code variant="preformated">
            {plan.actual?.waypoint_plan?.waypoint_job?.hcl_config}
          </Code>
        </div>
      </Page>
    )
  },
  { returnTo: '/dashboard' }
)
