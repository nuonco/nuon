'use server'

// import { revalidatePath } from 'next/cache'
import {
  deployComponents as deployAllComponents,
  reprovisionInstall as reprovisionInstallSandbox,
} from '@/lib'

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

interface IDeployComponents {
  installId: string
  orgId: string
}

export async function deployComponents({
  installId,
  orgId,
}: IDeployComponents) {
  return deployAllComponents({
    installId,
    orgId,
  })
}
