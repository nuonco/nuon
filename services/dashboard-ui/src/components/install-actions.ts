'use server'

import { revalidatePath } from 'next/cache'
import { postInstallReprovision } from '@/lib'

interface IReprovisionInstall {
  installId: string
  orgId: string
}

export async function reprovisionInstall({
  installId,
  orgId,
}: IReprovisionInstall) {
  return postInstallReprovision({ installId, orgId })
  // revalidatePath(`/${orgId}/installs/${installId}`)
}
