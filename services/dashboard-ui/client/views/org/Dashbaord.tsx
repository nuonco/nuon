import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { AnnouncementsList } from '@/components/orgs/AnnouncementsList'
import { PendingApprovals } from '@/components/orgs/PendingApprovals'
import {
  RecentActivities,
  type IActivity,
} from '@/components/orgs/RecentActivities'
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
import { ActiveWorkflows } from '@/components/workflows/ActiveWorkflows'
import { useActiveWorkflows } from '@/hooks/use-active-workflows'
import { useOrg } from '@/hooks/use-org'
import { useWorkflowApprovals } from '@/hooks/use-workflow-approvals'
import { getOrgStats, getRunnerJobs } from '@/lib'
import {
  getJobHref,
  getJobName,
  getJobExecutionStatus,
} from '@/utils/runner-utils'
import announcementsData from '@/content/dashboard-announcements.json'

const SANDBOX_BANNER_DISMISSED_KEY = 'nuon:dismissed-sandbox-banner'

function isSandboxBannerDismissed(orgId?: string): boolean {
  if (!orgId) return false
  try {
    const ids: string[] = JSON.parse(
      localStorage.getItem(SANDBOX_BANNER_DISMISSED_KEY) || '[]'
    )
    return ids.includes(orgId)
  } catch {
    return false
  }
}

function persistSandboxBannerDismiss(orgId: string) {
  try {
    const ids: string[] = JSON.parse(
      localStorage.getItem(SANDBOX_BANNER_DISMISSED_KEY) || '[]'
    )
    if (!ids.includes(orgId)) {
      localStorage.setItem(
        SANDBOX_BANNER_DISMISSED_KEY,
        JSON.stringify([...ids, orgId])
      )
    }
  } catch {
    // ignore storage failures
  }
}

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
  const [searchParams] = useSearchParams()
  const offset = Number(searchParams.get('offset') ?? 0)
  const { org } = useOrg()
  const { approvals } = useWorkflowApprovals()
  const { activeWorkflows } = useActiveWorkflows()

  useEffect(() => {
    if (!org) return
    if (!org.features?.['org-dashboard']) {
      navigate(
        org.features?.['org-runner'] ? `/${org.id}/runner` : `/${org.id}/apps`
      )
    }
  }, [org])

  const [sandboxBannerDismissed, setSandboxBannerDismissed] = useState(() =>
    isSandboxBannerDismissed(org?.id)
  )

  useEffect(() => {
    setSandboxBannerDismissed(isSandboxBannerDismissed(org?.id))
  }, [org?.id])

  const runnerId = org?.runner_group?.runners?.at(0)?.id

  const { data: stats } = useQuery({
    queryKey: ['org-stats', org?.id],
    queryFn: () => getOrgStats({ orgId: org!.id }),
    enabled: !!org?.id,
  })

  const { data: jobs } = useQuery({
    queryKey: ['runner-jobs', org?.id, runnerId, offset],
    queryFn: () =>
      getRunnerJobs({
        orgId: org!.id,
        runnerId: runnerId!,
        groups: ['build', 'deploy', 'operations', 'sync'],
        limit: 10,
        offset,
      }),
    enabled: !!org?.id && !!runnerId,
    refetchInterval: 20_000,
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
    <PageLayout>
      <PageTitle title={`Dashboard | ${org?.name}`} />
      <Breadcrumbs breadcrumbs={[{ path: `/${org?.id}`, text: org?.name }]} />
      <PageHeader className="border-b">
        <HeadingGroup>
          <Text variant="h3" weight="stronger" level={1} className="mb-4">
            Welcome to {org?.name}!
          </Text>
          <Text theme="neutral">
            Manage your applications and deployed installs.
          </Text>
        </HeadingGroup>
      </PageHeader>
      {org?.sandbox_mode && !sandboxBannerDismissed && (
        <div className="px-6 pt-6">
          <Banner
            theme="warn"
            onDismiss={() => {
              if (!org?.id) return
              persistSandboxBannerDismiss(org.id)
              setSandboxBannerDismissed(true)
            }}
          >
            <div className="flex flex-col gap-1">
              <Text weight="strong">Sandbox mode</Text>
              <Text variant="subtext" theme="neutral">
                This organization is running in sandbox mode. Installs use
                simulated infrastructure instead of deploying to a real cloud
                account.
              </Text>
            </div>
          </Banner>
        </div>
      )}
      <PageContent>
        <PageGrid className="md:divide-x flex-auto !grid-cols-1 md:!grid-cols-[1fr_400px]">
          <PageSection className="flex-1 border-r !gap-12">
            <div className="flex flex-col gap-4">
              <Text variant="h3" weight="strong">
                Overview
              </Text>
              <StatsGrid
                stats={[
                  { label: 'Total installs', value: stats?.install_count ?? 0 },
                  {
                    label: 'Active applications',
                    value: stats?.app_count ?? 0,
                  },
                  { label: 'Active workflows', value: activeWorkflows.length },
                  { label: 'Pending approvals', value: approvals.length },
                ]}
              />
            </div>
            <PendingApprovals />
            <ActiveWorkflows workflows={activeWorkflows} />
            <div className="flex flex-col gap-4">
              <Text variant="base" weight="strong">
                Recent activities
              </Text>
              <RecentActivities
                activities={recentActivities}
                pagination={jobs?.pagination}
              />
            </div>
          </PageSection>
          <PageSection className="w-full">
            <AnnouncementsList announcements={announcementsData.announcements} />
          </PageSection>
        </PageGrid>
      </PageContent>
    </PageLayout>
  )
}
