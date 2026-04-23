import { api } from '@/lib/api'

export const getInstallLabelKeys = ({ orgId }: { orgId: string }) =>
  api<Record<string, string[]>>({
    path: 'installs/label-keys',
    orgId,
  })
