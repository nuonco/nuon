// TODO(nnnat): remove once we have this API change on prod
// @ts-nocheck
import { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  AppSandboxConfig,
  AppSandboxVariables,
  DashboardContent,
  ErrorFallback,
  InstallCloudPlatform,
  InstallInputsSection,
  InstallPageSubNav,
  InstallStatuses,
  StatusBadge,
  Section,
  Text,
  Markdown,
} from '@/components'
import {
  getInstall,
  getInstallReadme,
  getInstallRunnerGroup,
  getOrg,
} from '@/lib'
import { RUNNERS } from '@/utils'

export default withPageAuthRequired(async function Install({ params }) {
  const orgId = params?.['org-id'] as string
  const installId = params?.['install-id'] as string
  const [install, runnerGroup, org] = await Promise.all([
    getInstall({ installId, orgId }),
    getInstallRunnerGroup({ installId, orgId }).catch(console.error),
    getOrg({ orgId }),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${org.id}/apps`, text: org.name },
        { href: `/${org.id}/installs`, text: 'Installs' },
        {
          href: `/${org.id}/installs/${install.id}`,
          text: install.name,
        },
      ]}
      heading={install.name}
      headingUnderline={install.id}
      statues={<InstallStatuses initInstall={install} shouldPoll />}
      meta={<InstallPageSubNav installId={installId} orgId={orgId} />}
    >
      <div className="flex flex-col lg:flex-row flex-auto">
        <Section heading="README" className="overflow-auto history">
          <ErrorBoundary

            fallbackRender={ErrorFallback}
          >
            <LoadInstallReadme installId={installId} orgId={orgId} />
          </ErrorBoundary>
        </Section>

        <div className="divide-y flex flex-col lg:w-[500px] border-l">
          {install?.install_inputs?.length &&
          install?.install_inputs.some(
            (input) => input.values || input?.redacted_values
          ) ? (
            <InstallInputsSection inputs={install.install_inputs} />
          ) : null}

          <Section className="flex-initial" heading="Active sandbox">
            <div className="flex flex-col gap-8">
              <AppSandboxConfig sandboxConfig={install?.app_sandbox_config} />
              <AppSandboxVariables
                variables={install?.app_sandbox_config?.variables}
              />
            </div>
          </Section>

          {RUNNERS && runnerGroup ? (
            <Section className="flex-initial" heading="Runner group">
              <div className="flex flex-col gap-4">
                <Text>{runnerGroup.runners?.length} runners in this group</Text>
                <div className="divide-y">
                  {runnerGroup.runners?.map((runner) => (
                    <div key={runner?.id} className="flex flex-col gap-2">
                      <StatusBadge status={runner?.status} />
                      <Text variant="med-12">{runner?.display_name}</Text>
                    </div>
                  ))}
                </div>
              </div>
            </Section>
          ) : null}

          <Section heading="Cloud platform">
            <InstallCloudPlatform install={install} />
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
})

const LoadInstallReadme: FC<{ installId: string; orgId: string }> = async ({
  installId,
  orgId,
}) => {
  const installReadme = await getInstallReadme({ installId, orgId })

  return (
    <Suspense fallback="Loading install readme...">
      <Markdown content={installReadme?.readme} />
    </Suspense>
  )
}



