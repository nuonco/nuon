import type { TInstall } from '@/types'
import { queryData } from '@/utils'
import type { IGetInstalls } from '../shared-interfaces'

export async function getInstalls({ orgId }: IGetInstalls) {
  return queryData<Array<TInstall>>({
    errorMessage: 'Unable to retrieve your installs.',
    orgId,
    path: 'installs',
  })
}