import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { FiChevronRight, FiCloud, FiClock } from 'react-icons/fi'
import {
  ComponentConfiguration,
  DashboardContent,
  Duration,
  Heading,
  Link,
  StatusBadge,
  Text,
  Time,
  ToolTip,
} from '@/components'
import {
  getBuild,
  getComponent,
  getComponentConfig,
  getOrg,
  getInstall,
  getDeploy,
} from '@/lib'

export default withPageAuthRequired(
  async function InstallComponentDeploy({ params }) {
    const deployId = params?.['deploy-id'] as string
    const installId = params?.['install-id'] as string
    const orgId = params?.['org-id'] as string
    const deploy = await getDeploy({ deployId, installId, orgId })
    const build = await getBuild({ orgId, buildId: deploy.build_id })
    const [component, componentConfig, install, org] = await Promise.all([
      getComponent({ componentId: build.component_id, orgId }),
      getComponentConfig({
        componentId: build.component_id,
        componentConfigId: build.component_config_connection_id,
        orgId,
      }),
      getInstall({ installId, orgId }),
      getOrg({ orgId }),
    ])

    return (
      <DashboardContent
        breadcrumb={[
          { href: `/beta/${org.id}`, text: org.name },
          { href: `/beta/${org.id}/installs`, text: 'Installs' },
          {
            href: `/beta/${org.id}/installs/${install.id}`,
            text: install.name,
          },
          {
            href: `/beta/${org.id}/installs/${install.id}/components/${deploy.install_component_id}`,
            text: component.name,
          },
          {
            href: `/beta/${org.id}/installs/${install.id}/components/${deploy.install_component_id}/deploys/${deploy.id}`,
            text: `${component.name} ${deploy.install_deploy_type}`,
          },
        ]}
        heading={`${component.name} ${deploy.install_deploy_type}`}
        headingUnderline={deploy.id}
        meta={
          <div className="flex gap-8 items-center justify-start pb-6">
            <Text variant="caption">
              <FiCloud />
              <Time time={deploy.created_at} variant="caption" />
            </Text>
            <Text variant="caption">
              <FiClock />
              <Duration
                beginTime={deploy.created_at}
                endTime={deploy.updated_at}
                variant="caption"
              />
            </Text>
          </div>
        }
        statues={
          <div className="flex gap-6 items-start justify-start">
            <span className="flex flex-col gap-2">
              <Text variant="overline">Status</Text>
              <StatusBadge
                descriptionAlignment="right"
                descriptionPosition="bottom"
                description={deploy.status_description}
                status={deploy.status}
              />
            </span>

            <span className="flex flex-col gap-2">
              <Text variant="overline">Type</Text>
              <Text variant="caption">{deploy.install_deploy_type}</Text>
            </span>

            <span className="flex flex-col gap-2">
              <Text variant="overline">Build:</Text>
              <Text variant="caption">{build.id}</Text>
            </span>

            <span className="flex flex-col gap-2">
              <Text variant="overline">Component</Text>
              <Text variant="label">{component.name}</Text>
              <Text variant="caption">{deploy.component_id}</Text>
            </span>

            <span className="flex flex-col gap-2">
              <Text variant="overline">Install</Text>
              <Text variant="label">{install.name}</Text>
              <Text variant="caption">{install.id}</Text>
            </span>
          </div>
        }
      >
        <div className="flex flex-col lg:flex-row flex-auto">
          <section className="flex flex-auto flex-col gap-4 px-6 py-8 border-r overflow-auto">
            <Heading>Deploy details</Heading>

            <Text>New runner logs here</Text>
          </section>
          <div
            className="divide-y flex flex-col lg:min-w-[550px]
lg:max-w-[550px]"
          >
            <section className="flex flex-col gap-6 px-6 py-8">
              <div className="flex items-center justify-between">
                <Heading>Component build</Heading>

                <Text variant="caption">
                  <Link
                    href={`/beta/${orgId}/apps/${component.app_id}/components/${component.id}/builds/${build.id}`}
                  >
                    Details
                    <FiChevronRight />
                  </Link>
                </Text>
              </div>

              <div className="flex items-start justify-start gap-6">
                <span className="flex flex-col gap-2">
                  <Text variant="overline">Status</Text>
                  <StatusBadge
                    description={build.status_description}
                    status={build.status}
                  />
                </span>

                <span className="flex flex-col gap-2">
                  <Text variant="overline">Build date</Text>
                  <Time
                    variant="caption"
                    time={build.created_at}
                    format="long"
                  />
                </span>

                <span className="flex flex-col gap-2">
                  <Text variant="overline">Build duration</Text>
                  <Duration
                    variant="caption"
                    beginTime={build.created_at}
                    endTime={build.updated_at}
                  />
                </span>
              </div>
            </section>
            <section className="flex flex-col gap-6 px-6 py-8">
              <Heading>Component config</Heading>

              <ComponentConfiguration config={componentConfig} />
            </section>
          </div>
        </div>
      </DashboardContent>
    )
  },
  { returnTo: '/dashboard' }
)
