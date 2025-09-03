import type { TInstall } from '@/types'
import { queryData } from '@/utils'
import type { IGetInstall } from '../shared-interfaces'

export async function getInstall({ installId, orgId }: IGetInstall) {
  return queryData<TInstall>({
    errorMessage: 'Unable to retrieve install.',
    orgId,
    path: `installs/${installId}`,
  })
}