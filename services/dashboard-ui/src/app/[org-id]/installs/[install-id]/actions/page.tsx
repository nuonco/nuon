import type { Metadata } from 'next'
import { Suspense } from 'react'
import { FileCodeIcon } from '@phosphor-icons/react/dist/ssr'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Link } from '@/components/common/Link'
import { PageSection } from '@/components/layout/PageSection'
import { Text } from '@/components/common/Text'
import { getInstallById, getOrgById } from '@/lib'
import type { TPageProps } from '@/types'
import { InstallActions } from './actions'

// NOTE: old layout stuff
import { ErrorBoundary as OldErrorBoundary } from 'react-error-boundary'
import {
  Link as OldLink,
  InstallManagementDropdown,
  InstallPageSubNav,
  InstallStatuses,
  DashboardContent,
  ErrorFallback,
  Loading,
  Section,
  Text as OldText,
  Time,
} from '@/components'

type TInstallPageProps = TPageProps<'org-id' | 'install-id'>

export async function generateMetadata({
  params,
}: TInstallPageProps): Promise<Metadata> {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const { data: install } = await getInstallById({ installId, orgId })

  return {
    title: `Actions | ${install.name} | Nuon`,
  }
}

export default async function InstallActionsPage({
  params,
  searchParams,
}: TInstallPageProps) {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const sp = await searchParams
  const [{ data: install }, { data: org }] = await Promise.all([
    getInstallById({ installId, orgId }),
    getOrgById({ orgId }),
  ])

  return org?.features?.['stratus-layout'] ? (
    <PageSection isScrollable>
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Actions
        </Text>
        <Text theme="neutral">
          View and manage all actions for this install.
        </Text>
      </HeadingGroup>
      <ErrorBoundary
        fallback={
          <Text>
            An error loading your install components, please refresh the page
            and try again.
          </Text>
        }
      >
        {/* old page stuff */}
        <Suspense
          fallback={<Loading variant="page" loadingText="Loading actions..." />}
        >
          <InstallActions
            installId={installId}
            orgId={orgId}
            offset={sp['offset'] || '0'}
            q={sp['q'] || ''}
            trigger_types={sp['trigger_types'] || ''}
          />
        </Suspense>
        {/* old page stuff */}
      </ErrorBoundary>
    </PageSection>
  ) : (
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
      <Section childrenClassName="flex flex-auto">
        <OldErrorBoundary fallbackRender={ErrorFallback}>
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
        </OldErrorBoundary>
      </Section>
    </DashboardContent>
  )
}
