'use server'

import { revalidatePath } from 'next/cache'
import {
  deployComponents as deployAllComponents,
  reprovisionInstall as reprovisionInstallSandbox,
  deployComponentBuild as deployComponentByBuildId,
  teardownInstallComponents,
  updateInstall as patchInstall,
  forgetInstall as forget,
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

interface ITeardownAllComponents {
  installId: string
  orgId: string
}

export async function teardownAllComponents({
  installId,
  orgId,
}: ITeardownAllComponents) {
  try {
    await teardownInstallComponents({
      installId,
      orgId,
    })
    revalidatePath(`/${orgId}/installs/${installId}`)
  } catch (error) {
    console.error(error)
    throw new Error(error.message)
  }
}

interface IUpdateInstall {
  installId: string
  orgId: string
  formData: FormData
}

export async function updateInstall({
  installId,
  orgId,
  formData: fd,
}: IUpdateInstall) {
  const formData = Object.fromEntries(fd)

  const inputs = Object.keys(formData).reduce((acc, key) => {
    if (key.includes('inputs:')) {
      let value: any = formData[key]
      if (value === 'on' || value === 'off') {
        value = Boolean(value === 'on').toString()
      }

      acc[key.replace('inputs:', '')] = value
    }

    return acc
  }, {})

  let data = {
    inputs,
    name: formData.name as string,
  }

  return patchInstall({
    data,
    installId,
    orgId,
  })
}

interface IForgetInstall {
  installId: string
  orgId: string
}

export async function forgetInstall(params: IForgetInstall) {
  return forget(params)
}
