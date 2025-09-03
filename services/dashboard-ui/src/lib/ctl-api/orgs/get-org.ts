import type { TOrg } from '@/types'
import { queryData } from '@/utils'
import type { IGetOrg } from '../shared-interfaces'

export async function getOrg({ orgId }: IGetOrg) {
  return queryData<TOrg>({
    errorMessage: 'Unable to retrieve organization.',
    orgId,
    path: 'orgs/current',
  })
}