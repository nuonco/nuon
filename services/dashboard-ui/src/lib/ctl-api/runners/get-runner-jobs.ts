// @ts-nocheck
// TODO(nnnat): URLSearchParams typing is terrible.
// What we're doing now is legit but TS doesn't think so.
import type { TRunnerJob } from '@/types'
import { API_URL } from '@/configs/api'
import { getFetchOpts } from '@/utils'
import type { IGetRunner } from '../shared-interfaces'

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
