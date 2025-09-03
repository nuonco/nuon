import type { TInstall } from '@/types'
import { mutateData } from '@/utils'
import type { IGetInstall } from '../shared-interfaces'

export interface IUpdateInstall extends IGetInstall {
  data: {
    name: string
    inputs?: Record<string, string>
  }
}

export async function updateInstall({
  data,
  installId,
  orgId,
}: IUpdateInstall) {
  return mutateData<TInstall>({
    errorMessage: 'Unable to update install.',
    data,
    orgId,
    method: 'PATCH',
    path: `installs/${installId}`,
  })
}