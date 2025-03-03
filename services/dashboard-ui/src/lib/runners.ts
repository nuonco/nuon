// @ts-nocheck
// TODO(nnnat): URLSearchParams typing is terrible.
// What we're doing now is legit but TS doesn't think so.
import type {
  TRunner,
  TRunnerJob,
  TLogStream,
  TOTELLog,
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

export interface IGetRunnerJobs extends IGetRunner {
  options?: {
    limit?: string
    groups?: Array<
      | 'sync'
      | 'build'
      | 'deploy'
      | 'sandbox'
      | 'runner'
      | 'actions'
      | 'operations'
      | 'health-checks'
    >
    statuses?: Array<
      | 'queued'
      | 'available'
      | 'in-progress'
      | 'finished'
      | 'failed'
      | 'timed-out'
      | 'not-attempted'
      | 'cancelled'
      | 'unknown'
    >
  }
}

export async function getRunnerJobs({
  orgId,
  runnerId,
  options = {},
}: IGetRunnerJobs) {
  const params = new URLSearchParams(options).toString()

  return queryData<Array<TRunnerJob>>({
    errorMessage: 'Unable to retrieve runner jobs.',
    orgId,
    path: `runners/${runnerId}/jobs${params ? '?' + params : params}`,
  })
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
