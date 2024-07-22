import { GoClock, GoCloud } from 'react-icons/go'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  Card,
  Code,
  ComponentBuildCard,
  ComponentConfigCard,
  DeployLogsCard,
  DeployPlanCard,
  Duration,
  Grid,
  Heading,
  Page,
  PageHeader,
  PageSummary,
  PageTitle,
  Status,
  Text,
  Time,
} from '@/components'
import { BuildProvider, InstallDeployProvider } from '@/context'
import {
  getBuild,
  getComponent,
  getDeploy,
  getDeployLogs,
  getInstall,
  getOrg,
} from '@/lib'
import type { TInstallDeployLogs } from '@/types'

export default withPageAuthRequired(
  async function ({ params }) {
    const orgId = params?.['org-id'] as string
    const installId = params?.['install-id'] as string
    const installComponentId = params?.['component-id'] as string
    const deployId = params?.['deploy-id'] as string

    const deploy = await getDeploy({ orgId, installId, deployId })
    const buildId = deploy?.build_id
    const componentId = deploy?.component_id

    const [component, build, logs, install, org] = await Promise.all([
      getComponent({ componentId, orgId }),
      getBuild({ orgId, buildId }),
      getDeployLogs({ orgId, deployId, installId }).catch(console.error),
      getInstall({ orgId, installId }),
      getOrg({ orgId }),
    ])

    return (
      <InstallDeployProvider initDeploy={deploy} shouldPoll>
        <Page
          header={
            <PageHeader
              info={
                <>
                  <Status
                    status={deploy?.status}
                    label={
                      deploy.install_deploy_type === 'install'
                        ? 'deploy'
                        : deploy.install_deploy_type
                    }
                    isLabelStatusText
                  />
                  <div className="flex flex-col flex-auto gap-1">
                    <Text variant="caption">
                      <b>Install ID:</b> {installId}
                    </Text>
                    <Text variant="caption">
                      <b>Component ID:</b> {componentId}
                    </Text>
                    <Text variant="caption">
                      <b>Build ID:</b> {buildId}
                    </Text>
                  </div>
                </>
              }
              title={
                <PageTitle
                  overline={deployId}
                  title={`${component?.name} deployment`}
                />
              }
              summary={
                <PageSummary>
                  <Text variant="caption">
                    <GoCloud />
                    <Time time={deploy.updated_at} />
                  </Text>
                  <Text variant="caption">
                    <GoClock />
                    <Duration
                      unitDisplay="short"
                      listStyle="long"
                      variant="caption"
                      beginTime={deploy.created_at}
                      endTime={deploy.updated_at}
                    />
                  </Text>
                </PageSummary>
              }
            />
          }
          links={[
            { href: orgId, text: org.name },
            { href: installId, text: install.name },
            {
              href: 'components/' + installComponentId,
              text: component?.name,
            },
            { href: 'deploys/' + deployId, text: deployId },
          ]}
        >
          <Grid variant="3-cols">
            <div className="flex flex-col gap-6">
              <Heading variant="subtitle">Component details</Heading>

              <BuildProvider initBuild={build}>
                <ComponentBuildCard />
              </BuildProvider>

              <ComponentConfigCard
                orgId={orgId}
                componentId={componentId}
                componentConfigId={build?.component_config_connection_id}
              />
            </div>

            <div className="flex flex-col gap-6 lg:col-span-2">
              <Heading variant="subtitle">Deploy details</Heading>
              {deploy?.status === 'failed' ||
                (deploy?.status === 'error' && (
                  <Card>
                    <Heading className="text-red-500">
                      Deploy {deploy?.status}
                    </Heading>
                    <Code>{deploy?.status_description}</Code>
                  </Card>
                ))}

              <DeployLogsCard
                initLogs={logs as TInstallDeployLogs}
                shouldPoll
              />

              <DeployPlanCard
                orgId={orgId}
                installId={installId}
                deployId={deployId}
              />
            </div>
          </Grid>
        </Page>
      </InstallDeployProvider>
    )
  },
  { returnTo: '/dashboard' }
)
