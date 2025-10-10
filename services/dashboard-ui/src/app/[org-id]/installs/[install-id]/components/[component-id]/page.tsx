import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { Suspense } from 'react'
import { BackLink } from '@/components/common/BackLink'
import { BackToTop } from '@/components/common/BackToTop'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'

import { getInstallById, getInstallComponentById, getOrgById } from '@/lib'
import type { TPageProps } from '@/types'
import { ComponentConfig } from './config'
import { ComponentDependencies } from './dependencies'
import { Deploys } from './deploys'
import { LatestOutputs } from './outputs'

// NOTE: old layout stuff
import { ErrorBoundary } from 'react-error-boundary'
import { CaretRightIcon } from '@phosphor-icons/react/dist/ssr'
import {
  DashboardContent,
  ErrorFallback,
  InstallComponentManagementDropdown,
  Link,
  Loading,
  Section,
  Text as OldText,
} from '@/components'
import { DriftedBanner } from '@/components/DriftedBanner'
import { TerraformWorkspace } from '@/components/InstallSandbox'

type TInstallPageProps = TPageProps<'org-id' | 'install-id' | 'component-id'>

export async function generateMetadata({
  params,
}: TInstallPageProps): Promise<Metadata> {
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

export default async function InstallComponentPage({
  params,
  searchParams,
}: TInstallPageProps) {
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
  const containerId = 'install-component-page'
  return org?.features?.['stratus-layout'] ? (
    <PageSection id={containerId} isScrollable className="!p-0 !gap-0">
      {/* old page layout */}

      <div className="p-6 border-b flex justify-between">
        <HeadingGroup>
          <BackLink className="mb-6" />
          <Text variant="base" weight="strong">
            {component?.name}
          </Text>
          <ID>{component.id}</ID>
        </HeadingGroup>

        <div>
          <InstallComponentManagementDropdown
            componentId={installComponent?.component_id}
            componentName={installComponent?.component?.name}
          />
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
        <div className="divide-y flex-auto flex flex-col md:col-span-8">
          {installComponent?.drifted_object ? (
            <Section className="!border-b-0">
              <DriftedBanner drifted={installComponent?.drifted_object} />
            </Section>
          ) : null}

          <Section
            actions={
              <OldText>
                <Link
                  href={`/${orgId}/apps/${component.app_id}/components/${component.id}`}
                >
                  Details
                  <CaretRightIcon />
                </Link>
              </OldText>
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
                <ComponentConfig
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
                  <LatestOutputs
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
                  <ComponentDependencies
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
      {/* old page layout */}
      <BackToTop containerId={containerId} />
    </PageSection>
  ) : (
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
          <InstallComponentManagementDropdown
            componentId={installComponent?.component_id}
            componentName={installComponent?.component?.name}
          />
        </div>
      }
    >
      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
        <div className="divide-y flex-auto flex flex-col md:col-span-8">
          {installComponent?.drifted_object ? (
            <Section className="!border-b-0">
              <DriftedBanner drifted={installComponent?.drifted_object} />
            </Section>
          ) : null}

          <Section
            actions={
              <OldText>
                <Link
                  href={`/${orgId}/apps/${component.app_id}/components/${component.id}`}
                >
                  Details
                  <CaretRightIcon />
                </Link>
              </OldText>
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
                <ComponentConfig
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
                  <LatestOutputs
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
                  <ComponentDependencies
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
