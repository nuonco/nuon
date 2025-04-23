'use server'

import { revalidatePath } from 'next/cache'
import {
  createComponentBuild as createBuild,
  createInstall,
  type ICreateInstallData,
} from '@/lib'

interface IRevalidateAppData {
  appId: string
  orgId: string
}

export async function revalidateAppData({ appId, orgId }: IRevalidateAppData) {
  revalidatePath(`/${orgId}/apps/${appId}`)
}

interface ICreateComponentBuild {
  appId: string
  componentId: string
  orgId: string
}

export async function createComponentBuild({
  appId,
  componentId,
  orgId,
}: ICreateComponentBuild) {
  try {
    await createBuild({
      componentId,
      orgId,
    })
    revalidatePath(`/${orgId}/apps/${appId}/components/${componentId}`)
  } catch (error) {
    console.error(error)
    throw new Error(error.message)
  }
}

interface ICreateAppInstall {
  appId: string
  orgId: string
  formData: FormData
  platform: string | 'aws' | 'azure'
}

export async function createAppInstall({
  appId,
  orgId,
  formData: fd,
  platform,
}: ICreateAppInstall) {
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

  let data: ICreateInstallData = {
    inputs,
    name: formData.name as string,
  }

  // if (platform === 'aws') {
  //   data = {
  //     aws_account: {
  //       iam_role_arn: formData?.iam_role_arn as string,
  //       region: formData?.region as string,
  //     },
  //     ...data,
  //   }
  // }

  // if (platform === 'azure') {
  //   data = {
  //     azure_account: {
  //       location: formData?.location as string,
  //       service_principal_app_id: formData?.service_principal_app_id as string,
  //       service_principal_password:
  //         formData?.service_principal_password as string,
  //       subscription_id: formData?.subscription_id as string,
  //       subscription_tenant_id: formData?.subscription_tenant_id as string,
  //     },
  //     ...data,
  //   }
  // }

  return createInstall({
    appId,
    orgId,
    data,
  })
}

interface IBuildComponents {
  appId: string
  componentIds: Array<string>
  orgId: string
}

export async function buildComponents({
  appId,
  componentIds,
  orgId,
}: IBuildComponents) {
  return Promise.all(
    componentIds.map(
      async (cId) => await createBuild({ componentId: cId, orgId })
    )
  ).then((builds) => {
    revalidatePath(`/${orgId}/apps/${appId}`)
    return builds
  })
}
