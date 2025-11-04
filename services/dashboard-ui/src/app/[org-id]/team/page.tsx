import type { Metadata } from 'next'
import { redirect } from 'next/navigation'
import { Suspense, type FC } from 'react'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Link } from '@/components/common/Link'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageSection } from '@/components/layout/PageSection'
import { Text } from '@/components/common/Text'
import { getOrgById } from '@/lib'
import { auth0 } from '@/lib/auth'
import type { TAccount, TInvite } from '@/types'
import { isNuonSession } from '@/utils/session-utils'

// NOTE: old layout stuff
import { ErrorBoundary as OldErrorBoundary } from 'react-error-boundary'
import {
  DashboardContent,
  ErrorFallback,
  Loading,
  OrgInviteModal,
  StatusBadge,
  Section,
  TeamMembersTable,
  OldText,
} from '@/components'
import { API_URL } from '@/configs/api'
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
    return org?.features?.['stratus-layout'] ? (
      <PageLayout
        breadcrumb={{
          baseCrumbs: [
            {
              path: `/${orgId}`,
              text: org?.name,
            },
            {
              path: `/${orgId}/installs`,
              text: 'Installs',
            },
          ],
        }}
        isScrollable
      >
        <PageHeader>
          <HeadingGroup>
            <Text variant="h3" weight="stronger" level={1}>
              Team
            </Text>
            <Text theme="neutral">Manage your organization team here.</Text>
          </HeadingGroup>
        </PageHeader>
        <PageContent>
          <PageSection className="border-t !pt-0">
            {/* old team component, needs updated */}

            <div className="flex-auto md:grid md:grid-cols-12 divide-x">
              <div className="divide-y flex flex-col flex-auto col-span-8">
                <Section heading="Members">
                  <OldErrorBoundary fallbackRender={ErrorFallback}>
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
                  </OldErrorBoundary>
                </Section>
              </div>
              <div className="divide-y flex flex-col flex-auto col-span-4">
                <Section heading="Invites">
                  <OldErrorBoundary fallbackRender={ErrorFallback}>
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
                  </OldErrorBoundary>
                </Section>
              </div>
            </div>

            {/* old team component, needs updated */}
          </PageSection>
        </PageContent>
      </PageLayout>
    ) : (
      <DashboardContent
        breadcrumb={[{ href: `/${orgId}`, text: 'Team' }]}
        heading={org?.name}
        headingUnderline={org?.id}
        statues={
          <div className="flex items-start gap-8">
            <span className="flex flex-col gap-2">
              <OldText className="text-cool-grey-600 dark:text-cool-grey-500">
                Status
              </OldText>
              <StatusBadge
                status={org?.status}
                description={org?.status_description}
                descriptionAlignment="right"
              />
            </span>
            <OrgInviteModal />
          </div>
        }
      >
        <div className="flex-auto md:grid md:grid-cols-12 divide-x">
          <div className="divide-y flex flex-col flex-auto col-span-8">
            <Section heading="Members">
              <OldErrorBoundary fallbackRender={ErrorFallback}>
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
              </OldErrorBoundary>
            </Section>
          </div>
          <div className="divide-y flex flex-col flex-auto col-span-4">
            <Section heading="Invites">
              <OldErrorBoundary fallbackRender={ErrorFallback}>
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
              </OldErrorBoundary>
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
  const session = await auth0.getSession()
  const members = await fetch(
    `${API_URL}/v1/orgs/current/accounts`,
    await getFetchOpts(orgId)
  )
    .then((res) => res.json() as Promise<Array<TAccount>>)
    .catch(console.error)

  return members ? (
    <TeamMembersTable
      members={
        isNuonSession(session?.user)
          ? members
          : members.filter((member) => !member?.email?.endsWith('nuon.co'))
      }
    />
  ) : (
    <OldText>No team members to show</OldText>
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
    <OldText>No invites to show</OldText>
  )
}
