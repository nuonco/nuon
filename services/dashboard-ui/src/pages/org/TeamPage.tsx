import { useParams, useSearchParams } from 'react-router-dom'
import { useOrg } from '@/hooks/use-org'
import { usePolling } from '@/hooks/use-polling'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { BackToTop } from '@/components/common/BackToTop'
import { TeamTable, TeamTableSkeleton } from '@/components/team/TeamTable'
import { InviteUserButton } from '@/components/team/InviteUserButton'
import { Badge } from '@/components/common/Badge'
import { Status } from '@/components/common/Status'
import { ResendOrgInviteButton } from '@/components/team/ResendOrgInvite'
import { EmptyState } from '@/components/common/EmptyState'
import type { TAccount, TOrgInvite } from '@/types'

export default function TeamPage() {
  const { orgId } = useParams()
  const { org } = useOrg()
  const [searchParams] = useSearchParams()
  const offset = searchParams.get('offset') || '0'
  
  const pageLimit = 20

  // Fetch team members
  const {
    data: membersResponse,
    isLoading: membersLoading,
    headers: membersHeaders,
  } = usePolling<TAccount[]>({
    path: `/api/ctl-api/v1/orgs/${orgId}/accounts?limit=${pageLimit}&offset=${offset}`,
    pollInterval: 30000,
    shouldPoll: true,
  })

  // Fetch pending invites
  const {
    data: invitesResponse,
    isLoading: invitesLoading,
  } = usePolling<TOrgInvite[]>({
    path: `/api/ctl-api/v1/orgs/${orgId}/invites`,
    pollInterval: 30000,
    shouldPoll: true,
  })

  const members = membersResponse || []
  const invites = invitesResponse || []
  
  const pagination = {
    limit: Number(membersHeaders?.['x-nuon-page-limit'] ?? pageLimit),
    hasNext: membersHeaders?.['x-nuon-page-next'] === 'true',
    offset: Number(membersHeaders?.['x-nuon-page-offset'] ?? '0'),
  }

  // Filter out Nuon employees for non-internal users
  const filteredMembers = members.filter(
    (member) => !member?.email?.endsWith('nuon.co')
  )

  // Filter pending invites (not accepted)
  const pendingInvites = invites.filter((i) => i?.status !== 'accepted')

  if (!org?.features?.['org-settings']) {
    return (
      <PageSection isScrollable>
        <Text theme="neutral">Team settings are not available for this organization.</Text>
      </PageSection>
    )
  }

  return (
    <PageSection isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${orgId}`, text: org?.name || '' },
          { path: `/${orgId}/team`, text: 'Team' },
        ]}
      />
      
      <div className="flex items-center justify-between">
        <HeadingGroup>
          <Text variant="h3" weight="stronger" level={1}>
            Team
          </Text>
          <Text theme="neutral">
            Manage your team members and permissions.
          </Text>
        </HeadingGroup>
        <InviteUserButton />
      </div>

      <div className="flex flex-col gap-8 mt-6">
        {/* Active Members Section */}
        <div className="flex flex-col gap-4">
          <Text variant="base" weight="strong">
            Active members
          </Text>
          {membersLoading ? (
            <TeamTableSkeleton />
          ) : (
            <TeamTable members={filteredMembers} pagination={pagination} />
          )}
        </div>

        {/* Pending Invites Section */}
        <div className="flex flex-col gap-4">
          <Text variant="base" weight="strong">
            Active invites
          </Text>
          {invitesLoading ? (
            <div className="flex flex-col gap-2">
              <div className="flex items-center gap-4 animate-pulse">
                <div className="h-6 w-20 bg-neutral-200 dark:bg-neutral-700 rounded" />
                <div className="h-4 w-32 bg-neutral-200 dark:bg-neutral-700 rounded" />
                <div className="h-5 w-16 bg-neutral-200 dark:bg-neutral-700 rounded" />
              </div>
            </div>
          ) : pendingInvites.length > 0 ? (
            <div className="flex flex-col gap-2">
              {pendingInvites.map((invite) => (
                <div className="flex items-center gap-4" key={invite?.id}>
                  <Status variant="badge" status={invite?.status} />
                  <Text variant="subtext">{invite?.email}</Text>
                  <Badge size="sm" variant="code">
                    {invite?.role_type === 'org_admin' ? 'Admin' : invite?.role_type}
                  </Badge>
                  <ResendOrgInviteButton invite={invite} size="sm" />
                </div>
              ))}
            </div>
          ) : (
            <EmptyState
              variant="table"
              title="No active invites"
              emptyMessage="No outstanding invites to this org"
            />
          )}
        </div>
      </div>

      <BackToTop />
    </PageSection>
  )
}
