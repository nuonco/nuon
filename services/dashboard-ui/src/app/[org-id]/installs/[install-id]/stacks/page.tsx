import type { Metadata } from 'next'
import { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { FileCodeIcon } from '@phosphor-icons/react/dist/ssr'
import {
  DashboardContent,
  ErrorFallback,
  InstallStatuses,
  InstallPageSubNav,
  InstallManagementDropdown,
  Link,
  Loading,
  Notice,
  Text,
  Time,
  StacksTable,
} from '@/components'
import { getInstall } from '@/lib'
import type { TInstallStack } from '@/types'
import { nueQueryData } from '@/utils'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const install: any = await getInstall({ installId, orgId })

  return {
    title: `${install.name} | Stacks`,
  }
}

export default async function InstallStack({ params }) {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const install = await getInstall({ installId, orgId })

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
      <section className="px-6 py-8">
        <ErrorBoundary fallbackRender={ErrorFallback}>
          <Suspense
            fallback={
              <Loading loadingText="Loading components..." variant="page" />
            }
          >
            <LoadInstallStacks installId={install?.id} orgId={orgId} />
          </Suspense>
        </ErrorBoundary>
      </section>
    </DashboardContent>
  )
}

const LoadInstallStacks: FC<{
  installId: string
  orgId: string
}> = async ({ installId, orgId }) => {
  const { data, error } = await nueQueryData<TInstallStack>({
    orgId,
    path: `installs/${installId}/stack`,
  })

  return error ? (
    <Notice>Can&apos;t load install stacks: {error?.error}</Notice>
  ) : data?.versions?.length ? (
    <StacksTable stack={data} />
  ) : (
    <Text>No install stacks.</Text>
  )
}
