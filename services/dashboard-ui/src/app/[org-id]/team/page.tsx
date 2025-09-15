import type { Metadata } from 'next'
import { redirect } from 'next/navigation'
import { Suspense, type FC } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import {
  DashboardContent,
  ErrorFallback,
  Loading,
  OrgInviteModal,
  StatusBadge,
  Section,
  TeamMembersTable,
  Text,
} from '@/components'
import { API_URL } from '@/configs/api'
import { getOrgById } from '@/lib'
import type { TAccount, TInvite } from '@/types'
import { getFetchOpts } from '@/utils'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId } = await params
  const { data: org } = await getOrgById({ orgId })

  return {
    title: `Team | ${org.name} | Nuon`,
  }
}

export default async function OrgTeam({ params }) {
  const { ['org-id']: orgId } = await params
  const { data: org } = await getOrgById({ orgId })

  if (org?.features?.['org-settings']) {
    return (
      <DashboardContent
        breadcrumb={[{ href: `/${orgId}`, text: 'Team' }]}
        heading={org?.name}
        headingUnderline={org?.id}
        statues={
          <div className="flex items-start gap-8">
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
            <OrgInviteModal />
          </div>
        }
      >
        <div className="flex-auto md:grid md:grid-cols-12 divide-x">
          <div className="divide-y flex flex-col flex-auto col-span-8">
            <Section heading="Members">
              <ErrorBoundary fallbackRender={ErrorFallback}>
                <Suspense
                  fallback={
                    <Loading
                      variant="stack"
                      loadingText="Loading org members..."
                    />
                  }
                >
                  <OrgMembers orgId={orgId} />
                </Suspense>
              </ErrorBoundary>
            </Section>
          </div>
          <div className="divide-y flex flex-col flex-auto col-span-4">
            <Section heading="Invites">
              <ErrorBoundary fallbackRender={ErrorFallback}>
                <Suspense
                  fallback={
                    <Loading
                      variant="stack"
                      loadingText="Loading org invites..."
                    />
                  }
                >
                  <OrgInvites orgId={orgId} />
                </Suspense>
              </ErrorBoundary>
            </Section>
          </div>
        </div>
      </DashboardContent>
    )
  } else {
    redirect(`/${orgId}/apps`)
  }
}

const OrgMembers: FC<{ orgId: string }> = async ({ orgId }) => {
  const members = await fetch(
    `${API_URL}/v1/orgs/current/accounts`,
    await getFetchOpts(orgId)
  )
    .then((res) => res.json() as Promise<Array<TAccount>>)
    .catch(console.error)

  return members ? (
    <TeamMembersTable members={members} />
  ) : (
    <Text>No team members to show</Text>
  )
}

const OrgInvites: FC<{ orgId: string }> = async ({ orgId }) => {
  const invites = await fetch(
    `${API_URL}/v1/orgs/current/invites`,
    await getFetchOpts(orgId)
  )
    .then((res) => res.json() as Promise<Array<TInvite>>)
    .catch(console.error)

  return invites && invites.length ? (
    <div className="flex flex-col divide-y">
      {invites.map((invite) => (
        <span className="text-sm py-2 flex items-center gap-2" key={invite.id}>
          <StatusBadge
            status={invite.status}
            isWithoutBorder
            isStatusTextHidden
          />{' '}
          {invite.email}
        </span>
      ))}
    </div>
  ) : (
    <Text>No invites to show</Text>
  )
}
