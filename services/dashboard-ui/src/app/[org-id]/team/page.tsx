import { Suspense, type FC } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { CaretRight, Heartbeat, Timer } from '@phosphor-icons/react/dist/ssr'
import {
  Config,
  ConfigContent,
  CancelRunnerJobButton,
  DashboardContent,
  Duration,
  EmptyStateGraphic,
  ErrorFallback,
  ID,
  Link,
  Loading,
  StatusBadge,
  RunnerHeartbeatChart,
  Section,
  SubNav,
  Text,
  Time,
  Timeline,
  ToolTip,
  Truncate,
} from '@/components'
import {
  getOrg,
  getRunner,
  getRunnerHealthChecks,
  getRunnerJobs,
  getRunnerLatestHeartbeat,
} from '@/lib'

export default async function OrgTeam({ params }) {
  const orgId = params?.['org-id'] as string
  const org = await getOrg({ orgId })

  return (
    <DashboardContent
      breadcrumb={[{ href: `/${orgId}`, text: org?.name }]}
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
      meta={
        <SubNav
          links={[
            { href: `/${orgId}`, text: 'Runner' },
            { href: `/${orgId}/team`, text: 'Team' },
          ]}
        />
      }
    >
      <Section>Team members</Section>
    </DashboardContent>
  )
}
