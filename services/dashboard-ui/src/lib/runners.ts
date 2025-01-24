import type { TRunner, TRunnerJob, TLogStream, TOTELLog } from '@/types'
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

export interface IGetRunnerJobs extends IGetRunner {}

export async function getRunnerJobs({ orgId, runnerId }: IGetRunnerJobs) {
  return queryData<Array<TRunnerJob>>({
    errorMessage: 'Unable to retrieve runner jobs.',
    orgId,
    path: `runners/${runnerId}/jobs`,
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

// TODO: janky log pagination
// investigate handling logs with sse

async function pageLogs({ orgId, logStreamId, next = '0' }) {
  return fetch(
    `${API_URL}/v1/log-streams/${logStreamId}/logs`,
    await getFetchOpts(orgId, { 'X-Nuon-API-Offset': next })
  ).then(async (res) => {
    if (res.ok) {
      if (res.headers.get('x-nuon-api-next')) {
        return [
          ...(await res.json()),
          ...(await pageLogs({
            orgId,
            logStreamId,
            next: res.headers.get('x-nuon-api-next'),
          })),
        ]
      } else {
        return res.json()
      }
    } else {
      throw new Error('Failed to fetch log stream logs')
    }
  })
}

export interface IGetLogStreamLogs {
  logStreamId: string
  orgId: string
}

export async function getLogStreamLogs({
  logStreamId,
  orgId,
}: IGetLogStreamLogs): Promise<Array<TOTELLog>> {
  return pageLogs({ logStreamId, orgId })
}

export interface ICancelRunnerJob {
  runnerJobId: string
  orgId: string
}

export async function cancelRunnerJob({ orgId, runnerJobId}: ICancelRunnerJob) {
  return mutateData<TRunnerJob>({
    data: {},
    errorMessage: "Unable to cancel runner job.",
    orgId,
    path: `runner-jobs/${runnerJobId}/cancel`
  })
}
