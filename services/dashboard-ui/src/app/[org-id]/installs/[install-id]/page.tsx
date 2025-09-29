import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { FileCodeIcon } from '@phosphor-icons/react/dist/ssr'
import {
  DashboardContent,
  ErrorFallback,
  InstallPageSubNav,
  InstallStatuses,
  InstallManagementDropdown,
  Link,
  Loading,
  Section,
  Text,
  Time,
} from '@/components'
import { getInstallById } from '@/lib'
import { CurrentInputs } from './inputs'
import { Readme } from './readme'

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

          <InstallManagementDropdown />
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
              <Readme installId={installId} orgId={orgId} />
            </Suspense>
          </ErrorBoundary>
        </Section>

        <div className="divide-y flex flex-col col-span-4">
          <Section className="flex-initial">
            <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={<Loading loadingText="Loading install inputs..." />}
              >
                <CurrentInputs installId={installId} orgId={orgId} />
              </Suspense>
            </ErrorBoundary>
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
}
