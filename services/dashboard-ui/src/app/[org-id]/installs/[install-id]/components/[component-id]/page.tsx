import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { FiChevronRight } from 'react-icons/fi'
import {
  ClickToCopy,
  ComponentConfiguration,
  CodeViewer,
  DashboardContent,
  DependentComponents,
  Duration,
  ErrorFallback,
  InstallComponentDeploys,
  InstallDeployLatestBuildButton,
  Link,
  Loading,
  StatusBadge,
  Section,
  Text,
  Time,
} from '@/components'
import {
  getComponent,
  getComponentBuilds,
  getComponentConfig,
  getInstall,
  getInstallComponent,
  getInstallComponentOutputs,
  getLatestComponentBuild,
} from '@/lib'
import type { TInstallComponent } from '@/types'

export default withPageAuthRequired(async function InstallComponent({
  params,
}) {
  const installComponentId = params?.['component-id'] as string
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string

  const [install, installComponent] = await Promise.all([
    getInstall({ installId, orgId }),
    getInstallComponent({
      installComponentId,
      installId,
      orgId,
    }),
  ])

  const [component, componentConfig, builds] = await Promise.all([
    getComponent({ componentId: installComponent.component_id, orgId }),
    getComponentConfig({
      componentId: installComponent?.component_id,
      orgId,
    }),
    getComponentBuilds({ componentId: installComponent?.component_id, orgId }),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/installs`, text: 'Installs' },
        {
          href: `/${orgId}/installs/${install.id}/components`,
          text: install.name,
        },
        {
          href: `/${orgId}/installs/${install.id}/components/${installComponent.id}`,
          text: component.name,
        },
      ]}
      heading={component.name}
      headingUnderline={installComponent.id}
      statues={
        <InstallDeployLatestBuildButton
          builds={builds}
          installId={installId}
          orgId={orgId}
        />
      }
    >
      <div className="flex flex-col lg:flex-row flex-auto">
        <div className="divide-y flex-auto  flex flex-col overlfow-auto">
          {component.dependencies && (
            <Section className="flex-initial" heading="Dependencies">
              <DependentComponents
                dependentIds={component.dependencies}
                installComponents={
                  install?.install_components as Array<TInstallComponent>
                }
                installId={installId}
                orgId={orgId}
              />
            </Section>
          )}
          <Section
            heading="Component config"
            childrenClassName="flex flex-col gap-4"
          >
            <ComponentConfiguration config={componentConfig} />
            <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={<Loading loadingText="Loading latest build..." />}
              >
                <LatestOutputs
                  componentId={installComponent?.component_id}
                  installId={installId}
                  orgId={orgId}
                />
              </Suspense>
            </ErrorBoundary>
          </Section>
        </div>
        <div className="border-l overflow-auto lg:min-w-[450px] lg:max-w-[450px]">
          <Section heading="Deploy history">
            <InstallComponentDeploys
              component={component}
              initDeploys={installComponent.install_deploys}
              installId={installId}
              installComponentId={installComponent.id}
              orgId={orgId}
              shouldPoll
            />
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
})

const LatestOutputs: FC<{
  componentId: string
  installId: string
  orgId: string
}> = async ({ componentId, installId, orgId }) => {
  const outputs = await getInstallComponentOutputs({
    componentId,
    installId,
    orgId,
  }).catch(console.error)
  
  return outputs ? (
    <div className="flex flex-col gap-2">
      <div className="flex items-center justify-between">
        <Text variant="med-12">Outputs</Text>
        <ClickToCopy className="hover:bg-black/10 rounded-md p-1 text-sm">
          <span className="hidden">{JSON.stringify(outputs)}</span>
        </ClickToCopy>
      </div>
      <CodeViewer
        initCodeSource={JSON.stringify(outputs, null, 2)}
        language="json"
      />
    </div>
  ) : null
}

const LatestBuild = async ({ component, orgId }) => {
  const build = await getLatestComponentBuild({
    componentId: component?.id,
    orgId,
  })

  return (
    <Section
      className="flex-initial"
      actions={
        <Text>
          <Link
            href={`/${orgId}/apps/${component.app_id}/components/${component.id}/builds/${build.id}`}
          >
            Details
            <FiChevronRight />
          </Link>
        </Text>
      }
      heading="Latest build"
    >
      <div className="flex items-end justify-between">
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
            <Duration beginTime={build.created_at} endTime={build.updated_at} />
          </span>
        </div>
      </div>
    </Section>
  )
}
