'use client'

import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Timeline, type ITimeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { useOrg } from '@/hooks/use-org'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import { useQuery } from '@/hooks/use-query'
import { useQueryParams } from '@/hooks/use-query-params'
import type { TPolicyReport, TRunnerJob } from '@/types'
import {
  getJobExecutionStatus,
  getJobHref,
  getJobName,
  type TJobGroup,
} from '@/utils/runner-utils'

export const RECENT_ACTIVITY_SEARCH_PARAM = 'recent-activity'
export const RECENT_ACTIVITY_LIMIT = 10
export const RECENT_ACTIVITY_GROUPS: TJobGroup[] = [
  'actions',
  'build',
  'deploy',
  'operations',
  'sandbox',
  'sync',
]
const HIDDEN_JOB_TYPES = ['fetch-image-metadata']

const getFilenameFromHeader = (contentDisposition: string | null) => {
  if (!contentDisposition) {
    return undefined
  }

  const match = contentDisposition.match(
    /filename\*?=(?:UTF-8'')?"?([^";]+)"?/i
  )

  if (!match?.[1]) {
    return undefined
  }

  try {
    return decodeURIComponent(match[1])
  } catch {
    return match[1]
  }
}

interface IRunnerRecentActivity
  extends Omit<ITimeline<TRunnerJob>, 'events' | 'renderEvent'>,
    IPollingProps {
  initJobs: Array<TRunnerJob>
  runnerId: string
}

const PolicyReportActions = ({ buildId }: { buildId: string }) => {
  const { org } = useOrg()
  const queryParams = useQueryParams({
    owner_type: 'component_builds',
    owner_id: buildId,
    format: 'opa',
    limit: 1,
    offset: 0,
  })

  const { data: policyReports } = useQuery<TPolicyReport[]>({
    path: `/api/orgs/${org?.id}/policy-reports${queryParams}`,
    enabled: !!org?.id,
    initIsLoading: false,
    initData: [],
    dependencies: [queryParams],
  })

  const policyReport = policyReports?.[0]
  if (!policyReport?.id || !org?.id) {
    return null
  }

  const denyCount = policyReport?.deny_count ?? 0
  const warnCount = policyReport?.warn_count ?? 0
  const passCount = policyReport?.pass_count ?? 0

  const handleExport = async () => {
    const exportWindow = window.open('', '_blank')
    if (!exportWindow) {
      return
    }

    exportWindow.document.title = 'Policy Report'
    exportWindow.document.body.innerText = 'Loading policy report...'

    const renderContent = (content: string) => {
      let parsed = content
      const normalized = content.trim()
      const tryParseJson = (value: string) =>
        JSON.stringify(JSON.parse(value), null, 2)

      try {
        parsed = tryParseJson(normalized)
      } catch {
        try {
          const decoded = atob(normalized)
          parsed = tryParseJson(decoded)
        } catch {
          parsed = content
        }
      }

      exportWindow.document.body.innerHTML = ''
      const pre = exportWindow.document.createElement('pre')
      pre.textContent = parsed
      pre.style.whiteSpace = 'pre-wrap'
      pre.style.wordBreak = 'break-word'
      pre.style.margin = '16px'
      exportWindow.document.body.appendChild(pre)
    }

    if (policyReport.content) {
      renderContent(policyReport.content)
      return
    }

    try {
      const listResponse = await fetch(
        `/api/orgs/${org.id}/policy-reports${queryParams}`
      )
      if (!listResponse.ok) {
        exportWindow.document.body.innerText =
          'Failed to load policy report.'
        return
      }

      const payload = (await listResponse.json()) as {
        data?: TPolicyReport[]
      }
      const reportContent = payload?.data?.[0]?.content
      if (!reportContent) {
        exportWindow.document.body.innerText =
          'Policy report content is unavailable.'
        return
      }

      renderContent(reportContent)
    } catch {
      exportWindow.document.body.innerText = 'Failed to load policy report.'
    }
  }

  return (
    <span className="flex items-center gap-3">
      <Text
        variant="subtext"
        theme="error"
        className="inline-flex items-center gap-1"
      >
        <Icon variant="XCircle" size={14} /> {denyCount}
      </Text>
      <Text
        variant="subtext"
        theme="warn"
        className="inline-flex items-center gap-1"
      >
        <Icon variant="WarningCircle" size={14} /> {warnCount}
      </Text>
      <Text
        variant="subtext"
        theme="success"
        className="inline-flex items-center gap-1"
      >
        <Icon variant="CheckCircle" size={14} /> {passCount}
      </Text>
      <Button
        aria-label="Export policy report"
        size="xs"
        variant="ghost"
        onClick={handleExport}
      >
        <Icon variant="FileText" size={14} />
      </Button>
    </span>
  )
}

export const RunnerRecentActivity = ({
  initJobs,
  pagination,
  runnerId,
  shouldPoll = false,
  pollInterval = 20000,
}: IRunnerRecentActivity) => {
  const { org } = useOrg()
  const queryParams = useQueryParams({
    offset: pagination?.offset,
    limit: 10,
  })
  const { data: jobs } = usePolling<TRunnerJob[]>({
    dependencies: [queryParams],
    path: `/api/orgs/${org?.id}/runners/${runnerId}/jobs${queryParams}`,
    shouldPoll,
    initData: initJobs,
    pollInterval,
  })

  const visibleJobs = jobs?.filter(
    (job) => !HIDDEN_JOB_TYPES.includes(job.type)
  )

  return (
    <Timeline<TRunnerJob>
      events={visibleJobs}
      pagination={pagination}
      renderEvent={(job) => {
        const jobHref = getJobHref(job)
        const jobTitle =
          jobHref === '' ? (
            <>
              {getJobName(job)} {getJobExecutionStatus(job)}
            </>
          ) : (
            <>
              <Link href={jobHref}>{getJobName(job)}</Link>{' '}
              {getJobExecutionStatus(job)}
            </>
          )

        return (
          <TimelineEvent
            key={job.id}
            actions={
              job.group === 'build' && job?.metadata?.component_build_id ? (
                <PolicyReportActions
                  buildId={job.metadata.component_build_id}
                />
              ) : null
            }
            caption={<ID>{job?.id}</ID>}
            createdAt={job?.created_at}
            status={job?.status}
            title={jobTitle}
          />
        )
      }}
    />
  )
}
