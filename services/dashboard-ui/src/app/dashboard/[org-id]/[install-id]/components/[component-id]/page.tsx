import React, { type FC, Suspense } from 'react'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  Card,
  ComponentConfig,
  ComponentDependencies,
  Grid,
  Heading,
  InstallComponentHeading,
  InstallComponentSummary,
  InstallDeploys,
  Page,
  PageHeader,
  LatestDeploy2,
} from '@/components'
import {
  getBuild,
  getComponent,
  getComponentConfig,
  getInstall,
  getInstallComponent,
} from '@/lib'
import type { TBuild, TInstall, TInstallDeploy } from '@/types'

const Build: FC<{
  buildId: string
  componentId: string
  orgId: string
}> = async ({ buildId, componentId, orgId }) => {
  let build: TBuild
  try {
    build = await getBuild({ buildId, componentId, orgId })
  } catch (error) {
    return <>No build found</>
  }

  return (
    <span className="text-xs">
      <b>Commit SHA:</b> {build?.vcs_connection_commit?.sha.slice(0, 7)}
    </span>
  )
}

const ComponentApp: FC<{ installId: string; orgId: string }> = async ({
  installId,
  orgId,
}) => {
  let install: TInstall
  try {
    install = await getInstall({ installId, orgId })
  } catch (error) {
    return <>App not found</>
  }

  return <>{install?.app?.name}</>
}

export default withPageAuthRequired(
  async function InstallComponentDashboard({ params }) {
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

    const [component, config] = await Promise.all([
      getComponent({ componentId, orgId }),
      getComponentConfig({ componentId, orgId }),
    ])

    return (
      <Page
        header={
          <PageHeader
            info={
              <div className="flex flex-col">
                <LatestDeploy2 {...installComponent} />
                <Suspense fallback={<span>Loading...</span>}>
                  <Build
                    buildId={buildId}
                    componentId={componentId}
                    orgId={orgId}
                  />
                </Suspense>
              </div>
            }
            title={<InstallComponentHeading component={component} />}
            summary={
              <InstallComponentSummary
                config={config}
                appName={
                  <Suspense fallback="loading...">
                    <ComponentApp installId={installId} orgId={orgId} />
                  </Suspense>
                }
              />
            }
          />
        }
        links={[
          {
            href: installComponent?.org_id,
            text: installComponent?.org_id,
          },

          {
            href: installComponent?.install_id,
            text: installComponent?.install_id,
          },

          {
            href: 'components/' + installComponent?.id,
            text: installComponent?.id,
          },
        ]}
      >
        <Grid variant="3-cols" >
          <div className="flex flex-col gap-6 overflow-hidden">
            <Heading variant="subtitle">Deploy history</Heading>
            <Card>
              <InstallDeploys
                deploys={
                  installComponent?.install_deploys as Array<TInstallDeploy>
                }
                installId={installComponent?.install_id}
              />
            </Card>
          </div>
          <div className="flex flex-col gap-6 lg:col-span-2 overflow-auto">
            <Heading variant="subtitle">Details</Heading>

            <Card>
              <Heading variant="subheading">Configuration</Heading>
              <ComponentConfig
                config={config}
                version={component?.config_versions}
              />
            </Card>

            {component?.dependencies?.length ? (
              <Card>
                <Heading variant="subheading">Dependencies</Heading>
                <ComponentDependencies deps={component?.dependencies} />
              </Card>
            ) : null}
          </div>
        </Grid>
      </Page>
    )
  },
  { returnTo: '/dashboard' }
)
