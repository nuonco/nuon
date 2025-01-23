'use server'

import { revalidatePath } from 'next/cache'
import { createComponentBuild as createBuild } from '@/lib'

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
