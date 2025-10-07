import { api } from '@/lib/api'
import type { TDriftedObject } from '@/types'

export const getInstallDriftedObjects = ({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) =>
  api<TDriftedObject[]>({
    path: `installs/${installId}/drifted-objects`,
    orgId,
  })
