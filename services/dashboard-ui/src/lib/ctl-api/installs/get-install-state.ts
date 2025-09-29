import { api } from '@/lib/api'

export const getInstallState = ({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) =>
  api<Record<string, any>>({
    path: `installs/${installId}/state`,
    orgId,
  })
