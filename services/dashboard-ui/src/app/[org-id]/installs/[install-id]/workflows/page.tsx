import type { Metadata } from 'next'
import { Suspense } from 'react'
import { FileCodeIcon } from '@phosphor-icons/react/dist/ssr'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Link } from '@/components/common/Link'
import { PageSection } from '@/components/layout/PageSection'
import { Text } from '@/components/common/Text'
import type { TPageProps } from '@/types'
import { getInstallById, getOrgById } from '@/lib'
import { InstallWorkflows } from './install-workflows'

// NOTE: old layout stuff
import {
  DashboardContent,
  Link as OldLink,
  Loading,
  InstallPageSubNav,
  InstallStatuses,
  InstallManagementDropdown,
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
    title: `Workflows | ${install.name} | Nuon`,
  }
}

export default async function InstallWorkflowsPage({
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
          Workflows
        </Text>
      </HeadingGroup>
      <ErrorBoundary fallback={<Text>Unable to load install workflows.</Text>}>
        <Suspense
          fallback={
            <Loading loadingText="Loading install history..." variant="page" />
          }
        >
          <InstallWorkflows
            installId={installId}
            orgId={orgId}
            offset={sp['workflows'] || '0'}
          />
        </Suspense>
      </ErrorBoundary>
    </PageSection>
  ) : (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/installs`, text: 'Installs' },
        {
          href: `/${orgId}/installs/${install.id}`,
          text: install.name,
        },
        {
          href: `/${orgId}/installs/${install.id}/workflows`,
          text: 'Workflows',
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
      <div className="flex flex-col lg:flex-row flex-auto">
        <Section heading="Install workflows" className="overflow-auto">
          <Suspense
            fallback={
              <Loading
                loadingText="Loading install history..."
                variant="page"
              />
            }
          >
            <InstallWorkflows
              installId={installId}
              orgId={orgId}
              offset={sp['workflows'] || '0'}
            />
          </Suspense>
        </Section>
      </div>
    </DashboardContent>
  )
}
