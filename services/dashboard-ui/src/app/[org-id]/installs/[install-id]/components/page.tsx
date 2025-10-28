import type { Metadata } from 'next'
import { Suspense } from 'react'
import { FileCodeIcon } from '@phosphor-icons/react/dist/ssr'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { PageSection } from '@/components/layout/PageSection'
import { Text } from '@/components/common/Text'
import { InstallComponentsTableSkeleton } from "@/components/install-components/InstallComponentsTable";
import { getInstallById, getOrgById } from '@/lib'
import type { TPageProps } from '@/types'
import { InstallComponentsTable } from "./components-table";


// NOTE: old layout stuff
import { ErrorBoundary } from 'react-error-boundary'
import {
  DashboardContent,
  ErrorFallback,
  Link,
  InstallStatuses,
  InstallPageSubNav,
  InstallManagementDropdown,
  Loading,
  Text as OldText,
  Time,
} from '@/components'
import { InstallComponents } from './components'

type TInstallPageProps = TPageProps<'org-id' | 'install-id'>

export async function generateMetadata({
  params,
}: TInstallPageProps): Promise<Metadata> {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const { data: install } = await getInstallById({ installId, orgId })

  return {
    title: `Components | ${install.name} | Nuon`,
  }
}

export default async function InstallComponentsPage({
  params,
  searchParams,
}: TInstallPageProps) {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const sp = await searchParams
  const [{ data: install }, { data: org }] = await Promise.all([
    getInstallById({ orgId, installId }),
    getOrgById({ orgId }),
  ])

  return org?.features?.['stratus-layout'] ? (
    <PageSection isScrollable>
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Install components
        </Text>
        <Text theme="neutral">
          View and manage all components for this install.
        </Text>
      </HeadingGroup>

      {/* old layout stuff */}
      <ErrorBoundary fallbackRender={ErrorFallback}>
        <Suspense
          fallback={
            <Loading loadingText="Loading components..." variant="page" />
          }
        >
          <InstallComponentsTable
            install={install}
            installId={install?.id}
            orgId={orgId}
            offset={sp['offset'] || '0'}
            q={sp['q'] || ''}
            types={sp['types'] || ''}
          />
        </Suspense>
      </ErrorBoundary>
      {/* old layout stuff */}
    </PageSection>
  ) : (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/installs`, text: 'Installs' },
        {
          href: `/${orgId}/installs/${install?.id}`,
          text: install?.name,
        },
        {
          href: `/${orgId}/installs/${install?.id}/components`,
          text: 'Components',
        },
      ]}
      heading={install?.name}
      headingUnderline={install?.id}
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
              <Link href={`/${orgId}/apps/${install?.app_id}`}>
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
      <section className="px-6 py-8">
        <ErrorBoundary fallbackRender={ErrorFallback}>
          <Suspense
            fallback={
              <Loading loadingText="Loading components..." variant="page" />
            }
          >
            <InstallComponents
              install={install}
              installId={install?.id}
              orgId={orgId}
              offset={sp['offset'] || '0'}
              q={sp['q'] || ''}
              types={sp['types'] || ''}
            />
          </Suspense>
        </ErrorBoundary>
      </section>
    </DashboardContent>
  )
}
