import type { ReactNode } from 'react'
import { Banner } from '@/components/common/Banner'
import { Duration } from '@/components/common/Duration'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { SectionHeader } from '@/components/layout/SectionHeader'
import type { TRunnerJob } from '@/types'
import { getStatusTheme } from '@/utils/status-utils'
import { humanize } from '@/utils/string-utils'

export interface IRunSummaryTiming {
  label: string
  time?: string
}

export interface IRunSummary {
  children?: ReactNode
  duration?: { beginTime?: string; endTime?: string }
  isLoading?: boolean
  jobHref?: (job: TRunnerJob) => string | undefined
  jobs?: TRunnerJob[]
  status?: { status?: string; status_human_description?: string }
  statusDescription?: string
  timings?: IRunSummaryTiming[]
  triggeredBy?: ReactNode
}

const jobLabel = (job: TRunnerJob) =>
  humanize(job?.operation ?? job?.type ?? 'Job')

export const RunSummary = ({
  children,
  duration,
  isLoading,
  jobHref,
  jobs,
  status,
  statusDescription,
  timings,
  triggeredBy,
}: IRunSummary) => {
  const hasFailed = getStatusTheme(status?.status ?? '') === 'error'
  const failedJobs = (jobs ?? []).filter(
    (job) =>
      getStatusTheme(job?.status_v2?.status ?? job?.status ?? '') === 'error'
  )
  const reason =
    status?.status_human_description ||
    statusDescription ||
    failedJobs.at(0)?.status_description

  return (
    <div className="flex flex-col gap-6">
      {hasFailed && reason ? (
        <Banner theme="error">
          <div className="flex flex-col gap-1 min-w-0">
            <Text weight="strong">Run failed</Text>
            <Text>{reason}</Text>
            {failedJobs.map((job) =>
              job.status_description && job.status_description !== reason ? (
                <Text key={job.id} variant="subtext">
                  {jobLabel(job)}: {job.status_description}
                </Text>
              ) : null
            )}
          </div>
        </Banner>
      ) : null}

      <div className="flex flex-col gap-4">
        <SectionHeader title="Timing" />
        <div className="flex flex-wrap gap-x-8 gap-y-4 items-start">
          <LabeledValue label="Status" loading={isLoading}>
            <Status status={status?.status ?? 'unknown'} />
          </LabeledValue>
          {timings?.map(({ label, time }) => (
            <LabeledValue key={label} label={label} loading={isLoading}>
              {time ? (
                <Time variant="subtext" time={time} format="short-datetime" />
              ) : (
                <Text variant="subtext" theme="neutral">
                  —
                </Text>
              )}
            </LabeledValue>
          ))}
          {duration ? (
            <LabeledValue label="Duration" loading={isLoading}>
              <Duration
                variant="subtext"
                beginTime={duration.beginTime}
                endTime={duration.endTime}
              />
            </LabeledValue>
          ) : null}
          {triggeredBy ? (
            <LabeledValue label="Triggered by" loading={isLoading}>
              {triggeredBy}
            </LabeledValue>
          ) : null}
        </div>
      </div>

      <div className="flex flex-col gap-4">
        <SectionHeader title="Execution" />
        <Table
          columns={[
            {
              accessorKey: 'operation',
              header: 'Job',
              cell: ({ row }) => {
                const job = row.original
                const href = jobHref?.(job)
                return href ? (
                  <Link href={href} variant="inline">
                    {jobLabel(job)}
                  </Link>
                ) : (
                  <Text variant="subtext">{jobLabel(job)}</Text>
                )
              },
            },
            {
              accessorKey: 'status',
              header: 'Status',
              cell: ({ row }) => (
                <Status
                  status={
                    row.original?.status_v2?.status ??
                    row.original?.status ??
                    'unknown'
                  }
                />
              ),
            },
            {
              accessorKey: 'started_at',
              header: 'Started',
              cell: ({ row }) =>
                row.original?.started_at ? (
                  <Time
                    variant="subtext"
                    time={row.original.started_at}
                    format="relative"
                  />
                ) : (
                  <Text variant="subtext" theme="neutral">
                    —
                  </Text>
                ),
            },
            {
              accessorKey: 'execution_time',
              header: 'Duration',
              cell: ({ row }) => (
                <Duration
                  variant="subtext"
                  beginTime={row.original?.started_at}
                  endTime={row.original?.finished_at}
                />
              ),
            },
            {
              accessorKey: 'execution_count',
              header: 'Attempts',
              cell: ({ row }) => (
                <Text variant="subtext">
                  {row.original?.execution_count ?? 1}
                </Text>
              ),
            },
          ]}
          data={jobs ?? []}
          enableSearch={false}
          isLoading={isLoading}
          skeletonRows={2}
          emptyStateProps={{
            variant: 'table',
            size: 'sm',
            emptyTitle: 'No jobs yet',
            emptyMessage: 'Execution jobs appear here once the run starts.',
          }}
        />
      </div>

      {children}
    </div>
  )
}
