'use server'

import { revalidatePath } from 'next/cache'
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
  try {
    await reprovisionInstallSandbox({ installId, orgId })
    revalidatePath(`/${orgId}/installs/${installId}`)
  } catch (error) {
    console.error(error)
    throw new Error(error.message)
  }
}

interface IDeployComponents {
  installId: string
  orgId: string
}

export async function deployComponents({
  installId,
  orgId,
}: IDeployComponents) {
  try {
    await deployAllComponents({
      installId,
      orgId,
    })
    revalidatePath(`/${orgId}/installs/${installId}`)
  } catch (error) {
    console.error(error)
    throw new Error(error.message)
  }
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
  try {
    await deployComponentByBuildId({
      buildId,
      installId,
      orgId,
    })
    revalidatePath(`/${orgId}/installs/${installId}`)
  } catch (error) {
    console.error(error)
    throw new Error(error.message)
  }
}

interface IRevalidateInstallData {
  installId: string
  orgId: string
}

export async function revalidateInstallData({
  orgId,
  installId,
}: IRevalidateInstallData) {
  revalidatePath(`/${orgId}/installs/${installId}`)
}
