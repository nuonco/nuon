import { mutateData } from '@/utils'
import type { IGetInstall } from '../shared-interfaces'

export interface IForgetInstall extends IGetInstall {}

export async function forgetInstall({ installId, orgId }: IForgetInstall) {
  return mutateData<boolean>({
    errorMessage: 'Unable to forget install.',
    orgId,
    path: `installs/${installId}/forget`,
  })
}