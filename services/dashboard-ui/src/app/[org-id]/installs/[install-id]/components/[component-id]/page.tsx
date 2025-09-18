import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { CaretRightIcon } from '@phosphor-icons/react/dist/ssr'
import {
  ClickToCopyButton,
  ComponentConfiguration,
  ComponentDependencies,
  CodeViewer,
  DashboardContent,
  ErrorFallback,
  InstallDeployLatestBuildButton,
  InstallComponentManagementDropdown,
  Link,
  Loading,
  Section,
  Text,
} from '@/components'
import {
  TerraformWorkspace,
  ValuesFileModal,
} from '@/components/InstallSandbox'
import {
  getInstallById,
  getInstallComponentOutputs,
  getInstallComponentById,
  getOrgById,
} from '@/lib'
import type {
  TAppConfig,
  TComponent,
  TComponentConfig,
  TInstall,
} from '@/types'
import { nueQueryData } from '@/utils'
import { Deploys } from './deploys'

export async function generateMetadata({ params }): Promise<Metadata> {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['component-id']: componentId,
  } = await params
  const [{ data: install }, { data: installComponent }] = await Promise.all([
    getInstallById({ installId, orgId }),
    getInstallComponentById({ componentId, installId, orgId }),
  ])

  return {
    title: `${installComponent?.component?.name} | ${install.name} | Nuon`,
  }
}

export default async function InstallComponent({ params, searchParams }) {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['component-id']: componentId,
  } = await params
  const sp = await searchParams
  const [
    { data: org },
    { data: install },
    { data: installComponent, error, status },
  ] = await Promise.all([
    getOrgById({ orgId }),
    getInstallById({ installId, orgId }),
    getInstallComponentById({ orgId, installId, componentId }),
  ])

  if (error) {
    console.error(
      'Error rendering install component page: ',
      `API status: ${status}`,
      error
    )
    if (status === 404) {
      notFound()
    } else {
      // TODO(nnnat): show error message
      notFound()
    }
  }

  const component = installComponent?.component

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/installs`, text: 'Installs' },
        {
          href: `/${orgId}/installs/${install.id}`,
          text: install.name,
        },
        {
          href: `/${orgId}/installs/${install?.id}/components`,
          text: 'Components',
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
          <InstallComponentManagementDropdown
            componentId={installComponent?.component_id}
            componentName={installComponent?.component?.name}
          />
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
                  <CaretRightIcon />
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
                <LoadComponentConfig
                  componentId={componentId}
                  install={install}
                  orgId={orgId}
                />
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
          </Section>
          {org?.features?.['terraform-workspace'] &&
          component?.type === 'terraform_module' ? (
            <Section
              className="flex-initial"
              childrenClassName="flex flex-col gap-4"
            >
              <ErrorBoundary fallbackRender={ErrorFallback}>
                <Suspense
                  fallback={
                    <Section heading="Terraform workspace">
                      <Loading
                        loadingText="Loading latest terraform workspace..."
                        variant="stack"
                      />
                    </Section>
                  }
                >
                  <TerraformWorkspace
                    orgId={orgId}
                    workspace={installComponent.terraform_workspace}
                  />
                </Suspense>
              </ErrorBoundary>
            </Section>
          ) : null}

          {component.dependencies && (
            <Section className="flex-initial" heading="Dependencies">
              <ErrorBoundary fallbackRender={ErrorFallback}>
                <Suspense
                  fallback={
                    <Loading
                      variant="stack"
                      loadingText="Loading component dependencies..."
                    />
                  }
                >
                  <LoadComponentDependencies
                    component={component}
                    orgId={orgId}
                    installId={installId}
                  />
                </Suspense>
              </ErrorBoundary>
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
                <Deploys
                  component={component}
                  installId={installId}
                  orgId={orgId}
                  offset={sp['offset'] || '0'}
                />
              </Suspense>
            </ErrorBoundary>
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
}

const LoadLatestOutputs: FC<{
  componentId: string
  installId: string
  orgId: string
}> = async ({ componentId, installId, orgId }) => {
  const { data: outputs, error } = await getInstallComponentOutputs({
    componentId,
    installId,
    orgId,
  })

  return outputs && !error ? (
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

const LoadComponentConfig: FC<{
  install: TInstall
  componentId: string
  orgId: string
}> = async ({ componentId, install, orgId }) => {
  const { data: config, error } = await nueQueryData<TAppConfig>({
    orgId,
    path: `apps/${install?.app_id}/config/${install?.app_config_id}?recurse=true`,
  })

  const componentConfig = config?.component_config_connections?.find(
    (c) => c.component_id === componentId
  )

  return error ? (
    <Text>{error?.error}</Text>
  ) : componentConfig ? (
    <>
      <ComponentConfiguration config={componentConfig} isNotTruncated />
      {componentConfig?.terraform_module?.variables_files?.length ? (
        <ValuesFileModal
          valuesFiles={componentConfig?.terraform_module?.variables_files}
        />
      ) : null}
    </>
  ) : (
    <Text>No component config found.</Text>
  )
}

const LoadComponentDependencies: FC<{
  component: TComponent
  installId: string
  orgId: string
}> = async ({ component, installId, orgId }) => {
  const { data, error } = await nueQueryData<Array<TComponent>>({
    orgId,
    path: `components/${component?.id}/dependencies`,
  })

  return (
    <div className="flex items-center gap-4">
      {error ? (
        <Text>{error?.error}</Text>
      ) : (
        <ComponentDependencies
          deps={data}
          installId={installId}
          name={component?.name}
        />
      )}
    </div>
  )
}
