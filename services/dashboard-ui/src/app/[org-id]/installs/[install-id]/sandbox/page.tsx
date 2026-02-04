import type { Metadata } from 'next'
import { AsyncBoundary } from '@/components/common/AsyncBoundary'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { ManagementDropdown } from '@/components/sandbox/management/ManagementDropdown'
import { getInstall, getInstallDriftedObjects, getOrg } from '@/lib'
import type { TPageProps } from '@/types'
import { Runs, RunsError, RunsSkeleton } from './runs'

// NOTE: old layout stuff
import { Loading, Section } from '@/components'
import { DriftedBanner } from '@/components/old/DriftedBanner'
import { TerraformWorkspace } from '@/components/old/InstallSandbox'
import { SandboxConfig } from './config'

type TInstallPageProps = TPageProps<'org-id' | 'install-id'>

export async function generateMetadata({
  params,
}: TInstallPageProps): Promise<Metadata> {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const { data: install } = await getInstall({ installId, orgId })

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
      getInstall({ installId, orgId }),
      getInstallDriftedObjects({ installId, orgId }),
      getOrg({ orgId }),
    ])

  const latestSandboxRun = install?.install_sandbox_runs?.at(0)
  const driftedObject = driftedObjects?.find(
    (drifted) =>
      drifted?.['target_type'] === 'install_sandbox_run' &&
      drifted?.['target_id'] === latestSandboxRun?.id
  )

  return (
    <PageSection isScrollable className="!p-0">
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${orgId}`,
            text: org?.name,
          },
          {
            path: `/${orgId}/installs`,
            text: 'Installs',
          },
          {
            path: `/${orgId}/installs/${installId}`,
            text: install?.name,
          },
          {
            path: `/${orgId}/installs/${installId}/sandbox`,
            text: 'Sandbox',
          },
        ]}
      />
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
              <Text variant="subtext">
                <Link href={`/${orgId}/apps/${install.app_id}`}>
                  Details
                  <Icon variant="CaretRightIcon" />
                </Link>
              </Text>
            }
            className="flex-initial"
            heading="Config"
            childrenClassName="flex flex-col gap-4"
          >
            <AsyncBoundary
              loadingFallback={
                <Loading
                  loadingText="Loading sandbox config..."
                  variant="stack"
                />
              }
              errorFallback={
                <span className="text-md">Unable to load sandbox config</span>
              }
            >
              <SandboxConfig
                appId={install?.app_id}
                appConfigId={install?.app_config_id}
                orgId={orgId}
              />
            </AsyncBoundary>
          </Section>

          <Section
            className="flex-initial"
            childrenClassName="flex flex-col gap-4"
          >
            <AsyncBoundary
              loadingFallback={
                <Loading
                  loadingText="Loading latest Terraform workspace..."
                  variant="stack"
                />
              }
              errorFallback={
                <span className="text-md">
                  Unable to load Terraform workspace
                </span>
              }
            >
              <TerraformWorkspace
                orgId={orgId}
                workspace={install?.sandbox?.terraform_workspace}
              />
            </AsyncBoundary>
          </Section>
        </div>

        <div className="divide-y flex flex-col md:col-span-4">
          <Section heading="Sandbox controls" className="flex-initial">
            <div className="flex items-center gap-4 flex-wrap">
              <ManagementDropdown />
            </div>
          </Section>
          <Section heading="Sandbox history">
            <AsyncBoundary
              loadingFallback={<RunsSkeleton />}
              errorFallback={<RunsError />}
            >
              <Runs
                installId={installId}
                orgId={orgId}
                offset={sp['offset'] || '0'}
              />
            </AsyncBoundary>
          </Section>
        </div>
      </div>

      {/* old layout stuff*/}
    </PageSection>
  )
}
