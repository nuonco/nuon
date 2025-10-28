import type { Metadata } from 'next'
import { Suspense } from 'react'
import { FileCodeIcon } from '@phosphor-icons/react/dist/ssr'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { PageSection } from '@/components/layout/PageSection'
import { Text } from '@/components/common/Text'
import { getInstallById, getOrgById } from '@/lib'
import { TPageProps } from '@/types'
import { InstallStacksTable, InstallStacksTableSkeleton } from './stacks-table'

// NOTE: old layout stuff
import { ErrorBoundary as OldErrorBoundary } from 'react-error-boundary'
import {
  DashboardContent,
  ErrorFallback,
  InstallStatuses,
  InstallPageSubNav,
  InstallManagementDropdown,
  Link as OldLink,
  Loading,
  Text as OldText,
  Time,
} from '@/components'
import { Stacks } from './stacks'

type TInstallPageProps = TPageProps<'org-id' | 'install-id'>

export async function generateMetadata({
  params,
}: TInstallPageProps): Promise<Metadata> {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const { data: install }: any = await getInstallById({ installId, orgId })

  return {
    title: `Stacks | ${install.name} | Nuon`,
  }
}

export default async function InstallStack({ params }: TInstallPageProps) {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const [{ data: install }, { data: org }] = await Promise.all([
    getInstallById({ installId, orgId }),
    getOrgById({
      orgId,
    }),
  ])

  return org?.features?.['stratus-layout'] ? (
    <PageSection isScrollable>
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Install stacks
        </Text>
        <Text variant="subtext" theme="neutral">
          View your install stack versions below.
        </Text>
      </HeadingGroup>

      <OldErrorBoundary fallbackRender={ErrorFallback}>
        <Suspense fallback={<InstallStacksTableSkeleton />}>
          <InstallStacksTable installId={install?.id} orgId={orgId} />
        </Suspense>
      </OldErrorBoundary>
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
              <OldLink href={`/${orgId}/apps/${install.app_id}`}>
                {install?.app?.name}
              </OldLink>
            </OldText>
          </span>
          <InstallStatuses />

          <InstallManagementDropdown />
        </div>
      }
      meta={<InstallPageSubNav installId={installId} orgId={orgId} />}
    >
      <section className="px-6 py-8">
        <OldErrorBoundary fallbackRender={ErrorFallback}>
          <Suspense
            fallback={
              <Loading loadingText="Loading components..." variant="page" />
            }
          >
            <Stacks installId={install?.id} orgId={orgId} />
          </Suspense>
        </OldErrorBoundary>
      </section>
    </DashboardContent>
  )
}
