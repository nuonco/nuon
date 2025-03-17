import type { Metadata } from 'next'
import Image from 'next/image'
import { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  AppSandboxConfig,
  AppSandboxVariables,
  Config,
  ConfigContent,
  DashboardContent,
  Duration,
  ErrorFallback,
  ID,
  InstallCloudPlatform,
  InstallInputs,
  InstallInputsModal,
  InstallPageSubNav,
  InstallStatuses,
  Loading,
  Notice,
  StatusBadge,
  Section,
  SectionHeader,
  Text,
  Time,
  Markdown,
} from '@/components'
import {
  InstallManagementDropdown,
  BreakGlassForm,
} from '@/components/Installs'
import {
  getInstall,
  getInstallCurrentInputs,
  getInstallReadme,
  getInstallRunnerGroup,
  getRunnerLatestHeartbeat,
} from '@/lib'
import { RUNNERS, USER_REPROVISION } from '@/utils'

export async function generateMetadata({ params }): Promise<Metadata> {
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string
  const install = await getInstall({ installId, orgId })

  return {
    title: `${install.name} | Generate break glass stack`,
  }
}

export default withPageAuthRequired(async function InstallBreakGlass({
  params,
}) {
  const orgId = params?.['org-id'] as string
  const installId = params?.['install-id'] as string
  const install = await getInstall({ installId, orgId })

  return (
    <DashboardContent
      breadcrumb={[
        {
          href: `/${orgId}/installs`,
          text: 'Installs',
        },
        { href: `/${orgId}/installs/${installId}`, text: install.name },
        {
          href: `/${orgId}/installs/${installId}/break-glass`,
          text: 'Access permissions',
        },
      ]}
    >
      <>
        <header className="px-6 py-8 flex flex-col border-b">
          <hgroup className="flex flex-col gap-2">
            <Text level={1} role="heading" variant="semi-18">
              Break glass permissions
            </Text>
            <Text
              variant="reg-12"
              className="text-cool-grey-600 dark:text-white/70"
            >
              Develop robust CloudFormation templates to streamline AWS
              infrastructure deployment for clients.
            </Text>
          </hgroup>
        </header>

        <Section
          heading={
            <div className="flex flex-col gap-2">
              <Text variant="semi-18">
                <Image
                  className=""
                  src={`/aws-cloudformation.svg`}
                  alt=""
                  height={24}
                  width={24}
                />
                CloudFormation
              </Text>
              <Text
                variant="reg-14"
                className="text-cool-grey-600 dark:text-white/70"
              >
                Note: Review access permissions thorougly before implementing
                modifications.
              </Text>
            </div>
          }
        >
          <BreakGlassForm install={install} />
        </Section>
      </>
    </DashboardContent>
  )
})
