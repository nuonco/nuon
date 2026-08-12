import { api } from '@/lib/api'

export const deleteRole = ({
  roleId,
  orgId,
}: {
  roleId: string
  orgId: string
}) =>
  api<null>({
    path: `roles/${roleId}`,
    method: 'DELETE',
    orgId,
  })
