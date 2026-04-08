import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { ComponentTypeFilterDropdown } from '@/components/components/ComponentTypeFilter'
import { ManagementDropdownContainer as ManagementDropdown } from '@/components/components/management/ManagementDropdown'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getComponents } from '@/lib'
import { ComponentsTable, parseComponentToTableData } from './ComponentsTable'

const LIMIT = 20

export const ComponentsTableContainer = ({
  pollInterval = 20000,
  shouldPoll = true,
}: {
  pollInterval?: number
  shouldPoll?: boolean
} = {}) => {
  const [searchParams] = useSearchParams()
  const { org } = useOrg()
  const { app } = useApp()
  const offset = Number(searchParams.get('offset') ?? 0)

  const { data: result, isLoading } = useQuery({
    queryKey: [
      'components',
      org.id,
      app.id,
      offset,
      searchParams.get('q'),
      searchParams.get('types'),
    ],
    queryFn: () =>
      getComponents({
        orgId: org.id,
        appId: app.id,
        offset,
        limit: LIMIT,
        q: searchParams.get('q') || undefined,
        types: searchParams.get('types') || undefined,
      }),
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
  })

  return (
    <ComponentsTable
      data={parseComponentToTableData(result?.data ?? [], org.id, app.id)}
      isLoading={isLoading}
      filterActions={
        <div className="flex items-center gap-3">
          <ComponentTypeFilterDropdown />
          <ManagementDropdown />
        </div>
      }
      pagination={{
        hasNext: result?.pagination?.hasNext ?? false,
        offset,
        limit: LIMIT,
      }}
    />
  )
}
