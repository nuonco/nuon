import { api } from '@/lib/api'

export const listServiceAccountPurposes = ({ orgId }: { orgId: string }) =>
  api<string[]>({
    path: 'service-accounts/purposes',
    orgId,
  })
