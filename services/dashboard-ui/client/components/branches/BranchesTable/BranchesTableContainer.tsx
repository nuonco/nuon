import { useParams } from 'react-router'
import type { TAppBranch } from '@/types'
import { BranchesTable, parseBranchesToTableData } from './BranchesTable'

interface IBranchesTableContainer {
  branches: TAppBranch[]
  isLoading: boolean
  pagination?: { hasNext: boolean; offset: number; limit: number }
}

export const BranchesTableContainer = ({
  branches,
  isLoading,
  pagination,
}: IBranchesTableContainer) => {
  const params = useParams()
  const orgId = params.orgId as string
  const appId = params.appId as string

  return (
    <BranchesTable
      data={parseBranchesToTableData(branches, orgId, appId)}
      isLoading={isLoading}
      pagination={pagination}
    />
  )
}
