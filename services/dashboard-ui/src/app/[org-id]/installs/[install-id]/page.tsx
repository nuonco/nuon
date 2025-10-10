import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { Suspense } from 'react'
import { FileCodeIcon } from '@phosphor-icons/react/dist/ssr'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Link } from '@/components/common/Link'
import { PageSection } from '@/components/layout/PageSection'
import { Text } from '@/components/common/Text'
import { getInstallById, getOrgById } from '@/lib'
import { CurrentInputs } from './inputs'
import { Readme } from './readme'

// NOTE: old install components
import { ErrorBoundary } from 'react-error-boundary'
import {
  DashboardContent,
  ErrorFallback,
  InstallPageSubNav,
  InstallStatuses,
  InstallManagementDropdown,
  Link as OldLink,
  Loading,
  Section,
  Text as OldText,
  Time,
} from '@/components'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const { data: install } = await getInstallById({ installId, orgId })

  return {
    title: `Overview | ${install.name} | Nuon`,
  }
}

export default async function Install({ params }) {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const [{ data: install, error, status }, { data: org }] = await Promise.all([
    getInstallById({ installId, orgId }),
    getOrgById({ orgId }),
  ])

  if (error) {
    if (status === 404) {
      notFound()
    } else {
      notFound()
    }
  }

  return org?.features?.['stratus-layout'] ? (
    <PageSection className="!pt-0" isScrollable>
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
    </PageSection>
  ) : (
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
