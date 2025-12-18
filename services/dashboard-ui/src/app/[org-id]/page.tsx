import type { Metadata } from 'next'
import { redirect } from 'next/navigation'
import { CaretRightIcon } from '@phosphor-icons/react/dist/ssr'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { AnnouncementCard } from '@/components/dashboard/AnnouncementCard'
import { RecentActivities } from '@/components/dashboard/RecentActivities'
import { StatsGrid } from '@/components/dashboard/StatsCard'
import { PageContent } from '@/components/layout/PageContent'
import { PageGrid } from '@/components/layout/PageGrid'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { getOrg, getOrgStats, getRunnerJobs } from '@/lib'
import { auth0 } from '@/lib/auth'
import type { TPageProps } from '@/types'
import announcementsData from '@/content/dashboard-announcements.json'

import {
  DashboardContent,
  Link as OldLink,
  StatusBadge,
  Section,
  Text as OldText,
} from '@/components'
import type { TRunnerJob } from '@/types'
import {
  getJobName,
  getJobExecutionStatus,
  getJobHref,
} from '@/utils/runner-utils'

// Helper to format duration
function formatDuration(ms: number): string {
  const seconds = Math.floor(ms / 1000)
  if (seconds < 60) return `${seconds}s`
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m`
  const hours = Math.floor(minutes / 60)
  return `${hours}h ${minutes % 60}m`
}

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId } = await params
  const { data: org } = await getOrg({ orgId })

  return {
    title: `${org.name} | Dashboard`,
  }
}

export default async function OrgDashboard({ params }: TPageProps<'org-id'>) {
  const { ['org-id']: orgId } = await params
  const session = await auth0.getSession()
  const [{ data: org, error }, { data: stats }] = await Promise.all([
    getOrg({ orgId }),
    getOrgStats({ orgId }),
  ])

  if (error && !org) {
    return (
      <main>
        <h1>Welcome, {session.user.name}!</h1>
        <p>Could not load your organization.</p>
        <div className="flex items-center gap-4">
          <Link href="/">Return home</Link>{' '}
          <Link href="/api/auth/logout">Log out</Link>{' '}
        </div>
      </main>
    )
  }

  if (org?.features?.['org-dashboard']) {
    // Fetch real activity from runner jobs
    const runner = org?.runner_group?.runners?.at(0)
    let recentActivities: Array<{
      id: string
      installName: string
      installId: string
      message: string
      status: string
      created_at: string
      duration?: string
      triggeredBy: string
      href?: string
    }> = []

    if (runner) {
      const { data: jobs } = await getRunnerJobs({
        orgId,
        runnerId: runner.id,
        groups: ['deploy', 'operations', 'sync'],
        limit: 10,
      })

      if (jobs) {
        recentActivities = jobs.map((job) => ({
          id: job.id,
          installName: getJobName(job),
          installId: job.metadata?.install_id || '',
          message: getJobExecutionStatus(job),
          status: job.status,
          created_at: job.created_at,
          duration: job.finished_at && job.started_at
            ? formatDuration(
                new Date(job.finished_at).getTime() -
                  new Date(job.started_at).getTime()
              )
            : undefined,
          triggeredBy: '', // Job doesn't have creator info directly
          href: getJobHref(job),
        }))
      }
    }

    return org?.features?.['stratus-layout'] ? (
      <PageLayout className="divide-y" isScrollable>
        <Breadcrumbs
          breadcrumbs={[
            {
              path: `/${orgId}`,
              text: org?.name,
            },
          ]}
        />
        <PageHeader>
          <HeadingGroup>
            <Text variant="h3" weight="stronger" level={1} role="heading">
              Welcome, {session.user.name?.split(' ')[0]}!
            </Text>
            <Text theme="neutral">
              Manage your applications and deployed installs.
            </Text>
          </HeadingGroup>
        </PageHeader>

        <PageContent>
          <PageGrid className="md:divide-x flex-auto !grid-cols-1 md:!grid-cols-[1fr_400px]">
            {/* Main Content */}
            <PageSection className="flex-1 border-r">
              <Text variant="h3" weight="strong">
                Overview
              </Text>

              <StatsGrid
                stats={[
                  { label: 'Total Installs', value: stats?.install_count || 0 },
                  { label: 'Active Applications', value: stats?.app_count || 0 },
                  { label: 'Active Runner', value: 0 },
                  { label: 'Active Installs', value: stats?.install_count || 0 },
                ]}
              />

              <Text variant="h3" weight="strong" className="mt-6">
                Recent activities
              </Text>

              <RecentActivities activities={recentActivities} orgId={orgId} />
            </PageSection>

            {/* Sidebar */}
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
    ) : (
      <DashboardContent
        breadcrumb={[{ href: `/${orgId}`, text: 'Dashboard' }]}
        heading={org?.name}
        headingUnderline={org?.id}
        statues={
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
        }
      >
        <div className="flex-auto md:grid md:grid-cols-12 divide-x">
          <div className="divide-y flex flex-col flex-auto col-span-8">
            <Section heading="Overview" className="flex-initial">
              <OldText variant="reg-12">TKTK</OldText>
            </Section>
            <Section className="flex-initial" heading="Workspaces">
              <OldText variant="reg-12">TKTK</OldText>
            </Section>
          </div>
          <div className="divide-y flex flex-col flex-auto col-span-4">
            <Section className="flex-initial">
              <div className="flex flex-col gap-3">
                <span>
                  <OldText variant="med-18">Introducing Nuon Actions!</OldText>
                  <OldText
                    className="text-cool-grey-600 dark:text-white/70"
                    variant="reg-12"
                  >
                    Mar 5, 2025
                  </OldText>
                </span>
                <OldText variant="reg-14" className="!leading-relaxed">
                  Nuon Actions allow you to create automated workflows that can
                  be run in installs. Actions are useful for debugging, running
                  scripts, and implementing health checks.
                </OldText>
                <OldLink
                  href="https://docs.nuon.co/concepts/nuon-actions"
                  target="_blank"
                  className="text-base"
                >
                  Check it out <CaretRightIcon />
                </OldLink>
              </div>
            </Section>
            <Section className="flex-initial" heading="Recent activity">
              <OldText variant="reg-12">TKTK</OldText>
            </Section>
          </div>
        </div>
      </DashboardContent>
    )
  } else {
    if (org?.features?.['org-runner']) {
      redirect(`/${orgId}/runner`)
    } else {
      redirect(`/${orgId}/apps`)
    }
  }
}
