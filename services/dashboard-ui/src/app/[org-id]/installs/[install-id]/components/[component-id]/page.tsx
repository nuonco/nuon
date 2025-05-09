import type { Metadata } from 'next'
import { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { CaretRight } from '@phosphor-icons/react/dist/ssr'
import {
  ClickToCopyButton,
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
import { InstallComponentManagementDropdown } from '@/components/InstallComponents/ManagementDropdown'
import { TerraformWorkspace } from '@/components/InstallSandbox/TerraformWorkspace'
import {
  getComponent,
  getComponentConfig,
  getInstall,
  getInstallComponentDeploys,
  getInstallComponentOutputs,
  getLatestComponentBuild,
  getInstallComponent,
  getOrg,
} from '@/lib'
import type { TComponent, TInstallComponent } from '@/types'

export async function generateMetadata({ params }): Promise<Metadata> {
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string
  const componentId = params?.['component-id'] as string
  const [install, component] = await Promise.all([
    getInstall({ installId, orgId }),
    getComponent({ componentId, orgId }),
  ])

  return {
    title: `${install.name} | ${component.name}`,
  }
}

export default withPageAuthRequired(async function InstallComponent({
  params,
}) {
  const componentId = params?.['component-id'] as string
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string

  const [org, install, component, installComponent] = await Promise.all([
    getOrg({ orgId }),
    getInstall({ installId, orgId }),
    getComponent({ componentId, orgId }),
    getInstallComponent({ orgId, installId, componentId }),
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
          href: `/${orgId}/installs/${install.id}/components/${componentId}`,
          text: component.name,
        },
      ]}
      heading={component.name}
      headingUnderline={component.id}
      statues={
        <div className="flex gap-8">
          <InstallDeployLatestBuildButton
            componentId={componentId}
            installId={installId}
            orgId={orgId}
          />
          <InstallComponentManagementDropdown component={component} />
        </div>
      }
    >
      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
        <div className="divide-y flex-auto flex flex-col md:col-span-8">
          <Section
            actions={
              <Text>
                <Link
                  href={`/${orgId}/apps/${component.app_id}/components/${component.id}`}
                >
                  Details
                  <CaretRight />
                </Link>
              </Text>
            }
            className="flex-initial"
            heading="Component config"
            childrenClassName="flex flex-col gap-4"
          >
            <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={
                  <Loading
                    loadingText="Loading component config..."
                    variant="stack"
                  />
                }
              >
                <LoadComponentConfig componentId={componentId} orgId={orgId} />
              </Suspense>
            </ErrorBoundary>
            {org?.features?.['terraform-workspace'] || (
              <ErrorBoundary fallbackRender={ErrorFallback}>
                <Suspense
                  fallback={<Loading loadingText="Loading latest outputs..." />}
                >
                  <LoadLatestOutputs
                    componentId={componentId}
                    installId={installId}
                    orgId={orgId}
                  />
                </Suspense>
              </ErrorBoundary>
            )}
            {org?.features?.['terraform-workspace'] &&
            component?.type === 'terraform_module' ? (
              <ErrorBoundary fallbackRender={ErrorFallback}>
                <Suspense
                  fallback={
                    <Loading loadingText="Loading latest terraform workspace..." />
                  }
                >
                  <TerraformWorkspace
                    orgId={orgId}
                    workspace={installComponent.terraform_workspace}
                  />
                </Suspense>
              </ErrorBoundary>
            ) : null}
          </Section>

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
        </div>
        <div className="divide-y flex flex-col md:col-span-4">
          <Section heading="Deploy history">
            <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={
                  <Loading
                    loadingText="Loading deploy history..."
                    variant="stack"
                  />
                }
              >
                <LoadDeployHistory
                  component={component}
                  installId={installId}
                  orgId={orgId}
                />
              </Suspense>
            </ErrorBoundary>
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
})

const LoadDeployHistory: FC<{
  component: TComponent
  installId: string
  orgId: string
}> = async ({ component, installId, orgId }) => {
  const deploys = await getInstallComponentDeploys({
    componentId: component.id,
    installId,
    orgId,
  }).catch(console.error)

  return deploys ? (
    <InstallComponentDeploys
      component={component}
      initDeploys={deploys}
      installId={installId}
      installComponentId={component.id}
      orgId={orgId}
      shouldPoll
    />
  ) : (
    <Text>Unable to load deploy history.</Text>
  )
}

const LoadLatestOutputs: FC<{
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
        <ClickToCopyButton textToCopy={JSON.stringify(outputs)} />
      </div>
      <CodeViewer
        initCodeSource={JSON.stringify(outputs, null, 2)}
        language="json"
      />
    </div>
  ) : null
}

const LoadComponentConfig: FC<{ componentId: string; orgId: string }> = async ({
  componentId,
  orgId,
}) => {
  const componentConfig = await getComponentConfig({
    componentId,
    orgId,
  }).catch(console.error)
  return componentConfig ? (
    <ComponentConfiguration config={componentConfig} isNotTruncated />
  ) : (
    <Text>No component config found.</Text>
  )
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
            <CaretRight />
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
