import { api } from '@/lib/api'
import type { TInstallComponent } from '@/types'

export const getInstallComponent = ({
  installId,
  componentId,
  orgId,
}: {
  installId: string
  componentId: string
  orgId: string
}) =>
  api<TInstallComponent>({
    path: `installs/${installId}/components/${componentId}`,
    orgId,
  })
