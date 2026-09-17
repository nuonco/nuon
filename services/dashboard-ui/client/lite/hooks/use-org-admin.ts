import { useQuery } from '@tanstack/react-query'
import { getAccount } from '@/lib'
import { useOrg } from '../providers/org-provider'

export const useOrgAdmin = (): boolean => {
  const { orgId } = useOrg()

  const { data: account } = useQuery({
    queryKey: ['account'],
    queryFn: getAccount,
    staleTime: 60_000,
  })

  if (!orgId) return false

  return (
    account?.roles?.some(
      (role) => role.org_id === orgId && role.role_type === 'org_admin'
    ) ?? false
  )
}
