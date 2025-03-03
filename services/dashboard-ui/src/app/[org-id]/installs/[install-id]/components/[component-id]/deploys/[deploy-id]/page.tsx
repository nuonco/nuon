import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  CalendarBlank,
  CaretRight,
  Timer,
} from '@phosphor-icons/react/dist/ssr'
import {
  CancelRunnerJobButton,
  ClickToCopy,
  ComponentConfiguration,
  DashboardContent,
  DeployStatus,
  Duration,
  InstallDeployIntermediateData,
  Link,
  LogStreamProvider,
  OperationLogsSection,
  StatusBadge,
  Section,
  Text,
  Time,
  ToolTip,
  Truncate,
} from '@/components'
import {
  getComponentBuild,
  getComponent,
  getComponentConfig,
  getInstall,
  getInstallDeploy,
  getInstallDeployPlan,
} from '@/lib'
import type { TInstallDeployPlan } from '@/types'
import { CANCEL_RUNNER_JOBS, DEPLOY_INTERMEDIATE_DATA } from '@/utils'

export default withPageAuthRequired(async function InstallComponentDeploy({
  params,
}) {
  const deployId = params?.['deploy-id'] as string
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string
  const deploy = await getInstallDeploy({
    installDeployId: deployId,
    installId,
    orgId,
  })
  const build = await getComponentBuild({ orgId, buildId: deploy.build_id })
  const [component, componentConfig, install, deployPlan] = await Promise.all([
    getComponent({ componentId: build.component_id, orgId }),
    getComponentConfig({
      componentId: build.component_id,
      componentConfigId: build.component_config_connection_id,
      orgId,
    }),
    getInstall({ installId, orgId }),
    getInstallDeployPlan({ deployId, installId, orgId }).catch(console.error),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/installs`, text: 'Installs' },
        {
          href: `/${orgId}/installs/${install.id}`,
          text: install.name,
        },
        {
          href: `/${orgId}/installs/${install.id}/components/${deploy.install_component_id}`,
          text: component.name,
        },
        {
          href: `/${orgId}/installs/${install.id}/components/${deploy.install_component_id}/deploys/${deploy.id}`,
          text: `${component.name} ${deploy.install_deploy_type}`,
        },
      ]}
      heading={`${component.name} ${deploy.install_deploy_type}`}
      headingUnderline={deploy.id}
      meta={
        <div className="flex gap-8 items-center justify-start pb-6">
          <Text>
            <CalendarBlank />
            <Time time={deploy.created_at} />
          </Text>
          <Text>
            <Timer />
            <Duration
              beginTime={deploy.created_at}
              endTime={deploy.updated_at}
            />
          </Text>
        </div>
      }
      statues={
        <div className="flex gap-6 items-start justify-start">
          <span className="flex flex-col gap-2">
            <DeployStatus
              descriptionAlignment="right"
              initDeploy={deploy}
              shouldPoll
            />
          </span>

          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Type
            </Text>
            <Text>{deploy.install_deploy_type}</Text>
          </span>

          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Build
            </Text>
            <Text variant="mono-12">
              <ToolTip alignment="right" tipContent={build.id}>
                <ClickToCopy>
                  <Truncate variant="small">{build.id}</Truncate>
                </ClickToCopy>
              </ToolTip>
            </Text>
          </span>

          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Component
            </Text>
            <Text variant="med-12">{component.name}</Text>
            <Text variant="mono-12">
              <ToolTip alignment="right" tipContent={deploy.component_id}>
                <ClickToCopy>
                  <Truncate variant="small">{deploy.component_id}</Truncate>
                </ClickToCopy>
              </ToolTip>
            </Text>
          </span>

          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Install
            </Text>
            <Text variant="med-12">{install.name}</Text>
            <Text variant="mono-12">
              <ToolTip alignment="right" tipContent={install.id}>
                <ClickToCopy>
                  <Truncate variant="small">{install.id}</Truncate>
                </ClickToCopy>
              </ToolTip>
            </Text>
          </span>
          {CANCEL_RUNNER_JOBS &&
          deploy?.status !== 'active' &&
          deploy?.status !== 'error' &&
          deploy?.status !== 'inactive' &&
          deploy?.runner_jobs?.length ? (
            <CancelRunnerJobButton
              jobType="deploy"
              runnerJobId={deploy?.runner_jobs?.at(-1)?.id}
              orgId={orgId}
            />
          ) : null}
        </div>
      }
    >
      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
        <div className="md:col-span-8">
          <LogStreamProvider initLogStream={deploy?.log_stream}>
            <OperationLogsSection heading="Deploy logs" />
          </LogStreamProvider>
        </div>

        <div className="divide-y flex flex-col md:col-span-4">
          <Section
            className="flex-initial"
            actions={
              <Text>
                <Link
                  href={`/${orgId}/apps/${component.app_id}/components/${component.id}/builds/${build.id}`}
                >
                  Details
                  <CaretRight />
                </Link>
              </Text>
            }
            heading="Component build"
          >
            <div className="flex items-start justify-start gap-6">
              <span className="flex flex-col gap-2">
                <StatusBadge
                  description={build.status_description}
                  status={build.status}
                  label="Status"
                />
              </span>

              <span className="flex flex-col gap-2">
                <Text className="text-cool-grey-600 dark:text-cool-grey-500">
                  Build date
                </Text>
                <Time time={build.created_at} />
              </span>

              <span className="flex flex-col gap-2">
                <Text className="text-cool-grey-600 dark:text-cool-grey-500">
                  Build duration
                </Text>
                <Duration
                  beginTime={build.created_at}
                  endTime={build.updated_at}
                />
              </span>
            </div>
          </Section>

          {DEPLOY_INTERMEDIATE_DATA &&
          (deployPlan as TInstallDeployPlan)?.actual?.waypoint_plan?.variables
            ?.intermediate_data?.nuon ? (
            <InstallDeployIntermediateData
              install={install}
              data={
                (deployPlan as TInstallDeployPlan)?.actual?.waypoint_plan
                  ?.variables?.intermediate_data
              }
            />
          ) : null}

          <Section heading="Component config">
            <ComponentConfiguration config={componentConfig} />
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
})
