import type { Metadata } from 'next'
import { Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { FileCodeIcon } from '@phosphor-icons/react/dist/ssr'
import {
  Link,
  InstallManagementDropdown,
  InstallPageSubNav,
  InstallStatuses,
  DashboardContent,
  ErrorFallback,
  Loading,
  Section,
  Text,
  Time,
} from '@/components'
import { getInstallById } from '@/lib'
import { InstallActions } from './actions'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const { data: install } = await getInstallById({ installId, orgId })

  return {
    title: `Actions | ${install.name} | Nuon`,
  }
}

export default async function InstallWorkflowRuns({ params, searchParams }) {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const sp = await searchParams
  const { data: install } = await getInstallById({ installId, orgId })

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/installs`, text: 'Installs' },
        { href: `/${orgId}/installs/${install.id}`, text: install.name },
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
      <Section childrenClassName="flex flex-auto">
        <ErrorBoundary fallbackRender={ErrorFallback}>
          <Suspense
            fallback={
              <Loading variant="page" loadingText="Loading actions..." />
            }
          >
            <InstallActions
              installId={installId}
              orgId={orgId}
              offset={sp['offset'] || '0'}
              q={sp['q'] || ''}
              trigger_types={sp['trigger_types'] || ''}
            />
          </Suspense>
        </ErrorBoundary>
      </Section>
    </DashboardContent>
  )
}
