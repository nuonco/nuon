import { mutateData } from '@/utils'
import type { IGetInstall } from '../shared-interfaces'

export interface IReprovisionInstall extends IGetInstall {}

export async function reprovisionInstall({
  installId,
  orgId,
}: IReprovisionInstall) {
  return mutateData({
    data: { error_behavior: 'string' },
    errorMessage: 'Unable to reprovision install.',
    orgId,
    path: `installs/${installId}/reprovision`,
  })
}