import type { Metadata } from 'next'
import { Suspense, type FC } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  DashboardContent,
  ErrorFallback,
  NoInstalls,
  OrgInstallsTable,
  Loading,
  Section,
} from '@/components'
import { getInstalls, getOrg } from '@/lib'

export async function generateMetadata({ params }): Promise<Metadata> {
  const orgId = params?.['org-id'] as string
  const org = await getOrg({ orgId })

  return {
    title: `${org.name} | Installs`,
  }
}

export default withPageAuthRequired(async function Installs({ params }) {
  const orgId = params?.['org-id'] as string
  return (
    <DashboardContent
      breadcrumb={[{ href: `/${orgId}/installs`, text: 'Installs' }]}
    >
      <Section>
        <ErrorBoundary fallbackRender={ErrorFallback}>
          <Suspense
            fallback={
              <Loading variant="page" loadingText="Loading installs..." />
            }
          >
            <LoadInstalls orgId={orgId} />
          </Suspense>
        </ErrorBoundary>
      </Section>
    </DashboardContent>
  )
})

const LoadInstalls: FC<{ orgId: string }> = async ({ orgId }) => {
  const installs = await getInstalls({ orgId })
  return installs?.length ? (
    <OrgInstallsTable orgId={orgId} installs={installs} />
  ) : (
    <NoInstalls />
  )
}
