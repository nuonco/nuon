import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  Card,
  ComponentBuildCard,
  ComponentConfigCard,
  ComponentConfigType,
  ComponentDependenciesCard,
  Grid,
  Heading,
  InstallComponentStatus,
  InstallDeploys,
  Page,
  PageHeader,
  PageSummary,
  PageTitle,
  LatestDeploy,
  Text,
} from '@/components'
import {
  BuildProvider,
  InstallProvider,
  InstallComponentProvider,
} from '@/context'
import {
  getBuild,
  getComponent,
  getInstall,
  getInstallComponent,
  getOrg,
} from '@/lib'
import type { TInstallDeploy } from '@/types'

export default withPageAuthRequired(
  async function ({ params }) {
    const orgId = params?.['org-id'] as string
    const installId = params?.['install-id'] as string
    const installComponentId = params?.['component-id'] as string

    const installComponent = await getInstallComponent({
      installComponentId,
      installId,
      orgId,
    })

    const buildId = installComponent?.install_deploys?.[0]?.build_id
    const componentId = installComponent?.component_id

    const [component, build, install, org] = await Promise.all([
      getComponent({ componentId, orgId }),
      getBuild({ orgId, buildId }),
      getInstall({ installId, orgId }),
      getOrg({ orgId }),
    ])

    return (
      <InstallProvider initInstall={install}>
        <InstallComponentProvider
          initInstallComponent={{ ...installComponent, org_id: orgId }}
          shouldPoll
        >
          <Page
            header={
              <PageHeader
                title={
                  <PageTitle overline={component.id} title={component.name} />
                }
                summary={
                  <PageSummary>
                    <Text variant="status">{install.app?.name}</Text>
                    <ComponentConfigType
                      orgId={orgId}
                      componentId={componentId}
                    />
                  </PageSummary>
                }
                info={
                  <div className="flex flex-col">
                    <InstallComponentStatus />
                    <LatestDeploy
                      {...(installComponent
                        ?.install_deploys?.[0] as TInstallDeploy)}
                    />
                  </div>
                }
              />
            }
            links={[
              {
                href: installComponent?.org_id,
                text: org?.name,
              },

              {
                href: installComponent?.install_id,
                text: install?.name,
              },

              {
                href: 'components/' + installComponent?.id,
                text: installComponent?.component?.name,
              },
            ]}
          >
            <Grid variant="3-cols">
              <div className="flex flex-col gap-6">
                <Heading variant="subtitle">Deploy history</Heading>
                <Card className="max-h-[40rem]">
                  <InstallDeploys />
                </Card>
              </div>
              <div className="flex flex-col gap-6 lg:col-span-2">
                <Heading variant="subtitle">Details</Heading>

                <div className="grid grid-cols-2 gap-6">
                  <BuildProvider initBuild={build}>
                    <ComponentBuildCard heading="Latest build" />
                  </BuildProvider>

                  <ComponentDependenciesCard
                    orgId={orgId}
                    componentId={componentId}
                  />
                </div>

                <ComponentConfigCard
                  heading="Latest config"
                  componentId={componentId}
                  orgId={orgId}
                />
              </div>
            </Grid>
          </Page>
        </InstallComponentProvider>
      </InstallProvider>
    )
  },
  { returnTo: '/dashboard' }
)
