'use server'

// import { revalidatePath } from 'next/cache'
import {
  deployComponents as deployAllComponents,
  reprovisionInstall as reprovisionInstallSandbox,
  deployComponentBuild as deployComponentByBuildId,
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

interface IDeployComponentBuild {
  buildId: string
  installId: string
  orgId: string
}

export async function deployComponentBuild({
  buildId,
  installId,
  orgId,
}: IDeployComponentBuild) {
  return deployComponentByBuildId({
    buildId,
    installId,
    orgId,
  })
}
