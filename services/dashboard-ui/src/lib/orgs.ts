import type { TOrg, TVCSConnection, TRunnerGroup } from '@/types'
import { mutateData, queryData, nueQueryData } from '@/utils'

export async function getOrgs() {
  return queryData<Array<TOrg>>({
    errorMessage: 'Unable to retrieve your organizations.',
    path: 'orgs',
  })
}

export interface IGetOrg {
  orgId: string
}

export async function getOrg({ orgId }: IGetOrg) {
  return queryData<TOrg>({
    errorMessage: 'Unable to retrieve organization.',
    orgId,
    path: 'orgs/current',
  })
}

export interface IGetVCSConnections extends IGetOrg {}

export async function getVCSConnections({ orgId }: IGetVCSConnections) {
  return queryData<Array<TVCSConnection>>({
    errorMessage: 'Unable to retrieve connected version control systems',
    orgId,
    path: `vcs/connections`,
  })
}

export async function postOrg(data: { name: string }) {
  return mutateData<TOrg>({
    data,
    errorMessage:
      'Unable to create your organization, refresh the page and try again.',
    path: 'orgs',
  })
}

export async function nueGetOrg({ orgId }: IGetOrg) {
  return nueQueryData<TOrg>({
    orgId,
    path: 'orgs/current',
  })
}

export interface IGetOrgRunnerGroup extends IGetOrg {}

export async function getOrgRunnerGroup({ orgId }: IGetOrgRunnerGroup) {
  return queryData<TRunnerGroup>({
    errorMessage: 'Unable to retrieve install runner group.',
    orgId,
    path: `orgs/current/runner-group`,
  })
}
