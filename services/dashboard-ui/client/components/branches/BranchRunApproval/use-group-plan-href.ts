import { useCallback, useMemo } from 'react'
import { useParams } from 'react-router'
import { useBranch } from '@/hooks/use-branch'
import { latestBranchConfig } from '@/utils/branch-utils'

export const getGroupName = (name?: string) =>
  name?.replace(/^plan install group:\s*/i, '').trim() || 'install group'

export const useGroupPlanHref = () => {
  const { branch } = useBranch()
  const { orgId, appId, branchId } = useParams()
  const groups = useMemo(
    () => latestBranchConfig(branch)?.install_groups ?? [],
    [branch]
  )

  return useCallback(
    (stepName?: string) => {
      const rolloutHref = `/${orgId}/apps/${appId}/branches/${branchId}/rollout`
      const groupName = getGroupName(stepName).toLowerCase()
      const group = groups.find(
        (item) => item.name?.toLowerCase() === groupName
      )
      if (!group?.id) return rolloutHref
      const params = new URLSearchParams({ group: group.id })
      return `${rolloutHref}?${params.toString()}`
    },
    [orgId, appId, branchId, groups]
  )
}
