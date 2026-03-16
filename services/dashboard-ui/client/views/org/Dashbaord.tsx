import { useEffect } from 'react'
import { useNavigate } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { AnnouncementCard } from '@/components/orgs/AnnouncementCard'
import { RecentActivities, type IActivity } from '@/components/orgs/RecentActivities'
import { StatsGrid } from '@/components/orgs/StatsGrid'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageContent } from '@/components/layout/PageContent'
import { PageGrid } from '@/components/layout/PageGrid'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useOrg } from '@/hooks/use-org'
import { getOrgStats, getRunnerJobs } from '@/lib'
import { getJobHref, getJobName, getJobExecutionStatus } from '@/utils/runner-utils'
import announcementsData from '@/content/dashboard-announcements.json'

function formatDuration(ms: number): string {
  const seconds = Math.floor(ms / 1000)
  if (seconds < 60) return `${seconds}s`
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m`
  const hours = Math.floor(minutes / 60)
  return `${hours}h ${minutes % 60}m`
}

export const Dashboard = () => {
  const navigate = useNavigate()
  const { org } = useOrg()

  useEffect(() => {
    if (!org) return
    if (!org.features?.['org-dashboard']) {
      navigate(org.features?.['org-runner'] ? `/${org.id}/runner` : `/${org.id}/apps`)
    }
  }, [org])

  const runnerId = org?.runner_group?.runners?.at(0)?.id

  const { data: stats } = useQuery({
    queryKey: ['org-stats', org?.id],
    queryFn: () => getOrgStats({ orgId: org!.id }),
    enabled: !!org?.id,
  })

  const { data: jobs } = useQuery({
    queryKey: ['runner-jobs', org?.id, runnerId],
    queryFn: () =>
      getRunnerJobs({
        orgId: org!.id,
        runnerId: runnerId!,
        groups: ['deploy', 'operations', 'sync'],
        limit: 10,
      }),
    enabled: !!org?.id && !!runnerId,
  })

  const recentActivities: IActivity[] = (jobs?.data ?? []).map((job) => ({
    id: job.id,
    installName: getJobName(job),
    installId: job.metadata?.install_id || '',
    message: getJobExecutionStatus(job),
    status: job.status,
    created_at: job.created_at,
    duration:
      job.finished_at && job.started_at
        ? formatDuration(
            new Date(job.finished_at).getTime() -
              new Date(job.started_at).getTime()
          )
        : undefined,
    triggeredBy: '',
    href: getJobHref(job),
  }))

  return (
    <PageLayout isScrollable>
      <PageTitle title={`Dashboard | ${org?.name}`} />
      <Breadcrumbs
        breadcrumbs={[{ path: `/${org?.id}`, text: org?.name }]}
      />
      <PageHeader>
        <HeadingGroup>
          <Text variant="h3" weight="stronger" level={1}>
            Welcome to {org?.name}!
          </Text>
          <Text theme="neutral">
            Manage your applications and deployed installs.
          </Text>
        </HeadingGroup>
      </PageHeader>
      <PageContent>
        <PageGrid className="md:divide-x flex-auto !grid-cols-1 md:!grid-cols-[1fr_400px]">
          <PageSection className="flex-1 border-r">
            <Text variant="h3" weight="strong">
              Overview
            </Text>
            <StatsGrid
              stats={[
                { label: 'Total Installs', value: stats?.install_count ?? 0 },
                { label: 'Active Applications', value: stats?.app_count ?? 0 },
                { label: 'Active Runner', value: 0 },
                { label: 'Active Installs', value: stats?.install_count ?? 0 },
              ]}
            />
            <Text variant="h3" weight="strong" className="mt-6">
              Recent activities
            </Text>
            <RecentActivities
              activities={recentActivities}
              orgId={org?.id ?? ''}
            />
          </PageSection>
          <PageSection className="w-full">
            <div className="flex flex-col gap-6">
              {announcementsData.announcements.map((announcement) => (
                <AnnouncementCard
                  key={announcement.id}
                  announcement={announcement}
                />
              ))}
            </div>
          </PageSection>
        </PageGrid>
      </PageContent>
    </PageLayout>
  )
}
