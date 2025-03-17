import type { Metadata } from 'next'
import { redirect } from 'next/navigation'
import { Suspense, type FC } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  DashboardContent,
  ErrorFallback,
  Loading,
  StatusBadge,
  Section,
  Text,
} from '@/components'
import { getOrg } from '@/lib'
import type { TAccount } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export async function generateMetadata({ params }): Promise<Metadata> {
  const orgId = params?.['org-id'] as string
  const org = await getOrg({ orgId })

  return {
    title: `${org.name} | Team`,
  }
}

export default withPageAuthRequired(async function OrgTeam({ params }) {
  const orgId = params?.['org-id'] as string
  const org = await getOrg({ orgId })

  if (org?.features?.['org-settings']) {
    return (
      <DashboardContent
        breadcrumb={[{ href: `/${orgId}`, text: 'Team' }]}
        heading={org?.name}
        headingUnderline={org?.id}
        statues={
          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Status
            </Text>
            <StatusBadge
              status={org?.status}
              description={org?.status_description}
              descriptionAlignment="right"
              shouldPoll
            />
          </span>
        }
      >
        <div className="flex-auto md:grid md:grid-cols-12 divide-x">
          <div className="divide-y flex flex-col flex-auto col-span-8">
            <Section heading="Members">
              <ErrorBoundary fallbackRender={ErrorFallback}>
                <Suspense
                  fallback={<Loading loadingText="Loading org members..." />}
                >
                  <OrgMembers orgId={orgId} />
                </Suspense>
              </ErrorBoundary>
            </Section>
          </div>
          <div className="divide-y flex flex-col flex-auto col-span-4">
            <Section heading="Invites">
              <Text variant="reg-12">TKTK</Text>
            </Section>
          </div>
        </div>
      </DashboardContent>
    )
  } else {
    redirect(`/${orgId}/apps`)
  }
})

const OrgMembers: FC<{ orgId: string }> = async ({ orgId }) => {
  const members = await fetch(
    `${API_URL}/v1/orgs/current/accounts`,
    await getFetchOpts(orgId)
  )
    .then((res) => res.json() as Promise<Array<TAccount>>)
    .catch(console.error)

  return (
    <div className="flex flex-col gap-2">
      {members && members?.length ? (
        members?.map((member) => <span key={member?.id}>{member?.email}</span>)
      ) : (
        <span>No members in this org</span>
      )}
    </div>
  )
}
