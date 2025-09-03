import type { TOrg } from '@/types'
import { queryData } from '@/utils'

export async function getOrgs() {
  return queryData<Array<TOrg>>({
    errorMessage: 'Unable to retrieve your organizations.',
    path: 'orgs',
  })
}