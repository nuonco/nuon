import type { Metadata } from 'next'
import { redirect } from 'next/navigation'
import { Suspense, type FC } from 'react'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { AsyncBoundary } from '@/components/common/AsyncBoundary'
import { Text } from '@/components/common/Text'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { TeamTable, TeamTableSkeleton } from '@/components/team/TeamTable'
import { InviteUserButton } from '@/components/team/InviteUser'
import { getOrg, getOrgAccounts } from '@/lib'
import { getSession } from '@/lib/auth-server'
import { isNuonSession } from '@/utils/session-utils'
import {
  InvitedUser,
  InvitedUserError,
  InvitedUserSkeleton,
} from './invited-user'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId } = await params
  const { data: org } = await getOrg({ orgId })

  return {
    title: `Team | ${org.name} | Nuon`,
  }
}

export default async function OrgTeam({ params, searchParams }) {
  const sp = await searchParams
  const { ['org-id']: orgId } = await params
  const { data: org } = await getOrg({ orgId })

  const pageLimit = 20
  const currentOffset = parseInt(sp['offset'] || '0', 10)
  if (currentOffset > 0) {
    const { data: members } = await getOrgAccounts({
      orgId,
      limit: pageLimit,
      offset: sp['offset'],
    })
    // If no members at this offset, redirect to previous page
    if (!members || members.length === 0) {
      const previousOffset = Math.max(0, currentOffset - pageLimit)
      const params = new URLSearchParams()
      if (previousOffset > 0) {
        params.set('offset', previousOffset.toString())
      }
      const redirectUrl = `/${orgId}/team${params.toString() ? `?${params.toString()}` : ''}`
      redirect(redirectUrl)
    }
  }

  if (org?.features?.['org-settings']) {
    return (
      <PageLayout isScrollable>
        <Breadcrumbs
          breadcrumbs={[
            {
              path: `/${orgId}`,
              text: org?.name,
            },
            {
              path: `/${orgId}/team`,
              text: 'Team',
            },
          ]}
        />
        <PageHeader className="flex items-center justify-between">
          <HeadingGroup>
            <Text variant="h3" weight="stronger" level={1}>
              Team
            </Text>
            <Text theme="neutral">
              Manage your team members and permissions.
            </Text>
          </HeadingGroup>
          <InviteUserButton />
        </PageHeader>
        <PageContent>
          <PageSection>
            <div>
              <Text variant="base" weight="strong">
                Active memebers
              </Text>
              <ErrorBoundary
                fallback={<Text theme="error">Error loading team members</Text>}
              >
                <Suspense fallback={<TeamTableSkeleton />}>
                  <StratusOrgMembers
                    orgId={orgId}
                    offset={sp['offset'] || '0'}
                  />
                </Suspense>
              </ErrorBoundary>
            </div>

            <div className="flex flex-col gap-4">
              <Text variant="base" weight="strong">
                Active invites
              </Text>
              <AsyncBoundary
                loadingFallback={<InvitedUserSkeleton />}
                errorFallback={<InvitedUserError />}
              >
                <InvitedUser orgId={orgId} />
              </AsyncBoundary>
            </div>
          </PageSection>
        </PageContent>
      </PageLayout>
    )
  } else {
    redirect(`/${orgId}/apps`)
  }
}

// New Stratus team members component
const StratusOrgMembers: FC<{
  orgId: string
  limit?: number
  offset?: string
}> = async ({ orgId, limit = 20, offset }) => {
  const session = await getSession()
  const {
    data: members,
    error,
    headers,
  } = await getOrgAccounts({
    orgId,
    limit,
    offset,
  })

  const pagination = {
    limit: Number(headers?.['x-nuon-page-limit'] ?? limit),
    hasNext: headers?.['x-nuon-page-next'] === 'true',
    offset: Number(headers?.['x-nuon-page-offset'] ?? '0'),
  }

  if (error || !members) {
    return <Text theme="error">Failed to load team members</Text>
  }

  const filteredMembers = isNuonSession(session?.user)
    ? members
    : members.filter((member) => !member?.email?.endsWith('nuon.co'))

  return <TeamTable members={filteredMembers} pagination={pagination} />
}
