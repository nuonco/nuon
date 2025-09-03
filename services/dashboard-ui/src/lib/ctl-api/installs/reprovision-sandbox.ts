import { mutateData } from '@/utils'
import type { IGetInstall } from '../shared-interfaces'

export interface IReprovisionSandbox extends IGetInstall {}

export async function reprovisionSandbox({
  installId,
  orgId,
}: IReprovisionSandbox) {
  return mutateData({
    data: { error_behavior: 'string' },
    errorMessage: 'Unable to reprovision sandbox.',
    orgId,
    path: `installs/${installId}/reprovision-sandbox`,
  })
}