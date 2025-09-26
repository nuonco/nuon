import type { Metadata } from 'next'
import { Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { CaretRightIcon, FileCodeIcon } from '@phosphor-icons/react/dist/ssr'
import {
  DashboardContent,
  DeprovisionSandboxModal,
  ErrorFallback,
  InstallStatuses,
  InstallPageSubNav,
  InstallManagementDropdown,
  Link,
  Loading,
  ReprovisionSandboxModal,
  Section,
  Text,
  Time,
} from '@/components'
import { TerraformWorkspace } from '@/components/InstallSandbox'
import { getInstallById } from '@/lib'

import { SandboxConfig } from './config'
import { SandboxRuns } from './sandbox-runs'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const { data: install } = await getInstallById({ installId, orgId })

  return {
    title: `Sandbox | ${install.name} | Nuon`,
  }
}

export default async function InstallComponent({ params, searchParams }) {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const sp = await searchParams
  const { data: install } = await getInstallById({ installId, orgId })

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/installs`, text: 'Installs' },
        {
          href: `/${orgId}/installs/${install.id}/components`,
          text: install.name,
        },
        {
          href: `/${orgId}/installs/${install.id}/sandbox`,
          text: 'Sandbox',
        },
      ]}
      heading={install.name}
      headingUnderline={install.id}
      headingMeta={
        <>
          Last updated <Time time={install?.updated_at} format="relative" />
        </>
      }
      statues={
        <div className="flex items-start gap-8">
          {install?.metadata?.managed_by &&
          install?.metadata?.managed_by === 'nuon/cli/install-config' ? (
            <span className="flex flex-col gap-2">
              <Text isMuted>Managed By</Text>
              <Text>
                <FileCodeIcon />
                Config File
              </Text>
            </span>
          ) : null}
          <span className="flex flex-col gap-2">
            <Text isMuted>App config</Text>
            <Text>
              <Link href={`/${orgId}/apps/${install.app_id}`}>
                {install?.app?.name}
              </Link>
            </Text>
          </span>
          <InstallStatuses />

          <InstallManagementDropdown
            orgId={orgId}
            hasInstallComponents={Boolean(install?.install_components?.length)}
            install={install}
          />
        </div>
      }
      meta={<InstallPageSubNav installId={installId} orgId={orgId} />}
    >
      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-y  md:divide-x">
        <div className="md:col-span-8 divide-y flex-auto flex flex-col">
          <Section
            actions={
              <Text>
                <Link href={`/${orgId}/apps/${install.app_id}`}>
                  Details
                  <CaretRightIcon />
                </Link>
              </Text>
            }
            className="flex-initial"
            heading="Config"
            childrenClassName="flex flex-col gap-4"
          >
            <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={
                  <Loading
                    loadingText="Loading sandbox config..."
                    variant="stack"
                  />
                }
              >
                <SandboxConfig
                  appId={install?.app_id}
                  appConfigId={install?.app_config_id}
                  orgId={orgId}
                />
              </Suspense>
            </ErrorBoundary>
          </Section>

          <Section
            className="flex-initial"
            childrenClassName="flex flex-col gap-4"
          >
            <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={
                  <Section heading="Terraform state">
                    <Loading
                      variant="stack"
                      loadingText="Loading latest Terraform workspace..."
                    />
                  </Section>
                }
              >
                <TerraformWorkspace
                  orgId={orgId}
                  workspace={install?.sandbox?.terraform_workspace}
                />
              </Suspense>
            </ErrorBoundary>
          </Section>
        </div>

        <div className="divide-y flex flex-col md:col-span-4">
          <Section heading="Sandbox controls" className="flex-initial">
            <div className="flex items-center gap-4 flex-wrap">
              <ReprovisionSandboxModal />
              <DeprovisionSandboxModal />
            </div>
          </Section>
          <Section heading="Sandbox history">
            <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={
                  <Loading
                    loadingText="Loading sandbox history..."
                    variant="stack"
                  />
                }
              >
                <SandboxRuns
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
