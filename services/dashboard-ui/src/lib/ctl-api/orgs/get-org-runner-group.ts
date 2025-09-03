import type { TRunnerGroup } from '@/types'
import { queryData } from '@/utils'
import type { IGetOrg } from '../shared-interfaces'

export interface IGetOrgRunnerGroup extends IGetOrg {}

export async function getOrgRunnerGroup({ orgId }: IGetOrgRunnerGroup) {
  return queryData<TRunnerGroup>({
    errorMessage: 'Unable to retrieve install runner group.',
    orgId,
    path: `orgs/current/runner-group`,
  })
}