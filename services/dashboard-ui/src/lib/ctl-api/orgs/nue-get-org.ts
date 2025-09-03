import type { TOrg } from '@/types'
import { nueQueryData } from '@/utils'
import type { IGetOrg } from '../shared-interfaces'

export async function nueGetOrg({ orgId }: IGetOrg) {
  return nueQueryData<TOrg>({
    orgId,
    path: 'orgs/current',
  })
}