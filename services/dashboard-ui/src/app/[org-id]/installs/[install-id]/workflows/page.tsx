import type { Metadata } from 'next'
import { type FC, Suspense } from 'react'
import { FileCodeIcon } from '@phosphor-icons/react/dist/ssr'
import {
  DashboardContent,
  Link,
  Loading,
  InstallHistory,
  InstallPageSubNav,
  InstallStatuses,
  InstallWorkflowHistory,
  InstallManagementDropdown,
  Pagination,
  Section,
  Text,
  Time,
} from '@/components'
import { getInstall, getInstallEvents, getOrg } from '@/lib'
import { TInstallWorkflow } from '@/types'
import { nueQueryData } from '@/utils'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const install = await getInstall({ installId, orgId })

  return {
    title: `${install.name} | Workflows`,
  }
}

export default async function Install({ params, searchParams }) {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const sp = await searchParams
  const [install, org] = await Promise.all([
    getInstall({ installId, orgId }),
    getOrg({ orgId }),
  ])

  return (
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
      <div className="flex flex-col lg:flex-row flex-auto">
        <Section heading="Install history" className="overflow-auto">
          <Suspense
            fallback={
              <Loading
                loadingText="Loading install history..."
                variant="page"
              />
            }
          >
            <LoadInstallWorkflows
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

const LoadInstallWorkflows: FC<{
  installId: string
  orgId: string
  offset?: string
  limit?: string
}> = async ({ installId, orgId, offset, limit = '20' }) => {
  const params = new URLSearchParams({ offset, limit }).toString()
  const {
    data: installWorkflows,
    error,
    headers,
  } = await nueQueryData<Array<TInstallWorkflow>>({
    orgId,
    path: `installs/${installId}/workflows${params ? '?' + params : params}`,
    headers: {
      'x-nuon-pagination-enabled': true,
    },
  })

  const pageData = {
    hasNext: headers.get('x-nuon-page-next') || 'false',
    offset: headers?.get('x-nuon-page-offset') || '0',
  }

  return installWorkflows ? (
    <div className="flex flex-col gap-4">
      <InstallWorkflowHistory installWorkflows={installWorkflows} shouldPoll />
      <Pagination
        param="workflows"
        pageData={pageData}
        position="center"
        limit={20}
      />
    </div>
  ) : (
    <Text>No install history yet.</Text>
  )
}

const LoadInstallHistory: FC<{ installId: string; orgId: string }> = async ({
  installId,
  orgId,
}) => {
  const events = await getInstallEvents({ installId, orgId })
  return (
    <InstallHistory
      initEvents={events}
      installId={installId}
      orgId={orgId}
      shouldPoll
    />
  )
}
