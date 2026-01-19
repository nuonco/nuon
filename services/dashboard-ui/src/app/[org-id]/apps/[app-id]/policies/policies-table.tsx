import { Banner } from '@/components/common/Banner'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import {
  PoliciesTable as Table,
  policiesTableColumns,
} from '@/components/policies/PoliciesTable'
import { getAppPoliciesConfigs, getComponents } from '@/lib'

export const PoliciesTable = async ({
  appId,
  orgId,
}: {
  appId: string
  orgId: string
}) => {
  const [policiesResult, componentsResult] = await Promise.all([
    getAppPoliciesConfigs({ appId, orgId }),
    getComponents({ appId, orgId }),
  ])

  const { data: policiesConfigs, error, status } = policiesResult
  const { data: components } = componentsResult

  const latestConfig = policiesConfigs
    ?.slice()
    .sort((a, b) => {
      const dateA = a.created_at ? new Date(a.created_at).getTime() : 0
      const dateB = b.created_at ? new Date(b.created_at).getTime() : 0
      return dateB - dateA
    })
    .at(0)
  const policies = latestConfig?.policies || []

  const componentNameToId: Record<string, string> = {}
  components?.forEach((c) => {
    if (c.name && c.id) {
      componentNameToId[c.name] = c.id
    }
  })

  return error && status !== 404 ? (
    <Banner theme="error">Can&apos;t load policies: {error?.error}</Banner>
  ) : (
    <Table
      policies={policies}
      orgId={orgId}
      appId={appId}
      componentNameToId={componentNameToId}
    />
  )
}

export const PoliciesTableSkeleton = () => {
  return <TableSkeleton columns={policiesTableColumns} skeletonRows={5} />
}
