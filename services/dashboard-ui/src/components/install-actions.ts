'use server'

import { reprovisionInstall as reprovisionInstallSandbox } from '@/lib'

interface IReprovisionInstall {
  installId: string
  orgId: string
}

export async function reprovisionInstall({
  installId,
  orgId,
}: IReprovisionInstall) {
  return reprovisionInstallSandbox({ installId, orgId })
}
