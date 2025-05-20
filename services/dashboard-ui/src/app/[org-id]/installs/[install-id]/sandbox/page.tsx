import type { Metadata } from 'next'
import { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { CaretRight } from '@phosphor-icons/react/dist/ssr'
import {
  AppSandboxConfig,
  AppSandboxVariables,
  DashboardContent,
  DeprovisionSandboxModal,
  ErrorFallback,
  InstallStatuses,
  InstallPageSubNav,
  InstallManagementDropdown,
  Link,
  Loading,
  Notice,
  ReprovisionSandboxModal,
  SandboxHistory,
  Section,
  Text,
  Time,
  ClickToCopyButton,
  CodeViewer,
} from '@/components'
import {
  TerraformWorkspace,
  ValuesFileModal,
} from '@/components/InstallSandbox'
import {
  getInstall,
  getInstallSandboxRuns,
  getInstallSandboxRun,
  getRunnerJob,
  getOrg,
} from '@/lib'
import type { TAppConfig } from '@/types'
import { nueQueryData } from '@/utils'

export async function generateMetadata({ params }): Promise<Metadata> {
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string
  const install: any = await getInstall({ installId, orgId })

  return {
    title: `${install.name} | Sandbox`,
  }
}

export default withPageAuthRequired(async function InstallComponent({
  params,
}) {
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string
  const [install, org] = await Promise.all([
    getInstall({ installId, orgId }),
    getOrg({ orgId }),
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
          <span className="flex flex-col gap-2">
            <Text isMuted>App config</Text>
            <Text>
              <Link href={`/${orgId}/apps/${install.app_id}`}>
                {install?.app?.name}
              </Link>
            </Text>
          </span>
          <InstallStatuses initInstall={install} shouldPoll />

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
                  <CaretRight />
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
                <LoadSandboxConfig
                  appId={install?.app_id}
                  appConfigId={install?.app_config_id}
                  orgId={orgId}
                />
              </Suspense>
            </ErrorBoundary>
            {org?.features?.['terraform-workspace'] || (
              <ErrorBoundary fallbackRender={ErrorFallback}>
                <Suspense
                  fallback={
                    <Loading
                      variant="stack"
                      loadingText="Loading latest sandbox outputs..."
                    />
                  }
                >
                  <LoadLatestOutputs
                    installId={installId}
                    orgId={orgId}
                    installSandboxRunId={
                      install?.install_sandbox_runs?.at(0)?.id
                    }
                  />
                </Suspense>
              </ErrorBoundary>
            )}
          </Section>

          {org?.features?.['terraform-workspace'] && (
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
                    workspace={install.sandbox.terraform_workspace}
                  />
                </Suspense>
              </ErrorBoundary>
            </Section>
          )}
        </div>

        <div className="divide-y flex flex-col md:col-span-4">
          <Section heading="Sandbox controls" className="flex-initial">
            <div className="flex items-center gap-4">
              <ReprovisionSandboxModal installId={installId} orgId={orgId} />
              <DeprovisionSandboxModal install={install} />
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
                <LoadSandboxHistory installId={installId} orgId={orgId} />
              </Suspense>
            </ErrorBoundary>
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
})

const LoadSandboxConfig: FC<{
  appId: string
  appConfigId: string
  orgId: string
}> = async ({ appId, appConfigId, orgId }) => {
  const { data, error } = await nueQueryData<TAppConfig>({
    orgId,
    path: `apps/${appId}/config/${appConfigId}?recurse=true`,
  })

  return error ? (
    <Notice>{error?.error}</Notice>
  ) : (
    <>
      <AppSandboxConfig sandboxConfig={data?.sandbox} />
      <AppSandboxVariables
        variables={data?.sandbox?.variables}
        isNotTruncated
      />
      <ValuesFileModal valuesFiles={data?.sandbox?.variables_files} />
    </>
  )
}

const LoadSandboxHistory: FC<{
  installId: string
  orgId: string
}> = async ({ installId, orgId }) => {
  const sandboxRuns = await getInstallSandboxRuns({
    installId,
    orgId,
  }).catch(console.error)

  return sandboxRuns ? (
    <SandboxHistory
      installId={installId}
      orgId={orgId}
      initSandboxRuns={sandboxRuns}
      shouldPoll
    />
  ) : (
    <Text>Unable to load sandbox history.</Text>
  )
}

const LoadLatestOutputs: FC<{
  installSandboxRunId: string
  installId: string
  orgId: string
}> = async ({ installId, orgId, installSandboxRunId }) => {
  const sandboxRun = await getInstallSandboxRun({
    installId,
    orgId,
    installSandboxRunId,
  })
  const runnerJob = await getRunnerJob({
    orgId,
    runnerJobId: sandboxRun?.runner_job?.id,
  }).catch(console.error)

  return runnerJob ? (
    <div className="flex flex-col gap-2">
      <div className="flex items-center justify-between">
        <Text variant="med-12">Outputs</Text>
        <ClickToCopyButton textToCopy={JSON.stringify(runnerJob.outputs)} />
      </div>
      <CodeViewer
        initCodeSource={JSON.stringify(runnerJob.outputs, null, 2)}
        language="json"
      />
    </div>
  ) : null
}
