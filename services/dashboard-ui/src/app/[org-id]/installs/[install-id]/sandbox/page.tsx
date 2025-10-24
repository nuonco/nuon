import type { Metadata } from 'next'
import { Suspense } from 'react'
import { CaretRightIcon, FileCodeIcon } from '@phosphor-icons/react/dist/ssr'
import { Link } from '@/components/common/Link'
import { PageSection } from '@/components/layout/PageSection'
import { getInstallById, getInstallDriftedObjects, getOrgById } from '@/lib'
import type { TPageProps } from '@/types'
import { SandboxConfig } from './config'
import { SandboxRuns } from './sandbox-runs'

// NOTE: old layout stuff
import { ErrorBoundary as OldErrorBoundary } from 'react-error-boundary'
import {
  DashboardContent,
  DeprovisionSandboxModal,
  ErrorFallback,
  InstallStatuses,
  InstallPageSubNav,
  InstallManagementDropdown,
  Link as OldLink,
  Loading,
  ReprovisionSandboxModal,
  Section,
  Text as OldText,
  Time,
} from '@/components'
import { DriftedBanner } from '@/components/old/DriftedBanner'
import { TerraformWorkspace } from '@/components/old/InstallSandbox'
import { SandboxManagementDropdown } from '@/components/old/InstallSandbox/ManagementDropdown'

type TInstallPageProps = TPageProps<'org-id' | 'install-id'>

export async function generateMetadata({
  params,
}: TInstallPageProps): Promise<Metadata> {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const { data: install } = await getInstallById({ installId, orgId })

  return {
    title: `Sandbox | ${install.name} | Nuon`,
  }
}

export default async function InstallSandboxPage({
  params,
  searchParams,
}: TInstallPageProps) {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const sp = await searchParams
  const [{ data: install }, { data: driftedObjects }, { data: org }] =
    await Promise.all([
      getInstallById({ installId, orgId }),
      getInstallDriftedObjects({ installId, orgId }),
      getOrgById({ orgId }),
    ])

  const latestSandboxRun = install?.install_sandbox_runs?.at(0)
  const driftedObject = driftedObjects?.find(
    (drifted) =>
      drifted?.['target_type'] === 'install_sandbox_run' &&
      drifted?.['target_id'] === latestSandboxRun?.id
  )

  return org?.features?.['stratus-layout'] ? (
    <PageSection isScrollable className="!p-0">
      {/* old layout stuff*/}

      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-y md:divide-x">
        <div className="md:col-span-8 divide-y flex-auto flex flex-col">
          {driftedObject ? (
            <Section className="!border-b-0 !pb-0">
              <DriftedBanner drifted={driftedObject} />
            </Section>
          ) : null}
          <Section
            actions={
              <OldText>
                <OldLink href={`/${orgId}/apps/${install.app_id}`}>
                  Details
                  <CaretRightIcon />
                </OldLink>
              </OldText>
            }
            className="flex-initial"
            heading="Config"
            childrenClassName="flex flex-col gap-4"
          >
            <OldErrorBoundary fallbackRender={ErrorFallback}>
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
            </OldErrorBoundary>
          </Section>

          <Section
            className="flex-initial"
            childrenClassName="flex flex-col gap-4"
          >
            <OldErrorBoundary fallbackRender={ErrorFallback}>
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
            </OldErrorBoundary>
          </Section>
        </div>

        <div className="divide-y flex flex-col md:col-span-4">
          <Section heading="Sandbox controls" className="flex-initial">
            <div className="flex items-center gap-4 flex-wrap">
              <SandboxManagementDropdown />
            </div>
          </Section>
          <Section heading="Sandbox history">
            <OldErrorBoundary fallbackRender={ErrorFallback}>
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
            </OldErrorBoundary>
          </Section>
        </div>
      </div>

      {/* old layout stuff*/}
    </PageSection>
  ) : (
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
              <OldText isMuted>Managed By</OldText>
              <OldText>
                <FileCodeIcon />
                Config File
              </OldText>
            </span>
          ) : null}
          <span className="flex flex-col gap-2">
            <OldText isMuted>App config</OldText>
            <OldText>
              <Link href={`/${orgId}/apps/${install.app_id}`}>
                {install?.app?.name}
              </Link>
            </OldText>
          </span>
          <InstallStatuses />

          <InstallManagementDropdown />
        </div>
      }
      meta={<InstallPageSubNav installId={installId} orgId={orgId} />}
    >
      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-y  md:divide-x">
        <div className="md:col-span-8 divide-y flex-auto flex flex-col">
          {driftedObject ? (
            <Section className="!border-b-0 !pb-0">
              <DriftedBanner drifted={driftedObject} />
            </Section>
          ) : null}
          <Section
            actions={
              <OldText>
                <OldLink href={`/${orgId}/apps/${install.app_id}`}>
                  Details
                  <CaretRightIcon />
                </OldLink>
              </OldText>
            }
            className="flex-initial"
            heading="Config"
            childrenClassName="flex flex-col gap-4"
          >
            <OldErrorBoundary fallbackRender={ErrorFallback}>
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
            </OldErrorBoundary>
          </Section>

          <Section
            className="flex-initial"
            childrenClassName="flex flex-col gap-4"
          >
            <OldErrorBoundary fallbackRender={ErrorFallback}>
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
            </OldErrorBoundary>
          </Section>
        </div>

        <div className="divide-y flex flex-col md:col-span-4">
          <Section heading="Sandbox controls" className="flex-initial">
            <div className="flex items-center gap-4 flex-wrap">
              <SandboxManagementDropdown />
            </div>
          </Section>
          <Section heading="Sandbox history">
            <OldErrorBoundary fallbackRender={ErrorFallback}>
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
            </OldErrorBoundary>
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
}
