import type { TBuild } from '@/types'
import { mutateData } from '@/utils'

interface ICreateComponentBuild {
  componentId: string
  orgId: string
}

export async function createComponentBuild({
  componentId,
  orgId,
}: ICreateComponentBuild) {
  return mutateData<TBuild>({
    data: { use_latest: true },
    errorMessage: 'Unable to kick off component build',
    orgId,
    path: `components/${componentId}/builds`,
  })
}