'use server'

import { revalidatePath } from 'next/cache'
import {
  deployComponents as deployAllComponents,
  reprovisionInstall as reprovision,
  reprovisionSandbox as reprovisionSBox,
  deployComponentBuild as deployComponentByBuildId,
  teardownInstallComponents,
  updateInstall as patchInstall,
  forgetInstall as forget,
} from '@/lib'
import { mutateData } from '@/utils'
import type { TInstall } from '@/types'

interface IReprovisionInstall {
  installId: string
  orgId: string
}

export async function reprovisionInstall({
  installId,
  orgId,
}: IReprovisionInstall) {
  try {
    await reprovision({ installId, orgId })
    revalidatePath(`/${orgId}/installs/${installId}`)
  } catch (error) {
    console.error(error)
    throw new Error(error.message)
  }
}

export async function reprovisionSandbox({
  installId,
  orgId,
}: IReprovisionInstall) {
  try {
    await reprovisionSBox({ installId, orgId })
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
    revalidatePath(`/${orgId}/installs/${installId}/components`)
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

  if (Object.keys(inputs)?.length > 0) {
    try {
      await mutateData({
        errorMessage: 'Unable to update install inputs',
        data: { inputs },
        method: 'PATCH',
        orgId,
        path: `installs/${installId}/inputs`,
      })
    } catch (error) {
      console.error(error?.message)
    }
  }

  let install: TInstall
  try {
    install = await patchInstall({
      data: {
        name: formData.name as string,
      },
      installId,
      orgId,
    }).then((ins) => {
      if (formData?.['form-control:update'] === 'update') {
        reprovision({ orgId, installId })
          .then(() => {
            deployAllComponents({ orgId, installId }).catch(console.error)
          })
          .catch(console.error)
      }

      return ins
    })
  } catch (error) {
    console.error(error?.message)
    throw new Error('unable to patch install')
  }

  return install
}

interface IForgetInstall {
  installId: string
  orgId: string
}

export async function forgetInstall(params: IForgetInstall) {
  return forget(params)
}

interface IDeleteComponents {
  installId: string
  orgId: string
  force?: boolean
}

export async function deleteComponents({
  installId,
  orgId,
  force = false,
}: IDeleteComponents) {
  // @ts-ignore
  const params = new URLSearchParams({ force })
  return mutateData({
    errorMessage: 'Unable to delete components',
    orgId,
    method: 'DELETE',
    path: `installs/${installId}/components?${params.toString()}`,
  })
}

interface IDeleteComponent {
  componentId: string
  installId: string
  orgId: string
  force?: boolean
}

export async function deleteComponent({
  componentId,
  installId,
  orgId,
  force = false,
}: IDeleteComponent) {
  // @ts-ignore
  const params = new URLSearchParams({ force })
  return mutateData({
    errorMessage: 'Unable to delete component',
    orgId,
    method: 'DELETE',
    path: `installs/${installId}/components/${componentId}?${params.toString()}`,
  })
}

interface IDeleteInstall {
  installId: string
  orgId: string
  force?: boolean
}

export async function deleteInstall({
  installId,
  orgId,
  force = false,
}: IDeleteInstall) {
  // @ts-ignore
  const params = new URLSearchParams({ force })
  return mutateData({
    errorMessage: 'Unable to delete install',
    orgId,
    method: 'DELETE',
    path: `installs/${installId}?${params.toString()}`,
  })
}
