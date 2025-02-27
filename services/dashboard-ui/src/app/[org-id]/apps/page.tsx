import { Suspense, type FC } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  DashboardContent,
  ErrorFallback,
  NoApps,
  Loading,
  OrgAppsTable,
  Section,
} from '@/components'
import { getApps, getOrg } from '@/lib'
// TODO(nnnat): move segment init script to org dashboard
import { SegmentAnalyticsSetOrg } from '@/utils'

export default withPageAuthRequired(async function Apps({ params }) {
  const orgId = params?.['org-id'] as string
  const org = await getOrg({ orgId })

  return (
    <>
      {process.env.SEGMENT_WRITE_KEY && <SegmentAnalyticsSetOrg org={org} />}
      <DashboardContent breadcrumb={[{ href: `/${orgId}/apps`, text: 'Apps' }]}>
        <Section>
          <ErrorBoundary fallbackRender={ErrorFallback}>
            <Suspense
              fallback={
                <Loading variant="page" loadingText="Loading apps..." />
              }
            >
              <LoadApps orgId={orgId} />
            </Suspense>
          </ErrorBoundary>
        </Section>
      </DashboardContent>
    </>
  )
})

const LoadApps: FC<{ orgId: string }> = async ({ orgId }) => {
  const apps = await getApps({ orgId })
  return apps?.length ? <OrgAppsTable apps={apps} orgId={orgId} /> : <NoApps />
}
