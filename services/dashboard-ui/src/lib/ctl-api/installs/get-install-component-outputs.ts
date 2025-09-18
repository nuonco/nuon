import { api } from '@/lib/api'
import type { TInstallComponentOutputs } from '@/types'

export const getInstallComponentOutputs = async ({
  componentId,
  installId,
  orgId,
}: {
  componentId: string
  installId: string
  orgId: string
}) =>
  api<TInstallComponentOutputs>({
    orgId,
    path: `installs/${installId}/components/${componentId}/outputs`,
  })
