import React, { type FC, Suspense } from 'react'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  Card,
  ComponentConfig,
  ComponentDependencies,
  Heading,
  InstallComponentHeading,
  InstallDeploys,
  Page,
} from '@/components'
import {
  getBuild,
  getComponent,
  getComponentConfig,
  getInstall,
  getInstallComponent,
} from '@/lib'
import type { TBuild, TInstallDeploy } from '@/types'

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

  return <>{build?.vcs_connection_commit?.message}</>
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

    const [component, config, install] = await Promise.all([
      getComponent({ componentId, orgId }),
      getComponentConfig({ componentId, orgId }),
      getInstall({ installId, orgId }),
    ])

    return (
      <Page
        heading={
          <InstallComponentHeading
            component={component}
            config={config}
            install={install}
            installComponent={installComponent}
            buildInfo={
              <Suspense fallback={<span>Loading...</span>}>
                <Build
                  buildId={buildId}
                  componentId={componentId}
                  orgId={orgId}
                />
              </Suspense>
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
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8 w-full h-fit overflow-hidden">
          <div className="flex flex-col gap-6 overflow-hidden">
            <Heading variant="subtitle">Deploy history</Heading>
            <Card>
              <InstallDeploys
                deploys={
                  installComponent?.install_deploys as Array<TInstallDeploy>
                }
                installId={install?.id}
              />
            </Card>
          </div>
          <div className="flex flex-col gap-6 lg:col-span-2 overflow-auto">
            <Suspense fallback={<span>Loading...</span>}>
              <Build
                buildId={buildId}
                componentId={componentId}
                orgId={orgId}
              />
            </Suspense>

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
        </div>
      </Page>
    )
  },
  { returnTo: '/dashboard' }
)
