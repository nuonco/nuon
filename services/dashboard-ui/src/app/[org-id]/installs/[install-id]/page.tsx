import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { FileCodeIcon } from '@phosphor-icons/react/dist/ssr'
import {
  DashboardContent,
  ErrorFallback,
  InstallInputs,
  InstallInputsModal,
  InstallPageSubNav,
  InstallStatuses,
  InstallManagementDropdown,
  Link,
  Loading,
  Notice,
  Section,
  SectionHeader,
  Text,
  Time,
  Markdown,
} from '@/components'
import {
  getInstallById,
  getInstallCurrentInputs,
  getInstallReadme,
} from '@/lib'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const { data: install } = await getInstallById({ installId, orgId })

  return {
    title: `Overview | ${install.name} | Nuon`,
  }
}

export default async function Install({ params }) {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const {
    data: install,
    error,
    status,
  } = await getInstallById({ installId, orgId }).catch((err) => {
    console.error(err)
    notFound()
  })

  if (error) {
    if (status === 404) {
      notFound()
    } else {
      notFound()
    }
  }

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/installs`, text: 'Installs' },
        {
          href: `/${orgId}/installs/${install.id}`,
          text: install.name,
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
      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
        <Section
          heading="README"
          className="md:col-span-8 !p-0"
          headingClassName="px-6 pt-6"
          childrenClassName="overflow-auto px-6 pb-6"
        >
          <ErrorBoundary fallbackRender={ErrorFallback}>
            <Suspense
              fallback={
                <Loading
                  variant="stack"
                  loadingText="Loading install README..."
                />
              }
            >
              <LoadInstallReadme installId={installId} orgId={orgId} />
            </Suspense>
          </ErrorBoundary>
        </Section>

        <div className="divide-y flex flex-col col-span-4">
          <Section className="flex-initial">
            <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={<Loading loadingText="Loading install inputs..." />}
              >
                <LoadInstallCurrentInputs installId={installId} orgId={orgId} />
              </Suspense>
            </ErrorBoundary>
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
}

const LoadInstallReadme: FC<{ installId: string; orgId: string }> = async ({
  installId,
  orgId,
}) => {
  const installReadme = await getInstallReadme({ installId, orgId }).catch(
    console.error
  )

  return installReadme ? (
    <div className="flex flex-col gap-3">
      {installReadme?.warnings?.length
        ? installReadme?.warnings?.map((warn, i) => (
            <Notice key={i.toString()} variant="warn">
              {warn}
            </Notice>
          ))
        : null}
      <Markdown content={installReadme?.readme} />
    </div>
  ) : (
    <Text variant="reg-12">No install README found</Text>
  )
}

const LoadInstallCurrentInputs: FC<{
  installId: string
  orgId: string
}> = async ({ installId, orgId }) => {
  const currentInputs = await getInstallCurrentInputs({ installId, orgId })

  return (
    <>
      <SectionHeader
        actions={
          currentInputs?.redacted_values ? (
            <InstallInputsModal currentInputs={currentInputs} />
          ) : undefined
        }
        className="mb-4"
        heading="Current inputs"
      />
      {currentInputs?.redacted_values ? (
        <InstallInputs currentInputs={currentInputs} />
      ) : (
        <Text>No inputs configured.</Text>
      )}
    </>
  )
}
