// @ts-nocheck
// TODO(nnnat): URLSearchParams typing is terrible.
// What we're doing now is legit but TS doesn't think so.
import type {
  TInstallDeployPlan,
  TRunner,
  TRunnerJob,
  TLogStream,
  TRunnerHeartbeat,
  TRunnerHealthCheck,
} from '@/types'
import { API_URL, getFetchOpts, mutateData, queryData } from '@/utils'

export interface IGetRunner {
  orgId: string
  runnerId: string
}

export async function getRunner({ orgId, runnerId }: IGetRunner) {
  return queryData<TRunner>({
    errorMessage: 'Unable to retrieve runner.',
    orgId,
    path: `runners/${runnerId}`,
  })
}

export type TRunnerJobGroup =
  | 'sync'
  | 'build'
  | 'deploy'
  | 'sandbox'
  | 'runner'
  | 'actions'
  | 'operations'
  | 'health-checks'

export type TRunnerJobStatus =
  | 'queued'
  | 'available'
  | 'in-progress'
  | 'finished'
  | 'failed'
  | 'timed-out'
  | 'not-attempted'
  | 'cancelled'
  | 'unknown'

export interface IGetRunnerJobs extends IGetRunner {
  options?: {
    offset?: string
    limit?: string
    groups?: Array<TRunnerJobGroup>
    statuses?: Array<TRunnerJobStatus>
  }
}

export type TPagination = {
  hasNext: string
  offset: string
}

export async function getRunnerJobs({
  orgId,
  runnerId,
  options = {},
}: IGetRunnerJobs): Promise<{
  runnerJobs: Array<TRunnerJob>
  pageData?: TPagination
}> {
  const params = new URLSearchParams(options).toString()

  // return queryData<Array<TRunnerJob>>({
  //   errorMessage: 'Unable to retrieve runner jobs.',
  //   orgId,
  //   path: `runners/${runnerId}/jobs${params ? '?' + params : params}`,
  // })

  const res = await fetch(
    `${API_URL}/v1/runners/${runnerId}/jobs${params ? '?' + params : params}`,
    await getFetchOpts(orgId, { 'x-nuon-pagination-enabled': true })
  )
  const runnerJobs = await res.json()

  return {
    runnerJobs,
    pageData: {
      hasNext: res.headers.get('x-nuon-page-next') || 'false',
      offset: res.headers?.get('x-nuon-page-offset') || '0',
    },
  }
}

export interface IGetRunnerJob extends Omit<IGetRunner, 'runnerId'> {
  runnerJobId: string
}

export async function getRunnerJob({ orgId, runnerJobId }: IGetRunnerJob) {
  return queryData<TRunnerJob>({
    errorMessage: 'Unable to retrieve runner job.',
    orgId,
    path: `runner-jobs/${runnerJobId}`,
  })
}

export interface IGetLogStream extends Omit<IGetRunner, 'runnerId'> {
  logStreamId: string
}

export async function getLogStream({ logStreamId, orgId }: IGetLogStream) {
  return queryData<TLogStream>({
    errorMessage: 'Unable to retrieve log stream.',
    orgId,
    path: `log-streams/${logStreamId}`,
  })
}

export interface ICancelRunnerJob {
  runnerJobId: string
  orgId: string
}

export async function cancelRunnerJob({
  orgId,
  runnerJobId,
}: ICancelRunnerJob) {
  return mutateData<TRunnerJob>({
    data: {},
    errorMessage: 'Unable to cancel runner job.',
    orgId,
    path: `runner-jobs/${runnerJobId}/cancel`,
  })
}

export interface IGetRunnerHealthChecks extends IGetRunner {}

export async function getRunnerHealthChecks({
  orgId,
  runnerId,
}: IGetRunnerHealthChecks) {
  return queryData<Array<TRunnerHealthCheck>>({
    errorMessage: 'Unable to retrieve runner recent health checks.',
    orgId,
    path: `runners/${runnerId}/recent-health-checks`,
  })
}

export interface IGetRunnerLatestHeartbeat extends IGetRunner {}

export async function getRunnerLatestHeartbeat({
  orgId,
  runnerId,
}: IGetRunnerLatestHeartbeat) {
  return queryData<TRunnerHeartbeat>({
    errorMessage: 'Unable to retrieve latest runner heartbeat.',
    orgId,
    path: `runners/${runnerId}/latest-heart-beat`,
  })
}

export interface IGetRunnerJobPlan {
  orgId: string
  runnerJobId: string
}

export async function getRunnerJobPlan({
  orgId,
  runnerJobId,
}: IGetRunnerJobPlan) {
  return queryData<any>({
    errorMessage: 'Unable to retrieve runner job plant',
    orgId,
    path: `runner-jobs/${runnerJobId}/plan`,
  })
}
