import { api } from '@/lib/api'

export const getInstallBranchNames = ({ orgId }: { orgId: string }) =>
  api<string[]>({
    path: 'installs/branch-names',
    orgId,
  })
